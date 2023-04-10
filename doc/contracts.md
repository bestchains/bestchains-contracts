# Contracts

## BasicContract

[`BasicContract`](../contracts/basic/interfaces.go) prvodies common logic to store `key-value pair` into database with a `counter index`


### Interfaces

```go
// IBasic provides common data Put/Get
type IBasic interface {
	// Total k/v paris stored
	Total(ctx context.ContextInterface) (uint64, error)
	// PutValue stores kval with pre-defined key calculation
	PutValue(ctx context.ContextInterface, val string) (string, error)
	// GetValueByIndex get kval with index
	GetValueByIndex(ctx context.ContextInterface, index string) (string, error)
	// GetValueByKID get kval with key id
	GetValueByKID(ctx context.ContextInterface, kid string) (string, error)
}

```

#### `Total`

1. description

returns total key-value pairs currently stored
2. args

No arguments required
3. returns

- `uint64`

total number

- `error`

error message if get total number failed

#### `PutValue`

1. description
stores value into database with dynamically calcuated `kid` with current counter index and value payload

2. args

- `val`: the value in bytes

3. returns

- `uint64`

the counter index for this `val`

- `string`

the hex encoded `kid` string for this `val`

- `error`

the error message if `PutValue` failed

##### How we calcuate key id?

```go
func calculateKID(counter *library.Counter, val []byte) string {
	hashedPubKey := sha3.Sum256(append(counter.Bytes(), val...))
	return hex.EncodeToString(hashedPubKey[12:])
}
```

arguments explained:

- [`counter`](../library/counter.go) which has the counter number which retrieved from statedb
- `val` wihch is the input `value`

#### `GetValueByIndex`

1. description
get value with `counter index`

2. args

- `index`


3. returns

- `string`

the val in bytes

- `error`

the error message if `GetValue` failed

#### `GetValueByKID`

1. description
get value with `key id`

2. args

- `kid`

3. returns

- `string`

the val in bytes

- `error`

the error message if `GetValueByKID` failed

### Events

1. `EventPutValue`

Emitted when `PutValue` successfully

```go
type EventPutValue struct {
	Index uint64
	KID     string
}
```

- `Index` to this `Value`
- `KID` to this `Value`
