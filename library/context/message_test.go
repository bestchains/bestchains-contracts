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

package context_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	// Generate a random private key for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	// Create a new message
	msg := &context.Message{
		Nonce:     123456,
		PublicKey: "",
		Signature: "",
	}

	// Test Marshal method
	t.Run("Marshal", func(t *testing.T) {
		expectedJSON := `{"nonce":123456,"publicKey":"","signature":""}`
		data, err := msg.Marshal()
		assert.NoError(t, err)
		assert.Equal(t, expectedJSON, string(data))
	})

	// Test Unmarshal method
	t.Run("Unmarshal", func(t *testing.T) {
		jsonData := []byte(`{"nonce":123456,"publicKey":"","signature":""}`)
		err := msg.Unmarshal(jsonData)
		assert.NoError(t, err)
		assert.Equal(t, uint64(123456), msg.Nonce)
		assert.Equal(t, "", msg.PublicKey)
		assert.Equal(t, "", msg.Signature)
	})

	// Test Base64EncodedStr and FromBase64EncodedStr methods
	t.Run("Base64EncodedStr and FromBase64EncodedStr", func(t *testing.T) {
		// Encode the message to base64-encoded string
		base64Str, err := msg.Base64EncodedStr()
		assert.NoError(t, err)

		// Create a new message and decode from the base64-encoded string
		decodedMsg := &context.Message{}
		err = decodedMsg.FromBase64EncodedStr(base64Str)
		assert.NoError(t, err)
		assert.Equal(t, msg, decodedMsg)
	})

	// Test GenerateSignature and VerifyAgainstArgs methods
	t.Run("GenerateSignature and VerifyAgainstArgs", func(t *testing.T) {
		// Generate a signature for the message
		err := msg.GenerateSignature(privateKey, "argument1", "argument2")
		assert.NoError(t, err)

		// Verify the signature against the arguments
		addr, err := msg.VerifyAgainstArgs("argument1", "argument2")
		assert.NoError(t, err)

		// Ensure the address is not zero
		assert.NotEqual(t, library.ZeroAddress, addr)
	})

	// Test VerifyAgainstArgs method with invalid arguments
	t.Run("VerifyAgainstArgs with invalid arguments", func(t *testing.T) {
		// Create a new message with a different nonce
		invalidMsg := &context.Message{
			Nonce:     654321,
			PublicKey: msg.PublicKey,
			Signature: msg.Signature,
		}

		// Verify the signature against different arguments (should fail)
		_, err := invalidMsg.VerifyAgainstArgs("argument1", "argument2")
		assert.ErrorIs(t, err, context.ErrInvalidMessage)
	})
}
