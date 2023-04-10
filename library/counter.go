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

import "errors"

var (
	ErrNilCounter = errors.New("nil counter")
)

type Counter struct {
	number uint64
}

func (counter *Counter) String() string {
	if counter == nil {
		return ""
	}
	return Uint64ToString(counter.number)
}

func (counter *Counter) Bytes() []byte {
	if counter == nil {
		return []byte{}
	}
	return []byte(Uint64ToString(counter.number))
}

func (counter *Counter) Current() uint64 {
	return counter.number
}

func (counter *Counter) Increment() {
	counter.number++
}

func (counter *Counter) Decrement() {
	counter.number--
}

func (counter *Counter) Reset() {
	counter.number = 0
}
