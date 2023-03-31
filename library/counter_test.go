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

func TestCounter(t *testing.T) {
	c := &Counter{}
	assert.Equal(t, uint64(0), c.Current())

	assert.Equal(t, []byte(Uint64ToString(0)), c.Bytes())

	c.Increment()
	assert.Equal(t, uint64(1), c.Current())

	c.Increment()
	assert.Equal(t, uint64(2), c.Current())

	c.Decrement()
	assert.Equal(t, uint64(1), c.Current())

	c.Reset()

	assert.Equal(t, uint64(0), c.Current())
}
