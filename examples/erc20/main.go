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
	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/contracts/token/erc20"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	nonceContract := new(nonce.Nonce)
	nonceContract.Name = "org.bestchains.com.NonceContract"
	nonceContract.TransactionContextHandler = new(context.Context)
	nonceContract.BeforeTransaction = context.BeforeTransaction

	erc20Contract := erc20.NewERC20(nonceContract)
	erc20Contract.Contract.Name = "org.bestchains.com.ERC20Contract"
	erc20Contract.TransactionContextHandler = new(context.Context)
	erc20Contract.BeforeTransaction = context.BeforeTransaction

	cc, err := contractapi.NewChaincode(erc20Contract)
	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
