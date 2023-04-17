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

package context

import (
	"encoding/json"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var (
	ErrEmptyEventName      = errors.New("empty event name")
	ErrNilEventPayload     = errors.New("nil event payload")
	ErrInvalidEventPayload = errors.New("invalid event payload")
)

type ContextInterface interface {
	contractapi.TransactionContextInterface

	Operator() library.Address

	SetMsgSender(library.Address)
	MsgSender() library.Address

	EmitEvent(event string, payload interface{}) error
}

type Context struct {
	contractapi.TransactionContext

	// operator who call this tx
	operator library.Address

	// msgSender who is responsible the payload
	msgSender library.Address
}

func BeforeTransaction(ctx ContextInterface) error {
	var err error

	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) > 1 {
		msg := new(Message)

		// DO NOT HAVE
		if err = msg.Unmarshal([]byte(args[0])); err != nil {
			return nil
		}

		// Validate Args
		msgSender, err := msg.VerifyAgainstArgs(args[1:]...)
		if err != nil {
			return err
		}

		// set msg sender
		ctx.SetMsgSender(msgSender)
	}
	return nil
}

func (ctx *Context) Operator() library.Address {
	if ctx.operator == library.ZeroAddress || ctx.operator.String() == "" {
		crt, err := ctx.GetClientIdentity().GetX509Certificate()
		if err != nil {
			return library.ZeroAddress
		}

		var operator = new(library.Address)
		err = operator.FromPublicKey(crt.PublicKey)
		if err != nil {
			return library.ZeroAddress
		}
		ctx.operator = *operator
	}

	return ctx.operator
}

func (ctx *Context) SetMsgSender(msgSender library.Address) {
	ctx.msgSender = msgSender
}

func (ctx *Context) MsgSender() library.Address {
	return ctx.msgSender
}

func (ctx *Context) EmitEvent(event string, payload interface{}) error {
	if event == "" {
		return ErrEmptyEventName
	}
	if payload == nil {
		return ErrNilEventPayload
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(ErrInvalidEventPayload, err.Error())
	}

	return ctx.GetStub().SetEvent(event, bytes)
}
