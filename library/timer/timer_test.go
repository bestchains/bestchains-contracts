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

package timer_test

import (
	"testing"
	"time"

	"github.com/bestchains/bestchains-contracts/library/timer"
	"github.com/stretchr/testify/assert"
)

func TestTimeStamp(t *testing.T) {
	t.Run("GetDeadline", func(t *testing.T) {
		ts := timer.TimeStamp{Deadline: 12345}
		assert.Equal(t, int64(12345), ts.GetDeadline())
	})

	t.Run("SetDeadline", func(t *testing.T) {
		ts := timer.TimeStamp{}
		ts.SetDeadline(54321)
		assert.Equal(t, int64(54321), ts.GetDeadline())
	})

	t.Run("Reset", func(t *testing.T) {
		ts := timer.TimeStamp{Deadline: 12345}
		ts.Reset()
		assert.Equal(t, int64(0), ts.GetDeadline())
	})

	t.Run("IsUnset", func(t *testing.T) {
		ts := timer.TimeStamp{}
		assert.True(t, ts.IsUnset())
		ts.SetDeadline(12345)
		assert.False(t, ts.IsUnset())
	})

	t.Run("IsStarted", func(t *testing.T) {
		ts := timer.TimeStamp{}
		assert.False(t, ts.IsStarted())
		ts.SetDeadline(12345)
		assert.True(t, ts.IsStarted())
	})

	t.Run("IsPending", func(t *testing.T) {
		now := time.Now().Unix()
		ts := timer.TimeStamp{Deadline: now + 10}
		assert.True(t, ts.IsPending())
		ts.SetDeadline(now - 10)
		assert.False(t, ts.IsPending())
	})

	t.Run("IsExpired", func(t *testing.T) {
		now := time.Now().Unix()
		ts := timer.TimeStamp{Deadline: now + 10}
		assert.False(t, ts.IsExpired())
		ts.SetDeadline(now - 10)
		assert.True(t, ts.IsExpired())
	})
}
