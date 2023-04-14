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

package erc20

import (
	"fmt"
	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

// Define key names for options
const (
	nameKey        = "name"
	symbolKey      = "symbol"
	decimalsKey    = "decimals"
	totalSupplyKey = "totalSupply"
)

// Define objectType names for prefix
const (
	allowancePrefix = "allowance"
	ApprovalPrefix  = "approval~account~spender"
	BalancePrefix   = "balance~account~id"
)

// Define key names for options

// ERC20 provides functions for transferring tokens between accounts
type ERC20 struct {
	contractapi.Contract

	nonce.INonce
}

func NewERC20(nonce nonce.INonce) *ERC20 {
	return &ERC20{
		INonce: nonce,
	}
}

var _ ISupply
var _ IERC20 = new(ERC20)

// TODO: Initializer

// TotalSupply returns the total token supply
func (erc20 *ERC20) TotalSupply(ctx context.ContextInterface) (uint64, error) {
	supply, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return 0, err
	}
	return library.BytesToUint64(supply)
}

// Name returns a descriptive name for fungible tokens in this contract.
func (erc20 *ERC20) Name(ctx context.ContextInterface) (string, error) {
	bytes, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Name: %s", err)
	}

	return string(bytes), nil
}

// Symbol returns an abbreviated name for fungible tokens in this contract.
func (erc20 *ERC20) Symbol(ctx context.ContextInterface) (string, error) {
	bytes, err := ctx.GetStub().GetState(symbolKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Symbol: %v", err)
	}

	return string(bytes), nil
}

// Decimal returns the decimal setting of fungible tokens in this contract.
// For example: if the decimal settings is 2, then transferring 525 tokens will be displayed to users as transferring '5.25 tokens'
func (erc20 *ERC20) Decimal(ctx context.ContextInterface) (uint8, error) {
	bytes, err := ctx.GetStub().GetState(decimalsKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get Decimal: %erc20", err)
	}

	dec, err := library.BytesToUint8(bytes)
	if err != nil {
		return 0, err
	}

	return dec, nil
}

// Mint creates new tokens and adds them to minter's account balance
// This function triggers a Transfer event
func (erc20 *ERC20) Mint(ctx context.ContextInterface, msg context.Message, to string, amount uint64) error {
	var err error
	toAddr := library.Address(to)

	if err = toAddr.Validate(); err != nil {
		return err
	}

	// Nonce Check & Increase
	if err = erc20.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return err
	}
	if _, err = erc20.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return err
	}

	// beforeTokenTransfer

	balanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{to})
	if err != nil {
		return err
	}
	preVal, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return err
	}
	preBalance, err := library.BytesToUint64(preVal)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(balanceKey, []byte(library.Uint64ToString(amount+preBalance)))
	if err != nil {
		return err
	}

	if err = ctx.EmitEvent("Transfer", &EventTransfer{
		Operator: ctx.MsgSender(),
		From:     library.ZeroAddress,
		To:       toAddr,
		Value:    amount,
	}); err != nil {
		return errors.Wrap(err, "Event TransferSingle")
	}

	// afterTokenTransfer

	return nil
}

// Burn redeems tokens the minter's account balance
// This function triggers a Transfer event
func (erc20 *ERC20) Burn(ctx context.ContextInterface, msg context.Message, amount uint64) error {
	var err error
	fromAddr := library.Address(ctx.MsgSender().String())

	if err = fromAddr.Validate(); err != nil {
		return err
	}

	// Nonce Check & Increase
	if err = erc20.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return err
	}
	if _, err = erc20.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return err
	}

	// beforeTokenTransfer

	balanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{ctx.MsgSender().String()})
	if err != nil {
		return err
	}
	preVal, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return err
	}
	preBalance, err := library.BytesToUint64(preVal)
	if err != nil {
		return err
	}

	if amount > preBalance {
		return fmt.Errorf("burning more than it remain")
	}

	err = ctx.GetStub().PutState(balanceKey, []byte(library.Uint64ToString(preBalance-amount)))
	if err != nil {
		return err
	}

	if err = ctx.EmitEvent("Transfer", &EventTransfer{
		Operator: ctx.MsgSender(),
		From:     fromAddr,
		To:       library.ZeroAddress,
		Value:    amount,
	}); err != nil {
		return errors.Wrap(err, "Event TransferSingle")
	}

	// afterTokenTransfer

	return nil
}

// Transfer transfers tokens from client account to recipient account.
// This function triggers a Transfer event.
func (erc20 *ERC20) Transfer(ctx context.ContextInterface, msg context.Message, to string, amount uint64) error {
	var err error

	// Nonce Check & Increase
	if err = erc20.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return err
	}
	if _, err = erc20.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return err
	}

	return _transfer(ctx, ctx.MsgSender().String(), to, amount)

}

func _transfer(ctx context.ContextInterface, from string, to string, amount uint64) error {
	var err error
	toAddr := library.Address(to)
	fromAddr := library.Address(from)

	if err = toAddr.Validate(); err != nil {
		return err
	}

	if err = fromAddr.Validate(); err != nil {
		return err
	}

	// beforeTokenTransfer

	senderBalanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{from})
	if err != nil {
		return err
	}
	receiverBalanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{to})
	if err != nil {
		return err
	}

	senderBalanceVal, err := ctx.GetStub().GetState(senderBalanceKey)
	if err != nil {
		return err
	}
	receiverBalanceVal, err := ctx.GetStub().GetState(receiverBalanceKey)
	if err != nil {
		return err
	}

	senderBalance, err := library.BytesToUint64(senderBalanceVal)
	if err != nil {
		return err
	}
	receiverBalance, err := library.BytesToUint64(receiverBalanceVal)
	if err != nil {
		return err
	}

	if senderBalance < amount {
		return fmt.Errorf("transferred more than it has")
	}

	err = ctx.GetStub().PutState(senderBalanceKey, []byte(library.Uint64ToString(amount+receiverBalance)))
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(receiverBalanceKey, []byte(library.Uint64ToString(senderBalance-amount)))
	if err != nil {
		return err
	}

	if err = ctx.EmitEvent("Transfer", &EventTransfer{
		Operator: ctx.MsgSender(),
		From:     fromAddr,
		To:       toAddr,
		Value:    amount,
	}); err != nil {
		return errors.Wrap(err, "Event Transfer")
	}

	// afterTokenTransfer

	return nil
}

// BalanceOf returns the balance of the given account
func (erc20 *ERC20) BalanceOf(ctx context.ContextInterface, account string) (uint64, error) {
	balanceKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{account})
	if err != nil {
		return 0, errors.Wrap(library.ErrInvalidCompositeKey, err.Error())
	}
	balance, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return 0, err
	}
	return library.BytesToUint64(balance)
}

// Approve allows the spender to withdraw from the calling client's token account
// The spender can withdraw multiple times if necessary, up to the value amount
// This function triggers an Approval event
func (erc20 *ERC20) Approve(ctx context.ContextInterface, msg context.Message, spender string, amountCap uint64) error {
	var err error
	spenderAddr := library.Address(spender)

	if err = spenderAddr.Validate(); err != nil {
		return err
	}

	// Nonce Check & Increase
	if err = erc20.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return err
	}
	if _, err = erc20.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return err
	}

	// Might have some check

	approvalKey, err := ctx.GetStub().CreateCompositeKey(ApprovalPrefix, []string{ctx.MsgSender().String(), spender, library.Uint64ToString(amountCap)})
	if err != nil {
		return err
	}
	if err = ctx.GetStub().PutState(approvalKey, []byte("Approved")); err != nil {
		return err
	}

	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{approvalKey})
	if err != nil {
		return err
	}
	if err = ctx.GetStub().PutState(allowanceKey, []byte(library.Uint64ToString(amountCap))); err != nil {
		return err
	}

	if err = ctx.EmitEvent("Approve", &EventApproval{
		Owner:     ctx.MsgSender(),
		Operator:  spenderAddr,
		Approved:  true,
		Allowance: amountCap,
	}); err != nil {
		return errors.Wrap(err, "Event Approve")
	}

	return nil
}

// IsApproved returns if the given owner account approves spender to withdraw from the owner
func (erc20 *ERC20) IsApproved(ctx context.ContextInterface, owner string, spender string) (bool, error) {
	approvalKey, err := ctx.GetStub().CreateCompositeKey(ApprovalPrefix, []string{owner, spender})
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

// Allowance returns the amount still available for the spender to withdraw from the owner
func (erc20 *ERC20) Allowance(ctx context.ContextInterface, owner string, spender string) (uint64, error) {
	isApproved, err := erc20.IsApproved(ctx, owner, spender)
	if err != nil {
		return 0, err
	}
	if !isApproved {
		return 0, fmt.Errorf("not approved")
	}

	approvalKey, err := ctx.GetStub().CreateCompositeKey(ApprovalPrefix, []string{owner, spender})
	if err != nil {
		return 0, err
	}

	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{approvalKey})
	if err != nil {
		return 0, err
	}
	allowance, err := ctx.GetStub().GetState(allowanceKey)

	return library.BytesToUint64(allowance)
}

// TransferFrom transfers the value amount from the "from" address to the "to" address
// This function triggers a Transfer event
func (erc20 *ERC20) TransferFrom(ctx context.ContextInterface, msg context.Message, from string, to string, amount uint64) error {

	var err error
	// Nonce Check & Increase
	if err = erc20.INonce.Check(ctx, ctx.MsgSender().String(), msg.Nonce); err != nil {
		return err
	}
	if _, err = erc20.INonce.Increment(ctx, ctx.MsgSender().String()); err != nil {
		return err
	}

	// check if approved
	isApproved, err := erc20.IsApproved(ctx, from, ctx.MsgSender().String())
	if err != nil {
		return err
	}
	if !isApproved {
		return fmt.Errorf("not approved")
	}

	// check if transfer amount is greater than allowed
	allowance, err := erc20.Allowance(ctx, from, ctx.MsgSender().String())
	if err != nil {
		return err
	}
	if amount > allowance {
		return fmt.Errorf("transfer amount greater than allowed amount")
	}

	// make a transfer
	if err = _transfer(ctx, from, to, amount); err != nil {
		return err
	}

	// reduce allowance by reclaiming approval
	approvalKey, err := ctx.GetStub().CreateCompositeKey(ApprovalPrefix, []string{from, to})
	if err != nil {
		return err
	}

	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{approvalKey})
	if err != nil {
		return err
	}

	if err = ctx.GetStub().PutState(allowanceKey, []byte(library.Uint64ToString(allowance-amount))); err != nil {
		return err
	}

	return nil
}
