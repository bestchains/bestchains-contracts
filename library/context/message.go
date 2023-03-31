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
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"

	"github.com/bestchains/bestchains-contracts/library"
	"github.com/pkg/errors"
)

var (
	ErrNotMessage            = errors.New("not a message")
	ErrInvalidMessage        = errors.New("invalid message")
	ErrInvalidMessageSender  = errors.New("invalid message sender")
	ErrAlgorithmNotSupported = errors.New("algorithm not supported yet")
	ErrInvalidSignature      = errors.New("invalid signature")
)

type SupportedAlgorithm string

const (
	ECDSA SupportedAlgorithm = "ECDSA"
)

type Message struct {
	Nonce     uint64 `json:"nonce"`
	PublicKey []byte `json:"publicKey"`
	Signature []byte `json:"signature"`
}

func (msg *Message) Marshal() ([]byte, error) {
	if msg == nil {
		msg = new(Message)
	}
	return json.Marshal(msg)
}

func (msg *Message) Unmarshal(bytes []byte) error {
	var err error
	if msg == nil {
		msg = new(Message)
	}
	if err = json.Unmarshal(bytes, msg); err != nil {
		return errors.Wrap(ErrNotMessage, err.Error())
	}

	return nil
}

func (msg *Message) VerifyAgainstArgs(args ...string) (library.Address, error) {
	payload := msg.GeneratePayload(args...)

	pub, err := x509.ParsePKIXPublicKey(msg.PublicKey)
	if err != nil {
		return library.ZeroAddress, errors.Wrap(ErrInvalidMessage, err.Error())
	}

	var msgSender = new(library.Address)
	if err = msgSender.FromPublicKey(pub); err != nil {
		return library.ZeroAddress, err
	}

	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		hashedPayload := GenerateHash(payload)
		if !ecdsa.VerifyASN1(pub, hashedPayload, msg.Signature) {
			return library.ZeroAddress, errors.Wrap(ErrInvalidMessage, ErrInvalidSignature.Error())
		}
	default:
		return library.ZeroAddress, ErrAlgorithmNotSupported
	}

	return *msgSender, nil
}

func (msg *Message) GenerateSignature(privkey *ecdsa.PrivateKey, args ...string) error {
	payload := msg.GeneratePayload(args...)
	signature, err := ecdsa.SignASN1(rand.Reader, privkey, GenerateHash(payload))
	if err != nil {
		return err
	}
	msg.Signature = signature

	pubBytes, err := x509.MarshalPKIXPublicKey(&privkey.PublicKey)
	if err != nil {
		return err
	}
	msg.PublicKey = pubBytes

	return nil
}

func (msg *Message) GeneratePayload(args ...string) []byte {
	payload := []byte(library.Uint64ToString(msg.Nonce))
	for _, arg := range args {
		payload = append(payload, []byte(arg)...)
	}
	return payload
}

func GenerateHash(payload []byte) []byte {
	return sha512.New().Sum(payload[:])
}
