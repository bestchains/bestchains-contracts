# Contracts

## DepositoryContract

[`DepositoryContract`](../contracts/depository/interfaces.go) prvodies common logic to store `key-value pair` depository into database with a `counter index`

### Interfaces

```go
// IDepository provides digital depository interfaces
type IDepository interface {
	nonce.INonce
	access.IAccessControl
	// Initialize the contract
	Initialize(ctx context.ContextInterface) error
	// EnableACL enable acl in Depository
	EnableACL(ctx context.ContextInterface) error
	// DisableACL disable acl in Depository
	DisableACL(ctx context.ContextInterface) error
	// Total k/v paris stored
	Total(ctx context.ContextInterface) (uint64, error)
	// PutValue stores kval with pre-defined key calculation
	PutValue(ctx context.ContextInterface, msg context.Message, val string) (string, error)
	// GetValueByIndex get kval with index
	GetValueByIndex(ctx context.ContextInterface, index string) (string, error)
	// GetValueByKID get kval with key id
	GetValueByKID(ctx context.ContextInterface, kid string) (string, error)
}
```