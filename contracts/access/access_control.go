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
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	RoleAdminPrefix = "role~admin"
	RolePrefix      = "role~account"

	True  = "true"
	False = "false"
)

var _ IAccessControl = new(AccessControlContract)

// AccessControlContract implements IAccessControl
type AccessControlContract struct {
	contractapi.Contract
	IOwnable
}

func NewAccessControlContract(ownable IOwnable) *AccessControlContract {
	accessControl := new(AccessControlContract)

	accessControl.IOwnable = ownable

	accessControl.Name = "org.bestchains.com.AccessControlContract"
	accessControl.TransactionContextHandler = new(context.Context)
	accessControl.BeforeTransaction = context.BeforeTransaction

	return accessControl
}

func (accessControl *AccessControlContract) Initialize(ctx context.ContextInterface, initOwner string) error {
	var err error

	if err = accessControl.IOwnable.Initialize(ctx, initOwner); err != nil {
		return errors.Wrap(err, "AccessControl: initialize")
	}

	return nil
}

// SetRoleAdmin only when operator is the owner
// - emit event `RoleAdminChanged`
func (accessControl *AccessControlContract) SetRoleAdmin(ctx context.ContextInterface, role []byte, adminRole []byte) error {
	var err error

	if string(role) == "" || string(adminRole) == "" {
		return errors.New("AccessControl: role and adminRole must not be empty")
	}

	if err = onlyOwner(ctx); err != nil {
		return errors.Wrap(err, "AccessControl: onlyOwner")
	}

	roleAdminKey, err := ctx.GetStub().CreateCompositeKey(RoleAdminPrefix, []string{string(role)})
	if err != nil {
		return errors.Wrap(err, "AccessControl: create role's composite key")
	}

	previousAdminRole, err := ctx.GetStub().GetState(roleAdminKey)
	if err != nil {
		return errors.Wrap(err, "AccessControl: create role's composite key")
	}

	if err = ctx.GetStub().PutState(roleAdminKey, adminRole); err != nil {
		return errors.Wrap(err, "AccessControl: create role's composite key")
	}

	if err = ctx.EmitEvent("RoleAdminChanged", &EventRoleAdminChanged{
		Role:              role,
		PreviousAdminRole: previousAdminRole,
		NewAdminRole:      adminRole,
	}); err != nil {
		return errors.Wrap(err, "AccessControl: create role's composite key")
	}

	return nil
}

// GetRoleAdmin returns role's admin role
func (accessControl *AccessControlContract) GetRoleAdmin(ctx context.ContextInterface, role []byte) ([]byte, error) {
	return getRoleAdmin(ctx, role)
}

func getRoleAdmin(ctx context.ContextInterface, role []byte) ([]byte, error) {
	roleAdminKey, err := ctx.GetStub().CreateCompositeKey(RoleAdminPrefix, []string{string(role)})
	if err != nil {
		return nil, err
	}

	val, err := ctx.GetStub().GetState(roleAdminKey)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func onlyRoleAdmin(ctx context.ContextInterface, role []byte, account library.Address) error {
	roleAdmin, err := getRoleAdmin(ctx, role)
	if err != nil {
		return err
	}

	if err = hasRole(ctx, roleAdmin, account); err != nil {
		return err
	}

	return nil
}

// HasRole returns if account has been granted `role`
func (accessControl *AccessControlContract) HasRole(ctx context.ContextInterface, role []byte, account string) (bool, error) {
	if err := hasRole(ctx, role, library.Address(account)); err != nil {
		return false, err
	}
	return true, nil
}

func hasRole(ctx context.ContextInterface, role []byte, account library.Address) error {
	roleKey, err := ctx.GetStub().CreateCompositeKey(RolePrefix, []string{string(role), account.String()})
	if err != nil {
		return err
	}

	val, err := ctx.GetStub().GetState(roleKey)
	if err != nil {
		return err
	}

	if string(val) != True {
		return errors.Errorf("AccessingControl: account %s is missing role %s", account, library.BytesToHexString(role))
	}

	return nil
}

// GrantRole grants `role` to `account` only when operator has `role`'s admin role
// - emit event `RoleGranted` if succ
func (accessControl *AccessControlContract) GrantRole(ctx context.ContextInterface, role []byte, account string) error {
	var err error

	if err = library.Address(account).Validate(); err != nil {
		return errors.Wrap(err, "AccessControl: invalid account")
	}

	if err = onlyRoleAdmin(ctx, role, ctx.Operator()); err != nil {
		return errors.Wrap(err, "AccessControl: onlyRoleAdmin")
	}

	roleKey, err := ctx.GetStub().CreateCompositeKey(RolePrefix, []string{string(role), account})
	if err != nil {
		return err
	}

	if err = ctx.GetStub().PutState(roleKey, []byte(True)); err != nil {
		return err
	}

	if err = ctx.EmitEvent("RoleGranted", &EventRoleGranted{
		Role:    role,
		Account: library.Address(account),
		Sender:  ctx.Operator(),
	}); err != nil {
		return errors.Wrap(err, "AccessControl: event")
	}

	return nil
}

// Revoke grants `role` to `account` only when operator has `role`'s admin role
// - emit event `RoleRevoked` if succ
func (accessControl *AccessControlContract) RevokeRole(ctx context.ContextInterface, role []byte, account string) error {
	var err error

	if err = onlyRoleAdmin(ctx, role, ctx.Operator()); err != nil {
		return errors.Wrap(err, "AccessControl: onlyRoleAdmin")
	}

	if err = revokeRole(ctx, role, library.Address(account)); err != nil {
		return errors.Wrap(err, "AccessControl: revokeRole")
	}

	return nil
}

func revokeRole(ctx context.ContextInterface, role []byte, account library.Address) error {
	roleKey, err := ctx.GetStub().CreateCompositeKey(RolePrefix, []string{string(role), account.String()})
	if err != nil {
		return err
	}

	if err = ctx.GetStub().DelState(roleKey); err != nil {
		return err
	}

	if err = ctx.EmitEvent("RoleRevoked", &EventRoleRevoked{
		Role:    role,
		Account: account,
		Sender:  ctx.Operator(),
	}); err != nil {
		return errors.Wrap(err, "AccessControl: event")
	}

	return nil
}

// RenounceRole by account itself
func (accessControl *AccessControlContract) RenounceRole(ctx context.ContextInterface, msg context.Message, role []byte, account string) error {
	var err error

	if ctx.MsgSender() != library.Address(account) {
		return errors.New("AccessControl: can only renounce roles for self")
	}

	if err = revokeRole(ctx, role, library.Address(account)); err != nil {
		return errors.Wrap(err, "AccessControl: revokeRole")
	}

	return nil
}
