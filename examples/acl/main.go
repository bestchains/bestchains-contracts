package main

import (
	"github.com/bestchains/bestchains-contracts/contracts/access"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	aclContract := access.NewAccessControlContract(access.NewOwnableContract())
	cc, err := contractapi.NewChaincode(aclContract)
	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
