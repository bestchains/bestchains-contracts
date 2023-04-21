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

package main

import (
	"github.com/bestchains/bestchains-contracts/contracts/access"
	"github.com/bestchains/bestchains-contracts/contracts/depository"
	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	depositoryContract := depository.NewDepositoryContract(
		nonce.NewNonceContract(),
		access.NewAccessControlContract(
			access.NewOwnableContract(),
		),
	)
	cc, err := contractapi.NewChaincode(depositoryContract)
	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
