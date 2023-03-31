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
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TestECDSAAddress = "0x2b5715a46e48462258fca67c53dee748f77755b6"

	TestInvalidAddress1 Address = "0x2b5715a46e48462258fca67c53dee748f77755b699"
	TestInvalidAddress2 Address = "412b5715a46e48462258fca67c53dee748f77755b6"
	TestInvalidAddress3 Address = "0x&2b5715a46e48462258fca67c53dee748f77755b"
	TestInvalidAddress4 Address = "0x0000000000000000000000000000000000000000"
)

func TestAddressGenerate(t *testing.T) {
	testECDSAPrivatekeyBytes, _ := os.ReadFile("./testdata/ecdsa.pem")
	block, _ := pem.Decode([]byte(testECDSAPrivatekeyBytes))
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	assert.Nil(t, err)

	ecdsaPrivKey, ok := privKey.(*ecdsa.PrivateKey)
	assert.True(t, ok)

	addr := new(Address)
	assert.Nil(t, addr.FromPublicKey(&ecdsaPrivKey.PublicKey))
	assert.Equal(t, TestECDSAAddress, addr.String())
}

func TestAddressValidate(t *testing.T) {
	assert.Equal(t, ErrInvalidAddressLength, TestInvalidAddress1.Validate())
	assert.Equal(t, ErrInvalidAddressMissingPrefix, TestInvalidAddress2.Validate())
	assert.Equal(t, ErrInvalidAddressBadCharacters, TestInvalidAddress3.Validate())
	assert.Equal(t, ErrInvalidAddressNull, TestInvalidAddress4.Validate())
}
