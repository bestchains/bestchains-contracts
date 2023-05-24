/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package depository

import (
	"encoding/hex"
	"strings"

	"github.com/bestchains/bestchains-contracts/contracts/access"
	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

const (
	IndexerKey       = "depository~index"
	DepositoryKey    = "depository~index~kid"
	DepositoryValKey = "depository~kid-val"

	EnableACLKey = "enable~acl"
)

var (
	RoleAdmin  = sha3.Sum256([]byte("role~admin"))
	RoleClient = sha3.Sum256([]byte("role~client"))
)

var _ IDepository = new(DepositoryContract)

// DepositoryContract implements IDepository interface
type DepositoryContract struct {
	contractapi.Contract

	nonce.INonce

	access.IAccessControl
}

// NewDepositoryContract creates a new DepositoryContract instance with the given nonce and access control contracts.
func NewDepositoryContract(nonceContract nonce.INonce, aclContract access.IAccessControl) *DepositoryContract {
	depositoryContract := new(DepositoryContract)

	// Set the name of the depository contract
	depositoryContract.Name = "org.bestchains.com.DepositoryContract"

	// Set the INonce and IAccessControl contracts
	depositoryContract.INonce = nonceContract
	depositoryContract.IAccessControl = aclContract

	// Set the TransactionContextHandler and BeforeTransaction
	depositoryContract.TransactionContextHandler = new(context.Context)
	depositoryContract.BeforeTransaction = context.BeforeTransaction

	return depositoryContract
}

// onlyRole checks if the caller has the specified role.
func (bc *DepositoryContract) onlyRole(ctx context.ContextInterface, role []byte) error {
	// Check if the caller has the specified role.
	result, err := bc.HasRole(ctx, role, ctx.MsgSender().String())
	if err != nil {
		return errors.Wrap(err, "onlyRole")
	}
	if !result {
		// Caller doesn't have the role.
		return errors.New("onlyRole: not authorized")
	}
	// Caller has the role.
	return nil
}

// Initialize initializes the DepositoryContract and returns an error if there is one.
func (bc *DepositoryContract) Initialize(ctx context.ContextInterface) error {
	// Call the parent's Initialize function.
	err := bc.IAccessControl.Initialize(ctx)
	if err != nil {
		// If there was an error, return it.
		return err
	}

	// If there was no error, return nil.
	return nil
}

// EnableACL enables the access control list
func (bc *DepositoryContract) EnableACL(ctx context.ContextInterface) error {
	err := ctx.GetStub().PutState(EnableACLKey, library.True.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// DisableACL disables the access control list
func (bc *DepositoryContract) DisableACL(ctx context.ContextInterface) error {
	err := ctx.GetStub().PutState(EnableACLKey, library.False.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// aclEnabled returns a boolean indicating whether the access control list is enabled
func (bc *DepositoryContract) aclEnabled(ctx context.ContextInterface) (bool, error) {
	val, err := ctx.GetStub().GetState(EnableACLKey)
	if err != nil {
		return false, err
	}

	return library.BytesToBool(val).Bool(), nil

}

// Total returns the total count
func (bc *DepositoryContract) Total(ctx context.ContextInterface) (uint64, error) {
	curr, err := currentCounter(ctx)
	if err != nil {
		return 0, err
	}
	return curr.Current(), nil
}

// BatchPutUntrustValue puts multiple untrusted values into the DepositoryContract.
// It receives a comma-separated string of values and returns a comma-separated string of corresponding KIDs (keys).
// If the batchVals string is empty, it returns an error.
func (bc *DepositoryContract) BatchPutUntrustValue(ctx context.ContextInterface, batchVals string) (string, error) {
	if batchVals == "" {
		return "", errors.New("empty batch value string")
	}
	vals := strings.Split(batchVals, ",")

	// Get the current counter value from the context.
	curr, err := currentCounter(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get counter")
	}

	// Initialize a slice to store the KIDs generated for each value.
	kids := make([]string, len(vals))

	// Initialize an EventBatchPutValue to record the batch operation.
	event := &EventBatchPutValue{
		Total: uint64(len(vals)),
		Items: make([]EventPutValue, len(vals)),
	}

	// Loop through each value and put it into the DepositoryContract.
	// Record the corresponding KID and other metadata in the event and kids slice.
	for i, val := range vals {
		index, kid, err := putValue(ctx, curr, val)
		if err != nil {
			return "", err
		}
		kids[i] = kid
		event.Items[i] = EventPutValue{
			Index:    index,
			KID:      kid,
			Operator: ctx.Operator().String(),
			Owner:    ctx.MsgSender().String(),
		}
	}

	// Increment the counter by the number of values added.
	err = incrementCounter(ctx, uint64(len(vals)))
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to increase counter")
	}

	// Emit the BatchPutUntrustValue event.
	err = ctx.EmitEvent("BatchPutUntrustValue", event)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to emit event BatchPutUntrustValue")
	}

	// Join the KIDs into a comma-separated string and return it.
	return strings.Join(kids, ","), nil
}

// PutUntrustValue adds an untrusted value to the DepositoryContract.
// It takes a context and a string value to add, and returns the resulting KID (key ID) and an error (if any).
func (bc *DepositoryContract) PutUntrustValue(ctx context.ContextInterface, val string) (string, error) {
	// Get the current counter value.
	curr, err := currentCounter(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get counter")
	}

	// Add the value to the depository and get its index and KID.
	index, kid, err := putValue(ctx, curr, val)
	if err != nil {
		return "", err
	}

	// Increment the counter.
	err = incrementCounter(ctx, 1)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to increase counter")
	}

	// Emit an event with information about the added value.
	err = ctx.EmitEvent("PutUntrustValue", &EventPutValue{
		Index:    index,
		KID:      kid,
		Operator: ctx.Operator().String(),
		Owner:    ctx.MsgSender().String(),
	})
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to emit EventPutUntrustValue")
	}

	// Return the KID of the added value.
	return kid, nil
}

// BatchPutValue puts multiple values into the DepositoryContract's state.
// It takes a batchVals string, which is a comma-separated list of values to be inserted.
// It returns a string representing the KIDs (Key IDs) of the inserted values and any error encountered.
func (bc *DepositoryContract) BatchPutValue(ctx context.ContextInterface, msg context.Message, batchVals string) (string, error) {
	// Check if access control is enabled.
	enabled, err := bc.aclEnabled(ctx)
	if err != nil {
		return "", errors.Wrap(err, "DepositoryContract: get aclEnabled")
	}
	if enabled {
		// If access control is enabled, check if the caller has the RoleClient role.
		if err = bc.onlyRole(ctx, RoleClient[:]); err != nil {
			return "", errors.Wrap(err, "onlyClient")
		}
	}

	// Check the nonce of the caller to prevent replay attacks.
	if err = bc.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return "", err
	}
	// Increment the nonce for the caller.
	if _, err = bc.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return "", err
	}

	// Make sure batchVals is not empty.
	if batchVals == "" {
		return "", errors.New("empty batch value string")
	}
	// Split batchVals into individual values.
	vals := strings.Split(batchVals, ",")

	// Get the current counter value.
	curr, err := currentCounter(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get counter")
	}

	// Initialize an array to store the KIDs of the inserted values.
	kids := make([]string, len(vals))

	// Initialize an event to store information about the inserted values.
	event := &EventBatchPutValue{
		Total: uint64(len(vals)),
		Items: make([]EventPutValue, len(vals)),
	}

	// Loop through each value in vals and insert it into the DepositoryContract's state.
	for i, val := range vals {
		// Insert the value into the state and get its index and KID.
		index, kid, err := putValue(ctx, curr, val)
		if err != nil {
			return "", err
		}
		kids[i] = kid

		// Add information about the inserted value to the event.
		event.Items[i] = EventPutValue{
			Index:    index,
			KID:      kid,
			Operator: ctx.Operator().String(),
			Owner:    ctx.MsgSender().String(),
		}
	}

	// Increment the counter by the number of inserted values.
	err = incrementCounter(ctx, uint64(len(vals)))
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to increase counter")
	}

	// Emit an event to indicate that the values have been inserted.
	err = ctx.EmitEvent("BatchPutUntrustValue", event)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to emit event BatchPutUntrustValue")
	}

	// Join the KIDs of the inserted values and return them.
	return strings.Join(kids, ","), nil
}

// PutValue adds a new value to the DepositoryContract and returns its KID.
// It checks ACL if enabled, increases the nonce, gets the current counter,
// puts the value into the ledger, increases the counter, and emits an event.
func (bc *DepositoryContract) PutValue(ctx context.ContextInterface, msg context.Message, val string) (string, error) {
	var err error

	// Check ACL if enabled
	enabled, err := bc.aclEnabled(ctx)
	if err != nil {
		return "", errors.Wrap(err, "DepositoryContract: get aclEnabled")
	}
	if enabled {
		if err = bc.onlyRole(ctx, RoleClient[:]); err != nil {
			return "", errors.Wrap(err, "onlyClient")
		}
	}

	// Increase nonce
	if err = bc.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return "", err
	}
	if _, err = bc.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return "", err
	}

	// Get current counter
	curr, err := currentCounter(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get counter")
	}

	// Put value into ledger
	index, kid, err := putValue(ctx, curr, val)
	if err != nil {
		return "", err
	}

	// Increase counter
	err = incrementCounter(ctx, 1)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to increase counter")
	}

	// Emit event PutValue
	err = ctx.EmitEvent("PutValue", &EventPutValue{
		Index:    index,
		KID:      kid,
		Operator: ctx.Operator().String(),
		Owner:    ctx.MsgSender().String(),
	})
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to emit EventPutValue")
	}

	return kid, nil
}

func putValue(ctx context.ContextInterface, curr *library.Counter, val string) (uint64, string, error) {
	// put value into database
	if val == "" {
		return 0, "", errors.New("empty input value")
	}

	kid := calculateKID(curr, []byte(val))

	// save key id
	depositoryKey, err := ctx.GetStub().CreateCompositeKey(DepositoryKey, []string{curr.String()})
	if err != nil {
		return 0, "", errors.Wrap(err, "Depository: invalid composite DepositoryKey")
	}
	err = ctx.GetStub().PutState(depositoryKey, []byte(kid))
	if err != nil {
		return 0, "", errors.Wrap(err, "Depository: failed to put DepositoryKey")
	}

	// save value
	depositoryValKey, err := ctx.GetStub().CreateCompositeKey(DepositoryValKey, []string{kid})
	if err != nil {
		return 0, "", errors.Wrap(err, "Depository: invalid composite DepositoryValKey")
	}
	err = ctx.GetStub().PutState(depositoryValKey, []byte(val))
	if err != nil {
		return 0, "", errors.Wrap(err, "Depository: failed to put DepositoryKey")
	}

	return curr.Current(), kid, nil
}

func calculateKID(counter *library.Counter, val []byte) string {
	hashedPubKey := sha3.Sum256(append(counter.Bytes(), val...))
	return hex.EncodeToString(hashedPubKey[12:])
}

func currentCounter(ctx context.ContextInterface) (*library.Counter, error) {
	val, err := ctx.GetStub().GetState(IndexerKey)
	if err != nil {
		return nil, errors.Wrap(err, "Depository: failed to read counter")
	}
	return library.BytesToCounter(val)
}

func incrementCounter(ctx context.ContextInterface, offset uint64) error {
	val, err := ctx.GetStub().GetState(IndexerKey)
	if err != nil {
		return errors.Wrap(err, "Depository: failed to read counter")
	}

	counter, err := library.BytesToCounter(val)
	if err != nil {
		return errors.Wrap(err, "Depository: invalid counter")
	}
	err = counter.Increment(offset)
	if err != nil {
		return errors.Wrap(err, "Depositoy: increment counter")
	}

	err = ctx.GetStub().PutState(IndexerKey, counter.Bytes())
	if err != nil {
		return errors.Wrap(err, "Depository: failed to update counter")
	}

	return nil
}

// GetValueByIndex retrieves the value at the index provided
// from the DepositoryContract.
// The index is expected to be a string representation of a byte slice
// that can be converted to a counter.
// If the value is found, it is returned as a string.
// If an error is encountered, it is returned with additional context.
func (bc *DepositoryContract) GetValueByIndex(ctx context.ContextInterface, index string) (string, error) {
	// Convert the index string to a counter
	counter, err := library.BytesToCounter([]byte(index))
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get val by counter or key id")
	}

	// Get the value associated with the counter
	val, err := getValByCounter(ctx, counter)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get val by counter or key id")
	}

	// Return the value as a string
	return string(val), nil
}

func getValByCounter(ctx context.ContextInterface, counter *library.Counter) ([]byte, error) {
	depositoryKey, err := ctx.GetStub().CreateCompositeKey(DepositoryKey, []string{counter.String()})
	if err != nil {
		return nil, errors.Wrap(err, "Depository: invalid composite DepositoryKey")
	}

	kid, err := ctx.GetStub().GetState(depositoryKey)
	if err != nil {
		return nil, errors.Wrap(err, "Depository: failed to get kid with index")
	}

	if kid == nil {
		return nil, errors.Errorf("Depository: kid with counter %s not found", counter.String())
	}

	return getValByKID(ctx, string(kid))
}

// GetValueByKID returns the value associated with the given key ID.
// It first retrieves the value from the context using the getValByKID helper function.
// If the value is found, it is returned as a string. Otherwise, an error is returned.
func (bc *DepositoryContract) GetValueByKID(ctx context.ContextInterface, kid string) (string, error) {
	val, err := getValByKID(ctx, kid)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func getValByKID(ctx context.ContextInterface, kid string) ([]byte, error) {
	depositoryValKey, err := ctx.GetStub().CreateCompositeKey(DepositoryValKey, []string{kid})
	if err != nil {
		return nil, errors.Wrap(err, "Depository: invalid composite DepositoryValKey")
	}

	val, err := ctx.GetStub().GetState(depositoryValKey)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, errors.Errorf("Depository: value with kid %s not found", kid)
	}

	return val, nil
}
