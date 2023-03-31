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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToUint64(t *testing.T) {
	val, err := BytesToUint64(nil)
	assert.Equal(t, uint64(0), val)
	assert.Nil(t, nil, err)

	val, err = BytesToUint64([]byte("invalid uint64"))
	assert.Equal(t, uint64(0), val)
	assert.NotNil(t, err)

	val, err = BytesToUint64([]byte("12"))
	assert.Equal(t, uint64(12), val)
	assert.Nil(t, err)
}

func TestUint64ToString(t *testing.T) {
	assert.Equal(t, Uint64ToString(12), "12")
}

func TestBytesToCounter(t *testing.T) {
	counter, err := BytesToCounter(nil)
	assert.Equal(t, uint64(0), counter.Current())
	assert.Nil(t, err)

	counter, err = BytesToCounter([]byte("invalid number"))
	assert.Nil(t, counter)
	assert.NotNil(t, err)

	counter, err = BytesToCounter([]byte("12"))
	assert.Equal(t, uint64(12), counter.Current())
	assert.Nil(t, err)
}
