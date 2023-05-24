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
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/pkg/errors"
)

var (
	// ErrNotMessage is returned when the provided data is not a message.
	ErrNotMessage = errors.New("not a message")

	// ErrInvalidMessage is returned when the provided message is invalid.
	ErrInvalidMessage = errors.New("invalid message")

	// ErrInvalidMessageSender is returned when the message sender is invalid.
	ErrInvalidMessageSender = errors.New("invalid message sender")

	// ErrAlgorithmNotSupported is returned when an unsupported algorithm is used for message signing.
	ErrAlgorithmNotSupported = errors.New("algorithm not supported yet")

	// ErrInvalidSignature is returned when the message signature is invalid.
	ErrInvalidSignature = errors.New("invalid signature")
)

// SupportedAlgorithm represents a supported message signing algorithm.
type SupportedAlgorithm string

const (
	// ECDSA is a supported message signing algorithm.
	ECDSA SupportedAlgorithm = "ECDSA"
)

type Message struct {
	Nonce     uint64 `json:"nonce"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
}

// Marshal returns the JSON encoding of a Message struct.
// If the input Message pointer is nil, a new Message is created.
func (msg *Message) Marshal() ([]byte, error) {
	if msg == nil {
		msg = new(Message)
	}
	return json.Marshal(msg)
}

// Unmarshal unmarshals a byte slice into a Message struct
func (msg *Message) Unmarshal(bytes []byte) error {
	var err error

	// If the Message is nil, create a new one
	if msg == nil {
		msg = new(Message)
	}

	// Unmarshal the bytes into the Message struct
	if err = json.Unmarshal(bytes, msg); err != nil {
		// If there is an error, wrap it with ErrNotMessage and return it
		return errors.Wrap(ErrNotMessage, err.Error())
	}

	// Return nil to indicate success
	return nil
}

// Base64EncodedStr returns the base64-encoded string representation of the Message struct.
func (msg *Message) Base64EncodedStr() (string, error) {
	// Marshal the Message struct into bytes.
	bytes, err := msg.Marshal()
	if err != nil {
		// If there was an error during marshaling, return an error with a wrapped message.
		return "", errors.Wrap(ErrInvalidMessage, err.Error())
	}

	// Return the base64-encoded string of the marshalled bytes.
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// FromBase64EncodedStr decodes a base64-encoded string and unmarshals it into a Message struct.
// It returns an error if the string could not be decoded or if the unmarshaling fails.
func (msg *Message) FromBase64EncodedStr(str string) error {
	// Decode the string using base64.StdEncoding.DecodeString
	bytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		// Return an error if the decoding fails
		return errors.Wrap(ErrInvalidMessage, err.Error())
	}

	// Unmarshal the decoded bytes into a Message struct using msg.Unmarshal
	return msg.Unmarshal(bytes)
}

// VerifyAgainstArgs verifies the message against the given arguments and returns the sender's address.
// If the message is invalid, it returns an error.
func (msg *Message) VerifyAgainstArgs(args ...string) (library.Address, error) {
	// Generate the payload using the provided arguments.
	payload := msg.GeneratePayload(args...)

	// Decode the public key from base64.
	rawPubKey, err := base64.StdEncoding.DecodeString(msg.PublicKey)
	if err != nil {
		return library.ZeroAddress, err
	}

	// Parse the public key.
	pub, err := x509.ParsePKIXPublicKey(rawPubKey)
	if err != nil {
		return library.ZeroAddress, errors.Wrap(ErrInvalidMessage, err.Error())
	}

	// Create a new address object for the sender.
	var msgSender = new(library.Address)

	// Convert the public key to an address.
	if err = msgSender.FromPublicKey(pub); err != nil {
		return library.ZeroAddress, err
	}

	// Decode the signature from base64.
	rawSignature, err := base64.StdEncoding.DecodeString(msg.Signature)
	if err != nil {
		return library.ZeroAddress, err
	}

	// Generate the hash of the payload.
	hashedPayload := GenerateHash(payload)

	// Verify the signature against the public key and hashed payload.
	if !VerifySignature(pub, hashedPayload, rawSignature) {
		return library.ZeroAddress, errors.Wrap(ErrInvalidMessage, ErrInvalidSignature.Error())
	}

	// Return the sender's address.
	return *msgSender, nil
}

// VerifySignature verifies that message was signed by the private key corresponding to the provided public key
func VerifySignature(pub crypto.PublicKey, message, sig []byte) bool {
	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		// For ECDSA keys, verify the signature using ASN.1 encoding
		return ecdsa.VerifyASN1(pub, message, sig)
	case ed25519.PublicKey:
		// For Ed25519 keys, verify the signature directly
		return ed25519.Verify(pub, message, sig)
	default:
		// If the public key type is unsupported, return false
		return false
	}
}

// GenerateSignature generates a cryptographic signature for the message using the provided private key and arguments.
// It sets the Signature and PublicKey fields of the Message struct.
func (msg *Message) GenerateSignature(privkey *ecdsa.PrivateKey, args ...string) error {
	// Generate the payload for the message using the provided arguments.
	payload := msg.GeneratePayload(args...)

	// Generate the cryptographic signature for the payload using the provided private key.
	signature, err := ecdsa.SignASN1(rand.Reader, privkey, GenerateHash(payload))
	if err != nil {
		return err
	}

	// Set the Signature field of the message to the base64-encoded signature.
	msg.Signature = base64.StdEncoding.EncodeToString(signature)

	// Set the PublicKey field of the message to the base64-encoded public key of the private key used to generate the signature.
	pubBytes, err := x509.MarshalPKIXPublicKey(&privkey.PublicKey)
	if err != nil {
		return err
	}
	msg.PublicKey = base64.StdEncoding.EncodeToString(pubBytes)

	// Return nil to indicate that no error occurred.
	return nil
}

// GeneratePayload concatenates the message nonce with the provided arguments
// to create a byte slice payload.
// The length of the returned payload will be equal to the length of the nonce
// plus the combined length of all of the arguments.
func (msg *Message) GeneratePayload(args ...string) []byte {
	// Start with the message nonce as the first element of the payload.
	payload := []byte(library.Uint64ToString(msg.Nonce))
	// Append each argument to the payload in sequence.
	for _, arg := range args {
		payload = append(payload, []byte(arg)...)
	}
	return payload
}

// GenerateHash returns the SHA-512 hash of the given payload.
// The returned hash is a byte slice.
func GenerateHash(payload []byte) []byte {
	return sha512.New().Sum(payload[:])
}
