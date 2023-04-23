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

/*
NewDepositoryContract creates a new instance of a DepositoryContract object
with the given nonceContract and aclContract parameters.

@param nonceContract: An instance implementing the INonce interface.
@param aclContract: An instance implementing the IAccessControl interface.

@return: A pointer to the newly created DepositoryContract object.
*/
func NewDepositoryContract(nonceContract nonce.INonce, aclContract access.IAccessControl) *DepositoryContract {
	DepositoryContract := new(DepositoryContract)
	DepositoryContract.Name = "org.bestchains.com.DepositoryContract"

	DepositoryContract.INonce = nonceContract
	DepositoryContract.IAccessControl = aclContract

	DepositoryContract.TransactionContextHandler = new(context.Context)
	DepositoryContract.BeforeTransaction = context.BeforeTransaction

	return DepositoryContract
}

/*
onlyRole checks if the caller has a specific role in the Depository contract.
- ctx: the context interface for the blockchain state.
- role: the role to check for.
Returns an error if the caller does not have the role or if there is an issue with the blockchain.
*/
func (bc *DepositoryContract) onlyRole(ctx context.ContextInterface, role []byte) error {
	result, err := bc.HasRole(ctx, role, ctx.MsgSender().String())
	if err != nil {
		return errors.Wrap(err, "onlyRole")
	}
	if !result {
		return errors.New("onlyRole: not authorized")
	}
	return nil
}

/*
Initialize initializes the DepositoryContract instance by calling the `Initialize` method of its parent `IAccessControl` instance. It returns an error if there was a problem during initialization.

Parameters:
- ctx: A context.ContextInterface instance representing the current execution context.

Returns:
- error: An error if there was a problem during initialization, or nil if initialization succeeded.
*/
func (bc *DepositoryContract) Initialize(ctx context.ContextInterface) error {
	err := bc.IAccessControl.Initialize(ctx)
	if err != nil {
		return err
	}

	return nil
}

/*
EnableACL enables access control list (ACL) on this DepositoryContract instance.

@param ctx context.ContextInterface: the context interface
@return error: returns an error if there was one during the state update
*/
func (bc *DepositoryContract) EnableACL(ctx context.ContextInterface) error {
	err := ctx.GetStub().PutState(EnableACLKey, library.True.Bytes())
	if err != nil {
		return err
	}
	return nil
}

/*
DisableACL disables the Access Control List (ACL) of the DepositoryContract.
It takes a context interface as the only argument.

Arguments:
- ctx (context.ContextInterface): The context interface to interact with the blockchain.

Returns:
- error: Returns an error if there was any issue while putting the EnableACLKey's value to false in the blockchain's state database. Returns nil otherwise.
*/
func (bc *DepositoryContract) DisableACL(ctx context.ContextInterface) error {
	err := ctx.GetStub().PutState(EnableACLKey, library.False.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (bc *DepositoryContract) aclEnabled(ctx context.ContextInterface) (bool, error) {
	val, err := ctx.GetStub().GetState(EnableACLKey)
	if err != nil {
		return false, err
	}

	return library.BytesToBool(val).Bool(), nil

}

/*
Total returns the total number of deposits made to the DepositoryContract.

Parameters:
- ctx (context.ContextInterface): The context interface for the function.

Return:
- uint64: The total number of deposits made to the DepositoryContract.
- error: An error, if the operation failed.
*/
func (bc *DepositoryContract) Total(ctx context.ContextInterface) (uint64, error) {
	curr, err := currentCounter(ctx)
	if err != nil {
		return 0, err
	}
	return curr.Current(), nil
}

/*
GetValueByKID retrieves the value associated with the given KID.

Args:
- ctx (context.ContextInterface): The context interface.
- kid (string): The KID to retrieve the value for.

Returns:
- (string): The value associated with the given KID.
- (error): An error if the value could not be retrieved.
*/
func (bc *DepositoryContract) PutUntrustValue(ctx context.ContextInterface, val string) (string, error) {
	// put value into ledger
	index, kid, err := putValue(ctx, val)
	if err != nil {
		return "", err
	}

	// emit event PutUntrustValue
	err = ctx.EmitEvent("PutUntrustValue", &EventPutUntrustValue{
		Index:    index,
		KID:      kid,
		Operator: ctx.Operator().String(),
		Owner:    ctx.MsgSender().String(),
	})
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to emit EventPutUntrustValue")
	}

	return kid, nil
}

/*
PutValue puts a new value into the depository. It takes a context interface, a message, and a string value as input.
It returns the key ID of the value and an error.

Params:
- ctx (context.ContextInterface): The context interface for the transaction.
- msg (context.Message): The message containing the transaction details.
- val (string): The value to be stored in the depository.

Returns:
- string: The key ID of the stored value.
- error: An error if any occurred during the transaction.
*/
func (bc *DepositoryContract) PutValue(ctx context.ContextInterface, msg context.Message, val string) (string, error) {
	var err error

	// check acl if enabled
	enabled, err := bc.aclEnabled(ctx)
	if err != nil {
		return "", errors.Wrap(err, "DepositoryContract: get aclEnabled")
	}
	if enabled {
		if err = bc.onlyRole(ctx, RoleClient[:]); err != nil {
			return "", errors.Wrap(err, "onlyClient")
		}
	}

	// increase nonce
	if err = bc.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return "", err
	}
	if _, err = bc.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return "", err
	}

	// put value into ledger
	index, kid, err := putValue(ctx, val)
	if err != nil {
		return "", err
	}

	// emit event PutValue
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

func putValue(ctx context.ContextInterface, val string) (uint64, string, error) {

	// put value into database
	if val == "" {
		return 0, "", errors.New("empty input value")
	}
	curr, err := currentCounter(ctx)
	if err != nil {
		return 0, "", errors.Wrap(err, "Depository: failed to get counter")
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

	// increase counter
	err = incrementCounter(ctx)
	if err != nil {
		return 0, "", errors.Wrap(err, "Depository: failed to increase counter")
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

func incrementCounter(ctx context.ContextInterface) error {
	val, err := ctx.GetStub().GetState(IndexerKey)
	if err != nil {
		return errors.Wrap(err, "Depository: failed to read counter")
	}

	counter, err := library.BytesToCounter(val)
	if err != nil {
		return errors.Wrap(err, "Depository: invalid counter")
	}
	counter.Increment()
	err = ctx.GetStub().PutState(IndexerKey, counter.Bytes())
	if err != nil {
		return errors.Wrap(err, "Depository: failed to update counter")
	}

	return nil
}

/*
GetValueByIndex retrieves a value from the DepositoryContract using an index.

@param ctx: A context interface.
@param index: A string representing the index to retrieve the value from.

@returns A string representing the retrieved value.
@returns An error if there was an issue retrieving the value.
*/
func (bc *DepositoryContract) GetValueByIndex(ctx context.ContextInterface, index string) (string, error) {
	// try counter
	counter, err := library.BytesToCounter([]byte(index))
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get val by counter or key id")
	}
	val, err := getValByCounter(ctx, counter)
	if err != nil {
		return "", errors.Wrap(err, "Depository: failed to get val by counter or key id")
	}

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

/*
GetValueByKID retrieves the value associated with the given KID.

Args:
- ctx (context.ContextInterface): The context interface.
- kid (string): The KID to retrieve the value for.

Returns:
- (string): The value associated with the given KID.
- (error): An error if the value could not be retrieved.
*/
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
