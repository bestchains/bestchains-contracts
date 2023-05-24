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
	"github.com/stretchr/testify/assert"
)

func TestBytesToUint64(t *testing.T) {
	t.Run("NilInput", func(t *testing.T) {
		ui64, err := library.BytesToUint64(nil)
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), ui64)
	})

	t.Run("ValidInput", func(t *testing.T) {
		input := []byte("12345")
		ui64, err := library.BytesToUint64(input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(12345), ui64)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		input := []byte("invalid")
		_, err := library.BytesToUint64(input)
		assert.Error(t, err)
	})
}

func TestUint64ToString(t *testing.T) {
	input := uint64(12345)
	str := library.Uint64ToString(input)
	assert.Equal(t, "12345", str)
}

func TestBytesToCounter(t *testing.T) {
	t.Run("NilInput", func(t *testing.T) {
		counter, err := library.BytesToCounter(nil)
		assert.NoError(t, err)
		assert.Equal(t, library.NewCounter(0), counter)
	})

	t.Run("ValidInput", func(t *testing.T) {
		input := []byte("12345")
		counter, err := library.BytesToCounter(input)
		assert.NoError(t, err)
		assert.Equal(t, library.NewCounter(12345), counter)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		input := []byte("invalid")
		_, err := library.BytesToCounter(input)
		assert.Error(t, err)
	})
}

func TestBytesToHexString(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03, 0x04}
	hexStr := library.BytesToHexString(input)
	assert.Equal(t, "01020304", hexStr)
}

func TestBytesToUint8(t *testing.T) {
	t.Run("NilInput", func(t *testing.T) {
		ui8, err := library.BytesToUint8(nil)
		assert.NoError(t, err)
		assert.Equal(t, uint8(0), ui8)
	})

	t.Run("ValidInput", func(t *testing.T) {
		input := []byte("42")
		ui8, err := library.BytesToUint8(input)
		assert.NoError(t, err)
		assert.Equal(t, uint8(42), ui8)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		input := []byte("invalid")
		_, err := library.BytesToUint8(input)
		assert.Error(t, err)
	})
}

func TestBytesToBool(t *testing.T) {
	t.Run("NilInput", func(t *testing.T) {
		b := library.BytesToBool(nil)
		assert.Equal(t, library.False, b)
	})

	t.Run("TrueInput", func(t *testing.T) {
		input := []byte("true")
		b := library.BytesToBool(input)
		assert.Equal(t, library.True, b)
	})

	t.Run("FalseInput", func(t *testing.T) {
		input := []byte("false")
		b := library.BytesToBool(input)
		assert.Equal(t, library.False, b)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		input := []byte("invalid")
		b := library.BytesToBool(input)
		assert.Equal(t, library.False, b)
	})
}
