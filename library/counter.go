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

package library

import (
	"errors"

	safemath "github.com/bestchains/bestchains-contracts/library/math"
)

var (
	// ErrNilCounter is returned when a nil counter is used.
	ErrNilCounter = errors.New("nil counter")
)

// Counter is a data type that represents a counter that can be incremented,
// decremented, and reset.
type Counter struct {
	number uint64
}

// NewCounter creates a new counter with the specified initial value.
func NewCounter(initNumber uint64) *Counter {
	return &Counter{
		number: initNumber,
	}
}

// String returns the string representation of the counter.
func (counter *Counter) String() string {
	if counter == nil {
		return ""
	}
	return Uint64ToString(counter.number)
}

// Bytes returns the byte slice representation of the counter.
func (counter *Counter) Bytes() []byte {
	if counter == nil {
		return []byte{}
	}
	return []byte(Uint64ToString(counter.number))
}

// Current returns the current value of the counter.
func (counter *Counter) Current() uint64 {
	return counter.number
}

// Increment adds the specified offset to the counter.
func (counter *Counter) Increment(offset uint64) error {
	if counter == nil {
		return ErrNilCounter
	}
	succ, newCount := safemath.TryAdd(counter.Current(), offset)
	if !succ {
		return safemath.ErrMathOpOverflowed
	}
	counter.number = newCount
	return nil
}

// Decrement subtracts the specified offset from the counter.
func (counter *Counter) Decrement(offset uint64) error {
	if counter == nil {
		return ErrNilCounter
	}
	succ, newCount := safemath.TrySub(counter.number, offset)
	if !succ {
		return safemath.ErrMathOpOverflowed
	}
	counter.number = newCount
	return nil
}

// Reset sets the counter to zero.
func (counter *Counter) Reset() {
	counter.number = 0
}
