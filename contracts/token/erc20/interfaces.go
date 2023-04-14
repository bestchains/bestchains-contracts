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
	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
)

type EventTransfer struct {
	Operator library.Address `json:"operator"`
	From     library.Address `json:"from"`
	To       library.Address `json:"to"`
	Value    uint64          `json:"value"`
}

type EventApproval struct {
	Owner     library.Address `json:"owner"`
	Operator  library.Address `json:"operator"`
	Approved  bool            `json:"approved"`
	Allowance uint64          `json:"allowance"`
}

type IERC20 interface {
	Name(ctx context.ContextInterface) (string, error)
	Symbol(ctx context.ContextInterface) (string, error)
	Decimal(ctx context.ContextInterface) (uint8, error) // Same as ETH

	Mint(ctx context.ContextInterface, msg context.Message, to string, amount uint64) error
	Burn(ctx context.ContextInterface, msg context.Message, amount uint64) error

	BalanceOf(ctx context.ContextInterface, account string) (uint64, error)

	Approve(ctx context.ContextInterface, msg context.Message, spender string, amountCap uint64) error
	IsApproved(ctx context.ContextInterface, owner string, spender string) (bool, error)

	Allowance(ctx context.ContextInterface, ownerAccount string, spenderAccount string) (uint64, error)

	Transfer(ctx context.ContextInterface, msg context.Message, to string, amount uint64) error
	TransferFrom(ctx context.ContextInterface, msg context.Message, from string, to string, amount uint64) error
}

type ISupply interface {
	TotalSupply(ctx context.ContextInterface) (uint64, error)
}
