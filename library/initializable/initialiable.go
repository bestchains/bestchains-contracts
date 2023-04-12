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

package initializable

import (
	"errors"

	"github.com/bestchains/bestchains-contracts/library/context"
)

type Status string

const (
	Initialized Status = "Initialized"
)

var (
	ErrAlreadyInitialized = errors.New("already initialized")
)

type Initializable struct{}

// Initializable checks whether this has been initialied agains that key
func (init *Initializable) TryInitialize(ctx context.ContextInterface, key string) error {
	val, err := ctx.GetStub().GetState(key)
	if err != nil {
		return err
	}
	if string(val) == string(Initialized) {
		return ErrAlreadyInitialized
	}
	return ctx.GetStub().PutState(key, []byte(Initialized))
}
