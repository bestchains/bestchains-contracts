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

package access

import (
	"github.com/pkg/errors"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/bestchains/bestchains-contracts/library/initializable"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	OwnableInitializedKey = "owner~initialized"
	OwnerKey              = "owner~account"
)

var _ IOwnable = new(OwnableContract)

// OwnableContract implements IOwnable which provides basic access control mechanism where there is an owner can be granted exclusive access to specific functions.
type OwnableContract struct {
	contractapi.Contract
	initializable *initializable.Initializable
}

func NewOwnableContract() *OwnableContract {
	ownable := new(OwnableContract)
	ownable.Name = "org.bestchains.com.OwnableContract"
	ownable.TransactionContextHandler = new(context.Context)
	ownable.initializable = &initializable.Initializable{}
	ownable.BeforeTransaction = context.BeforeTransaction

	return ownable
}

// Initialize will set the owner
func (ownable *OwnableContract) Initialize(ctx context.ContextInterface) error {
	var err error

	initOwner := ctx.Operator().String()

	if err := ownable.initializable.TryInitialize(ctx, OwnableInitializedKey); err != nil {
		return err
	}

	if err = transferOwnership(ctx, library.ZeroAddress, library.Address(initOwner)); err != nil {
		return err
	}

	return nil
}

// Owner returns the address of the current owner
// - ZeroAddress will be returned if current owner is empty
func (ownable *OwnableContract) Owner(ctx context.ContextInterface) (string, error) {
	currOwner, err := owner(ctx)
	if err != nil {
		return library.ZeroAddress.String(), err
	}
	return currOwner.String(), nil
}

func owner(ctx context.ContextInterface) (library.Address, error) {
	owner, err := ctx.GetStub().GetState(OwnerKey)
	if err != nil {
		return library.ZeroAddress, err
	}
	addr := library.Address(owner)
	if addr.EmptyAddress() {
		return library.ZeroAddress, nil
	}
	if err := addr.Validate(); err != nil {
		return library.ZeroAddress, err
	}
	return addr, nil
}

func onlyOwner(ctx context.ContextInterface) error {
	currOwner, err := owner(ctx)
	if err != nil {
		return err
	}

	if currOwner != ctx.Operator() {
		return errors.New("Ownable: caller is not the owner")
	}

	return nil
}

// RenounceOwnership will reset owner to ZeroAddress
// - only current Owner has this permission
func (ownable *OwnableContract) RenounceOwnership(ctx context.ContextInterface) error {
	var err error

	if err = onlyOwner(ctx); err != nil {
		return err
	}

	previousOwner, _ := owner(ctx)
	if err = transferOwnership(ctx, previousOwner, library.ZeroAddress); err != nil {
		return err
	}

	return nil
}

// TransferOwnership will transfer ownership to newOwner
// - only current Owner has this permission
func (ownable *OwnableContract) TransferOwnership(ctx context.ContextInterface, newOwner string) error {
	var err error

	if err = onlyOwner(ctx); err != nil {
		return err
	}

	previousOwner, _ := owner(ctx)
	if err = transferOwnership(ctx, previousOwner, library.Address(newOwner)); err != nil {
		return err
	}
	return nil
}

func transferOwnership(ctx context.ContextInterface, previousOwner library.Address, newOwner library.Address) error {
	var err error

	if err = newOwner.Validate(); err != nil {
		return err
	}
	if err = ctx.GetStub().PutState(OwnerKey, newOwner.Bytes()); err != nil {
		return err
	}

	ctx.EmitEvent("OwnershipTransferred", &EventOwnershipTransferred{
		PreviousOwner: previousOwner,
		NewOwner:      newOwner,
	})

	return nil
}
