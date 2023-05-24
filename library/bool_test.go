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

func TestBool(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		b := library.True
		assert.Equal(t, "true", b.String())

		b = library.False
		assert.Equal(t, "false", b.String())
	})

	t.Run("Bytes", func(t *testing.T) {
		b := library.True
		assert.Equal(t, []byte("true"), b.Bytes())

		b = library.False
		assert.Equal(t, []byte("false"), b.Bytes())
	})

	t.Run("Bool", func(t *testing.T) {
		b := library.True
		assert.True(t, b.Bool())

		b = library.False
		assert.False(t, b.Bool())
	})
}
