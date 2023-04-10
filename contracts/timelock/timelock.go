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

package timelock

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"time"

	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/bestchains/bestchains-contracts/library/timer"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Operation is an operation:
// whose detail is {Payload},
// will do {OpType} work,
// set to be available after {TimeStamp}.
type Operation struct {
	Timestamp timer.TimeStamp
	Payload   []byte
	OpType    string
}

// Entry is a key-value pair.
type Entry struct {
	Key   string
	Value string
}

var _ ITimeLock = new(TimeLock)

// TimeLock provides simple key-value Get/Put with time lock function as a usage example
type TimeLock struct {
	contractapi.Contract
}

// Schedule sets a time lock for a {key}-{value} pair, which will be released for any get/set/del after {duration}.
func (tlc *TimeLock) Schedule(ctx context.ContextInterface, key string, value string, duration int64) (string, error) {

	// Input check
	if key == "" || value == "" {
		return "", fmt.Errorf("input error")
	}

	if duration < 0 {
		return "", fmt.Errorf("duration cannot be negative")
	}

	// Form a dictionary entry
	payload, err := json.Marshal(Entry{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return "", err
	}

	var releaseTime timer.TimeStamp
	releaseTime.SetDeadline(time.Now().Unix() + duration)

	// Emit an operation to be saved
	newOperation := Operation{
		OpType:    "put",
		Timestamp: releaseTime,
		Payload:   payload,
	}

	// Marshal event into json
	event, mErr := json.Marshal(newOperation)
	if mErr != nil {
		return "", mErr
	}

	// Hash event json data as id
	hash := sha3.Sum256(event)
	hashString := hex.EncodeToString(hash[12:])

	// Save json data
	err = ctx.GetStub().PutState(hashString, event)
	if err != nil {
		return "", err
	}

	return hashString, nil
}

// Execute trys to examine whether the given event {hash} has a time lock, and if so, check if it's released.
func (tlc *TimeLock) Execute(ctx context.ContextInterface, opHash string) error {

	// Get json data from ledger
	iByte, err := ctx.GetStub().GetState(opHash)
	if err != nil {
		return fmt.Errorf("get state failed: %s", err)
	}
	if iByte == nil {
		return fmt.Errorf("no data from GetState")
	}

	// Unmarshal event
	var operation Operation
	err = json.Unmarshal(iByte, &operation)
	if err != nil {
		return fmt.Errorf("event unmarshal failed: %s", err)
	}

	// Unmarshal entry
	var entry Entry
	err = json.Unmarshal(operation.Payload, &entry)
	if err != nil {
		return fmt.Errorf("entry unmarshal failed: %s", err)
	}

	// Save the entry
	if operation.Timestamp.IsExpired() {
		if operation.OpType == "put" {
			err = ctx.GetStub().PutState(entry.Key, []byte(entry.Value))
			if err != nil {
				return err
			}
		}
	} else if operation.Timestamp.IsPending() {
		return fmt.Errorf("entry is time locked 'til unix time: %v, the time now is %v", operation.Timestamp.Deadline, time.Now().Unix())
	}
	return nil
}

// GetValue returns the value of the given {key}.
func (tlc *TimeLock) GetValue(ctx context.ContextInterface, key string) (string, error) {
	bytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
