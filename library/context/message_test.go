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

package context

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"testing"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Unmarshal(t *testing.T) {
	// Test valid JSON input
	actualMsg := &Message{
		Nonce:     1,
		PublicKey: []byte("1234567890"),
		Signature: []byte("MTIzNDU2Nzg5MA=="),
	}
	raw, err := actualMsg.Marshal()
	assert.NoError(t, err)

	unmarshalledMsg := new(Message)
	assert.NoError(t, unmarshalledMsg.Unmarshal(raw))

	// Test invalid input
	invalidJSON := `{`
	err = actualMsg.Unmarshal([]byte(invalidJSON))
	assert.ErrorContains(t, err, ErrNotMessage.Error())
}

func TestMessage_VerifyAgainstArgs(t *testing.T) {
	// Generate a private key and address for testing
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	var testAddress library.Address
	err = testAddress.FromPublicKey(&privKey.PublicKey)
	assert.NoError(t, err)

	// Generate a message for testing
	pub, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	assert.NoError(t, err)
	testMsg := &Message{
		Nonce:     1,
		PublicKey: pub,
	}

	assert.NoError(t, testMsg.GenerateSignature(privKey, "arg1", "arg2"))

	// Test valid signature
	msgSender, err := testMsg.VerifyAgainstArgs("arg1", "arg2")
	assert.NoError(t, err)
	assert.Equal(t, testAddress, msgSender)

	// Test invalid signature
	testMsg.Signature = []byte("invalid")
	msgSender, err = testMsg.VerifyAgainstArgs("arg1", "arg2")
	assert.ErrorContains(t, err, ErrInvalidSignature.Error())
	assert.Equal(t, library.ZeroAddress, msgSender)
}

func TestMessage_GeneratePayload(t *testing.T) {
	testMsg := &Message{Nonce: 1}
	actualPayload := testMsg.GeneratePayload("arg1", "arg2")

	expectedPayload := []byte("1arg1arg2")
	assert.Equal(t, expectedPayload, actualPayload)
}
