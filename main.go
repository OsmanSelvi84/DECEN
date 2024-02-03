package main

import(
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"ChainedRelations/chaincodeTranscript"
	"log"	
)

func main(){

	newTranscript, err := contractapi.NewChaincode(&chaincodeTranscript.SmartContract{})
	if err != nil {
		log.Panic(err)
	}

	if err := newTranscript.Start(); err != nil {
		log.Panic(err)
	}

}
