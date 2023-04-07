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
