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

package math_test

import (
	"testing"

	"github.com/bestchains/bestchains-contracts/library/math"
	"github.com/stretchr/testify/assert"
)

func TestTryAdd(t *testing.T) {
	t.Run("Addition without overflow", func(t *testing.T) {
		ok, result := math.TryAdd(10, 20)
		assert.True(t, ok)
		assert.Equal(t, uint64(30), result)
	})

	t.Run("Addition with overflow", func(t *testing.T) {
		ok, result := math.TryAdd(18446744073709551615, 1)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), result)
	})
}

func TestTrySub(t *testing.T) {
	t.Run("Subtraction without underflow", func(t *testing.T) {
		ok, result := math.TrySub(30, 20)
		assert.True(t, ok)
		assert.Equal(t, uint64(10), result)
	})

	t.Run("Subtraction with underflow", func(t *testing.T) {
		ok, result := math.TrySub(10, 20)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), result)
	})
}

func TestTryMul(t *testing.T) {
	t.Run("Multiplication without overflow", func(t *testing.T) {
		ok, result := math.TryMul(10, 20)
		assert.True(t, ok)
		assert.Equal(t, uint64(200), result)
	})

	t.Run("Multiplication with overflow", func(t *testing.T) {
		ok, result := math.TryMul(18446744073709551615, 2)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), result)
	})
}

func TestTryDiv(t *testing.T) {
	t.Run("Division without error", func(t *testing.T) {
		ok, result := math.TryDiv(30, 5)
		assert.True(t, ok)
		assert.Equal(t, uint64(6), result)
	})

	t.Run("Division by zero", func(t *testing.T) {
		ok, result := math.TryDiv(10, 0)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), result)
	})
}

func TestTryMod(t *testing.T) {
	t.Run("Modulus without error", func(t *testing.T) {
		ok, result := math.TryMod(10, 3)
		assert.True(t, ok)
		assert.Equal(t, uint64(1), result)
	})

	t.Run("Modulus by zero", func(t *testing.T) {
		ok, result := math.TryMod(10, 0)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), result)
	})
}
