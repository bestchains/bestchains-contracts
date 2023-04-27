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

package nonce

import (
	"errors"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var (
	ErrInvalidMessageNonce = errors.New("nonce mistmatch")
)

var _ INonce = new(Nonce)

type Nonce struct {
	contractapi.Contract
}

func NewNonceContract() INonce {
	nonceContract := new(Nonce)
	nonceContract.Name = "org.bestchains.com.NonceContract"
	nonceContract.TransactionContextHandler = new(context.Context)
	nonceContract.BeforeTransaction = context.BeforeTransaction
	return nonceContract
}

func (nonce *Nonce) Check(ctx context.ContextInterface, account string, dstNonce uint64) error {
	curr, err := nonce.Current(ctx, account)
	if err != nil {
		return err
	}
	if curr != dstNonce {
		return ErrInvalidMessageNonce
	}
	return nil
}

func (nonce *Nonce) Current(ctx context.ContextInterface, account string) (uint64, error) {
	nonceKey, err := ctx.GetStub().CreateCompositeKey(NoncePrefix, []string{account})
	if err != nil {
		return 0, err
	}
	val, err := ctx.GetStub().GetState(nonceKey)
	if err != nil {
		return 0, err
	}

	counter, err := library.BytesToCounter(val)
	if err != nil {
		return 0, err
	}

	return counter.Current(), nil
}

func (nonce *Nonce) Increment(ctx context.ContextInterface, account string) (uint64, error) {
	nonceKey, err := ctx.GetStub().CreateCompositeKey(NoncePrefix, []string{account})
	if err != nil {
		return 0, err
	}
	val, err := ctx.GetStub().GetState(nonceKey)
	if err != nil {
		return 0, err
	}
	counter, err := library.BytesToCounter(val)
	if err != nil {
		return 0, err
	}
	err = counter.Increment(1)
	if err != nil {
		return 0, err
	}
	err = ctx.GetStub().PutState(nonceKey, counter.Bytes())
	if err != nil {
		return 0, err
	}
	return counter.Current(), nil
}
