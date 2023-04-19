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

package basic

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
	IndexerKey  = "basic~index"
	BasicKey    = "basic~index~kid"
	BasicValKey = "basic~kid-val"
)

var (
	RoleAdmin  = sha3.Sum256([]byte("role~admin"))
	RoleClient = sha3.Sum256([]byte("role~client"))
)

var _ IBasic = new(BasicContract)

// BasicContract provides simple key-value Get/Put
type BasicContract struct {
	contractapi.Contract

	nonce.INonce
	// Ownable
	access.IAccessControl
}

func NewBasicContract(nonceContract nonce.INonce, aclContract access.IAccessControl) *BasicContract {
	basicContract := new(BasicContract)
	basicContract.Name = "org.bestchains.com.BasicContract"

	basicContract.INonce = nonceContract
	basicContract.IAccessControl = aclContract

	basicContract.TransactionContextHandler = new(context.Context)
	basicContract.BeforeTransaction = context.BeforeTransaction

	return basicContract
}

func (bc *BasicContract) onlyRole(ctx context.ContextInterface, role []byte) error {
	result, err := bc.HasRole(ctx, role, ctx.MsgSender().String())
	if err != nil {
		return errors.Wrap(err, "onlyRole")
	}
	if !result {
		return errors.New("onlyRole: not authorized")
	}
	return nil
}

// Initialize the contract by setting transaction operator as owner
func (bc *BasicContract) Initialize(ctx context.ContextInterface) error {
	err := bc.IAccessControl.Initialize(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Total key-value paris stored
func (bc *BasicContract) Total(ctx context.ContextInterface) (uint64, error) {
	curr, err := currentCounter(ctx)
	if err != nil {
		return 0, err
	}
	return curr.Current(), nil
}

// PutValue stores kval with pre-defined key calculation which returns
func (bc *BasicContract) PutValue(ctx context.ContextInterface, msg context.Message, val string) (string, error) {
	var err error

	// check msg sender has this permission
	if err = bc.onlyRole(ctx, RoleClient[:]); err != nil {
		return "", errors.Wrap(err, "onlyClient")
	}

	// TODO: Gas calculation

	// increase nonce
	if err = bc.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return "", err
	}
	if _, err = bc.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return "", err
	}

	if val == "" {
		return "", errors.New("empty input value")
	}
	curr, err := currentCounter(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to get counter")
	}

	kid := calculateKID(curr, []byte(val))

	// save key id
	basicKey, err := ctx.GetStub().CreateCompositeKey(BasicKey, []string{curr.String()})
	if err != nil {
		return "", errors.Wrap(err, "Basic: invalid composite BasicKey")
	}
	err = ctx.GetStub().PutState(basicKey, []byte(kid))
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to put BasicKey")
	}

	// save value
	basicValKey, err := ctx.GetStub().CreateCompositeKey(BasicValKey, []string{kid})
	if err != nil {
		return "", errors.Wrap(err, "Basic: invalid composite BasicValKey")
	}
	err = ctx.GetStub().PutState(basicValKey, []byte(val))
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to put BasicKey")
	}

	// increase counter
	err = incrementCounter(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to increase counter")
	}

	err = ctx.EmitEvent("PutValue", &EventPutValue{
		Index: curr.Current(),
		KID:   kid,
	})
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to emit EventPutValue")
	}

	return kid, nil
}

// calculateKID with `sha3(counter,kval)` which returns hex encoded string
func calculateKID(counter *library.Counter, val []byte) string {
	hashedPubKey := sha3.Sum256(append(counter.Bytes(), val...))
	return hex.EncodeToString(hashedPubKey[12:])
}

func currentCounter(ctx context.ContextInterface) (*library.Counter, error) {
	val, err := ctx.GetStub().GetState(IndexerKey)
	if err != nil {
		return nil, errors.Wrap(err, "Basic: failed to read counter")
	}
	return library.BytesToCounter(val)
}

func incrementCounter(ctx context.ContextInterface) error {
	val, err := ctx.GetStub().GetState(IndexerKey)
	if err != nil {
		return errors.Wrap(err, "Basic: failed to read counter")
	}

	counter, err := library.BytesToCounter(val)
	if err != nil {
		return errors.Wrap(err, "Basic: invalid counter")
	}
	counter.Increment()
	err = ctx.GetStub().PutState(IndexerKey, counter.Bytes())
	if err != nil {
		return errors.Wrap(err, "Basic: failed to update counter")
	}

	return nil
}

// GetValue get kval with counter index or key id
func (bc *BasicContract) GetValueByIndex(ctx context.ContextInterface, index string) (string, error) {
	// try counter
	counter, err := library.BytesToCounter([]byte(index))
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to get val by counter or key id")
	}
	val, err := getValByCounter(ctx, counter)
	if err != nil {
		return "", errors.Wrap(err, "Basic: failed to get val by counter or key id")
	}

	return string(val), nil
}

func getValByCounter(ctx context.ContextInterface, counter *library.Counter) ([]byte, error) {
	basicKey, err := ctx.GetStub().CreateCompositeKey(BasicKey, []string{counter.String()})
	if err != nil {
		return nil, errors.Wrap(err, "Basic: invalid composite BasicKey")
	}

	kid, err := ctx.GetStub().GetState(basicKey)
	if err != nil {
		return nil, errors.Wrap(err, "Basic: failed to get kid with index")
	}

	if kid == nil {
		return nil, errors.Errorf("Basic: kid with counter %s not found", counter.String())
	}

	return getValByKID(ctx, string(kid))
}

// GetValue get kval with counter index or key id
func (bc *BasicContract) GetValueByKID(ctx context.ContextInterface, kid string) (string, error) {
	val, err := getValByKID(ctx, kid)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func getValByKID(ctx context.ContextInterface, kid string) ([]byte, error) {
	basicValKey, err := ctx.GetStub().CreateCompositeKey(BasicValKey, []string{kid})
	if err != nil {
		return nil, errors.Wrap(err, "Basic: invalid composite BasicValKey")
	}

	val, err := ctx.GetStub().GetState(basicValKey)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, errors.Errorf("Basic: value with kid %s not found", kid)
	}

	return val, nil
}
