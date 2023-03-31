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
	"errors"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
)

var (
	ErrTokenAlreadyExist = errors.New("token already exists")
	ErrLengthMismatch    = errors.New("ids and amounts must have same length")
)

type ID uint64

func (id ID) String() string {
	return library.Uint64ToString(uint64(id))
}

func (id *ID) FromBytes(bytes []byte) error {
	ui64, err := library.BytesToUint64(bytes)
	if err != nil {
		return err
	}
	*id = ID(ui64)
	return nil
}

type EventTransferSingle struct {
	Operator library.Address `json:"operator"`
	From     library.Address `json:"from"`
	To       library.Address `json:"to"`
	ID       ID              `json:"id"`
	Value    uint64          `json:"value"`
}

type EventTransferBatch struct {
	Operator library.Address `json:"operator"`
	From     library.Address `json:"from"`
	To       library.Address `json:"to"`
	IDs      []ID            `json:"ids"`
	Values   []uint64        `json:"values"`
}

type EventApprovalForAll struct {
	Owner    string `json:"owner"`
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
}

type EventURI struct {
	Value string `json:"value"`
	ID    ID     `json:"id"`
}

type IERC1155 interface {
	SetURI(ctx context.ContextInterface, id ID, uri string) error
	URI(ctx context.ContextInterface, id ID) (string, error)

	BalanceOf(ctx context.ContextInterface, account string, id ID) (uint64, error)
	BalanceOfBatch(ctx context.ContextInterface, accounts []string, id []ID) ([]uint64, error)

	Mint(ctx context.ContextInterface, to string, id ID, amount uint64) error
	MintBatch(ctx context.ContextInterface, to string, ids []ID, amount []uint64) error

	SetApprovalForAll(ctx context.ContextInterface, msg context.Message, operator string, approved bool) error
	IsApprovedFroAll(ctx context.ContextInterface, account string, operator string) (bool, error)

	SafeTransferFrom(ctx context.ContextInterface, msg context.Message, from string, to string, id ID, amount uint64) error
	SafeBatchTransferFrom(ctx context.ContextInterface, msg context.Message, from string, to string, ids []uint64, amounts []uint64) error
}

type ISupply interface {
	TotalSupply(ctx context.ContextInterface, id ID) (uint64, error)
	Exists(ctx context.ContextInterface, id ID) (bool, error)
}

type Burnable interface {
	Burn(ctx context.ContextInterface, msg context.Message, id ID, amount uint64) error
}

type Pausable interface {
	Pause(ctx context.ContextInterface) error
	Unpause(ctx context.ContextInterface) error
}
