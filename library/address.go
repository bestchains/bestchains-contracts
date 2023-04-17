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
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"regexp"

	"github.com/pkg/errors"

	"golang.org/x/crypto/sha3"
)

var (
	ErrInvalidAddressLength        = errors.New("invalid address length (42)")
	ErrInvalidAddressMissingPrefix = errors.New("missing prefix 0x")
	ErrInvalidAddressBadCharacters = errors.New("contains invalid characters")
	ErrInvalidAddressNull          = errors.New("null address")
	ErrUnknownAddressAlg           = errors.New("unknown crypto algorithm for address")
)

const (
	AddressPrefix         = "0x"
	ZeroAddress   Address = "0x0000000000000000000000000000000000000000"
)

// Address is the hex string of ethereum account address
type Address string

func (addr Address) String() string {
	return string(addr)
}

func (addr Address) Bytes() []byte {
	return []byte(addr)
}

func (addr *Address) FromString(addrStr string) {
	*addr = Address(addrStr)
}

func (addr *Address) FromPublicKey(pub interface{}) error {
	publicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return ErrUnknownAddressAlg
	}
	// Serialize the public key
	serializedPubKey := elliptic.Marshal(elliptic.P256(), publicKey.X, publicKey.Y)

	// Hash the public key using Keccak-256
	hashedPubKey := sha3.Sum256(serializedPubKey)

	// Truncate the hash and add a prefix to get the final Ethereum address
	addr.FromString(AddressPrefix + hex.EncodeToString(hashedPubKey[12:]))

	return nil
}

func (addr Address) EmptyAddress() bool {
	if addr == ZeroAddress || addr == "" || addr == AddressPrefix {
		return true
	}
	return false
}

func (addr Address) Validate() error {
	// Check length
	if len(addr) != 42 {
		return ErrInvalidAddressLength
	}

	// Check prefix
	if addr[:2] != "0x" {
		return ErrInvalidAddressMissingPrefix
	}

	// Check hex format
	match, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", addr.String())
	if !match {
		return ErrInvalidAddressBadCharacters
	}

	// Check null address
	if addr == ZeroAddress {
		return ErrInvalidAddressNull
	}

	return nil
}
