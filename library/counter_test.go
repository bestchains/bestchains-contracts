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

package library_test

import (
	"testing"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/math"
	"github.com/stretchr/testify/assert"
)

func TestNewCounter(t *testing.T) {
	initNumber := uint64(42)
	counter := library.NewCounter(initNumber)
	assert.NotNil(t, counter)
	assert.Equal(t, initNumber, counter.Current())
}

func TestCounterString(t *testing.T) {
	t.Run("NilCounter", func(t *testing.T) {
		counter := (*library.Counter)(nil)
		str := counter.String()
		assert.Equal(t, "", str)
	})

	t.Run("ValidCounter", func(t *testing.T) {
		initNumber := uint64(42)
		counter := library.NewCounter(initNumber)
		str := counter.String()
		assert.Equal(t, "42", str)
	})
}

func TestCounterBytes(t *testing.T) {
	t.Run("NilCounter", func(t *testing.T) {
		counter := (*library.Counter)(nil)
		bytes := counter.Bytes()
		assert.Empty(t, bytes)
	})

	t.Run("ValidCounter", func(t *testing.T) {
		initNumber := uint64(42)
		counter := library.NewCounter(initNumber)
		bytes := counter.Bytes()
		assert.Equal(t, []byte("42"), bytes)
	})
}

func TestCounterCurrent(t *testing.T) {
	initNumber := uint64(42)
	counter := library.NewCounter(initNumber)
	current := counter.Current()
	assert.Equal(t, initNumber, current)
}

func TestCounterIncrement(t *testing.T) {
	t.Run("NilCounter", func(t *testing.T) {
		counter := (*library.Counter)(nil)
		err := counter.Increment(10)
		assert.Equal(t, library.ErrNilCounter, err)
	})

	t.Run("ValidIncrement", func(t *testing.T) {
		initNumber := uint64(42)
		offset := uint64(10)
		counter := library.NewCounter(initNumber)
		err := counter.Increment(offset)
		assert.NoError(t, err)
		assert.Equal(t, initNumber+offset, counter.Current())
	})

	t.Run("Overflow", func(t *testing.T) {
		initNumber := uint64(18446744073709551615) // Max uint64 value
		offset := uint64(10)
		counter := library.NewCounter(initNumber)
		err := counter.Increment(offset)
		assert.Equal(t, math.ErrMathOpOverflowed, err)
		assert.Equal(t, initNumber, counter.Current())
	})
}

func TestCounterDecrement(t *testing.T) {
	t.Run("NilCounter", func(t *testing.T) {
		counter := (*library.Counter)(nil)
		err := counter.Decrement(10)
		assert.Equal(t, library.ErrNilCounter, err)
	})

	t.Run("ValidDecrement", func(t *testing.T) {
		initNumber := uint64(42)
		offset := uint64(10)
		counter := library.NewCounter(initNumber)
		err := counter.Decrement(offset)
		assert.NoError(t, err)
		assert.Equal(t, initNumber-offset, counter.Current())
	})

	t.Run("Overflow", func(t *testing.T) {
		initNumber := uint64(10)
		offset := uint64(20)
		counter := library.NewCounter(initNumber)
		err := counter.Decrement(offset)
		assert.Equal(t, math.ErrMathOpOverflowed, err)
		assert.Equal(t, initNumber, counter.Current())
	})
}

func TestCounterReset(t *testing.T) {
	initNumber := uint64(42)
	counter := library.NewCounter(initNumber)
	counter.Reset()
	assert.Equal(t, uint64(0), counter.Current())
}
