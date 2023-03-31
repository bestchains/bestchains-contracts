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

import "strconv"

func BytesToUint64(input []byte) (uint64, error) {
	if input == nil {
		return 0, nil
	}
	ui64, err := strconv.ParseUint(string(input), 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

func Uint64ToString(input uint64) string {
	return strconv.FormatUint(input, 10)
}

func BytesToCounter(input []byte) (*Counter, error) {
	if input == nil {
		return &Counter{number: 0}, nil
	}
	count, err := BytesToUint64(input)
	if err != nil {
		return nil, err
	}
	return &Counter{number: count}, nil
}
