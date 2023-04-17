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

package erc1155

import (
	"strconv"

	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

const (
	NameKey   = "name"
	SymbolKey = "symbol"

	URIKey = "uri"

	SupplyPrefix   = "supply~id"
	BalancePrefix  = "balance~account~id"
	ApprovalPrefix = "approval~account~operator"
)

var _ ISupply = new(ERC1155)
var _ IERC1155 = new(ERC1155)

type ERC1155 struct {
	contractapi.Contract

	nonce.INonce
}

func NewERC1155(nonce nonce.INonce) *ERC1155 {
	erc1155Contract := new(ERC1155)

	erc1155Contract.Name = "org.bestchains.com.ERC1155Contract"
	erc1155Contract.TransactionContextHandler = new(context.Context)
	erc1155Contract.BeforeTransaction = context.BeforeTransaction

	erc1155Contract.INonce = nonce

	return erc1155Contract
}

func (erc1155 *ERC1155) Initialize(ctx context.ContextInterface, name string, symbol string) error {
	return nil
}

/* ISupply */

func (erc1155 *ERC1155) TotalSupply(ctx context.ContextInterface, id ID) (uint64, error) {
	idString := strconv.FormatUint(uint64(id), 10)
	supplyKey, err := ctx.GetStub().CreateCompositeKey(SupplyPrefix, []string{idString})
	if err != nil {
		return 0, errors.Wrap(library.ErrInvalidCompositeKey, err.Error())
	}
	supply, err := ctx.GetStub().GetState(supplyKey)
	if err != nil {
		return 0, err
	}
	return library.BytesToUint64(supply)
}

func (erc1155 *ERC1155) Exists(ctx context.ContextInterface, id ID) (bool, error) {
	total, err := erc1155.TotalSupply(ctx, id)
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

/* IERC1155 */

func (erc1155 *ERC1155) SetURI(ctx context.ContextInterface, id ID, uri string) error {
	// TODO: permission check

	idString := strconv.FormatUint(uint64(id), 10)
	uriKey, err := ctx.GetStub().CreateCompositeKey(URIKey, []string{idString})
	if err != nil {
		return errors.Wrap(library.ErrInvalidCompositeKey, err.Error())
	}
	_, err = ctx.GetStub().GetState(uriKey)
	if err != nil {
		return err
	}

	return nil
}

func (erc1155 *ERC1155) URI(ctx context.ContextInterface, id ID) (string, error) {
	uriKey, err := ctx.GetStub().CreateCompositeKey(URIKey, []string{id.String()})
	if err != nil {
		return "", errors.Wrap(err, "")
	}

	uri, err := ctx.GetStub().GetState(uriKey)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	// TODO: permission check
	return string(uri), nil
}

func (erc1155 *ERC1155) BalanceOf(ctx context.ContextInterface, account string, id ID) (uint64, error) {
	balanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{account, id.String()})
	if err != nil {
		return 0, errors.Wrap(library.ErrInvalidCompositeKey, err.Error())
	}
	balance, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return 0, err
	}
	return library.BytesToUint64(balance)
}

func (erc1155 *ERC1155) BalanceOfBatch(ctx context.ContextInterface, accounts []string, ids []ID) ([]uint64, error) {
	if len(accounts) != len(ids) {
		return nil, errors.New("accounts and ids must have the same length")
	}

	balances := make([]uint64, len(accounts))
	for index, account := range accounts {
		balance, err := erc1155.BalanceOf(ctx, account, ids[index])
		if err != nil {
			return nil, errors.Wrapf(err, "balanceOf: %s %d", account, ids[index])
		}
		balances[index] = balance
	}

	return balances, nil
}

// Mint by operator
func (erc1155 *ERC1155) Mint(ctx context.ContextInterface, to string, id ID, amount uint64) error {
	var err error

	toAddr := library.Address(to)

	if err = erc1155.beforeTokenTransfer(ctx, library.ZeroAddress, toAddr, []ID{id}, []uint64{amount}); err != nil {
		return err
	}

	balanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{to, id.String()})
	if err != nil {
		return err
	}
	val, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return err
	}
	preBalance, err := library.BytesToUint64(val)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(balanceKey, []byte(library.Uint64ToString(amount+preBalance)))
	if err != nil {
		return err
	}

	if err = ctx.EmitEvent("TransferSingle", &EventTransferSingle{
		Operator: ctx.MsgSender(),
		From:     library.ZeroAddress,
		To:       toAddr,
		ID:       id,
		Value:    amount,
	}); err != nil {
		return errors.Wrap(err, "Event TransferSingle")
	}

	if err = erc1155.afterTokenTransfer(ctx, library.ZeroAddress, toAddr, []ID{id}, []uint64{amount}); err != nil {
		return err
	}

	return nil
}

func (erc1155 *ERC1155) beforeTokenTransfer(ctx context.ContextInterface, from library.Address, to library.Address, ids []ID, amounts []uint64) error {
	if len(ids) != len(amounts) {
		return ErrLengthMismatch
	}
	return nil
}

func (erc1155 *ERC1155) afterTokenTransfer(ctx context.ContextInterface, from library.Address, to library.Address, ids []ID, amounts []uint64) error {
	return nil
}

func (erc1155 *ERC1155) MintBatch(ctx context.ContextInterface, to string, ids []ID, amounts []uint64) error {
	var err error

	toAddr := library.Address(to)
	if err = erc1155.beforeTokenTransfer(ctx, library.ZeroAddress, toAddr, ids, amounts); err != nil {
		return err
	}

	// batch mint logic
	for index, id := range ids {
		balanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{to, id.String()})
		if err != nil {
			return err
		}

		val, err := ctx.GetStub().GetState(balanceKey)
		if err != nil {
			return err
		}
		preBalance, err := library.BytesToUint64(val)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(balanceKey, []byte(library.Uint64ToString(amounts[index]+preBalance)))
		if err != nil {
			return err
		}
	}

	if err = ctx.EmitEvent("TransferBatch", &EventTransferBatch{
		Operator: ctx.MsgSender(),
		From:     library.ZeroAddress,
		To:       toAddr,
		IDs:      ids,
		Values:   amounts,
	}); err != nil {
		return errors.Wrap(err, "Event TransferBatch")
	}

	if err = erc1155.afterTokenTransfer(ctx, library.ZeroAddress, toAddr, ids, amounts); err != nil {
		return err
	}

	return nil
}

func (erc1155 *ERC1155) SetApprovalForAll(ctx context.ContextInterface, msg context.Message, operator string, approved bool) error {
	var err error

	// Nonce Check & Increase
	if err = erc1155.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return err
	}
	if _, err = erc1155.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return err
	}

	// Approve
	approvalKey, err := ctx.GetStub().CreateCompositeKey(ApprovalPrefix, []string{ctx.MsgSender().String(), operator})
	if err != nil {
		return err
	}
	if err = ctx.GetStub().PutState(approvalKey, []byte("Approved")); err != nil {
		return err
	}

	return nil
}

func (erc1155 *ERC1155) IsApprovedFroAll(ctx context.ContextInterface, account string, operator string) (bool, error) {
	approvalKey, err := ctx.GetStub().CreateCompositeKey(ApprovalPrefix, []string{account, operator})
	if err != nil {
		return false, err
	}
	approved, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return false, err
	}

	if string(approved) != "Approved" {
		return false, nil
	}

	return true, nil
}

func (erc1155 *ERC1155) SafeTransferFrom(ctx context.ContextInterface, msg context.Message, from string, to string, id ID, amount uint64) error {
	// TODO: permission check

	return nil
}

func (erc1155 *ERC1155) SafeBatchTransferFrom(ctx context.ContextInterface, msg context.Message, from string, to string, ids []uint64, amounts []uint64) error {
	// TODO: permission check

	return nil
}
