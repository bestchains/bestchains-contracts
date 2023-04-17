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
	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
)

// EventOwnershipTransferred emit when the owner changed
type EventOwnershipTransferred struct {
	PreviousOwner library.Address
	NewOwner      library.Address
}

// IOwnable defines the interfaces which ownable contract must implement
type IOwnable interface {
	Initialize(ctx context.ContextInterface) error
	Owner(ctx context.ContextInterface) (string, error)
	RenounceOwnership(ctx context.ContextInterface) error
	TransferOwnership(ctx context.ContextInterface, newOwner string) error
}

// EventRoleAdminChanged emit when role's admin role changed
type EventRoleAdminChanged struct {
	Role              []byte
	PreviousAdminRole []byte
	NewAdminRole      []byte
}

// EventRoleGranted emit when a account granted a role
type EventRoleGranted struct {
	Role    []byte
	Account library.Address
	Sender  library.Address
}

// EventRoleRevoked emit when a account's role got revoked
type EventRoleRevoked struct {
	Role    []byte
	Account library.Address
	Sender  library.Address
}

// IAccessControl defines the interfaces which access control contract must implement
type IAccessControl interface {
	IOwnable
	Initialize(ctx context.ContextInterface) error
	SetRoleAdmin(ctx context.ContextInterface, role []byte, adminRole []byte) error
	GetRoleAdmin(ctx context.ContextInterface, role []byte) ([]byte, error)
	HasRole(ctx context.ContextInterface, role []byte, account string) (bool, error)
	GrantRole(ctx context.ContextInterface, role []byte, account string) error
	RevokeRole(ctx context.ContextInterface, role []byte, account string) error
	RenounceRole(ctx context.ContextInterface, msg context.Message, role []byte, account string) error
}
