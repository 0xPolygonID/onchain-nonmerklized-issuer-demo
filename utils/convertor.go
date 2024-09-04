package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"

	core "github.com/iden3/go-iden3-core/v2"
)

var contractAddress = flag.String("contract_address", "", "Contract address for converting to DID")
var network = flag.String("network", string(core.Polygon), "Network name")
var chain = flag.String("chain", string(core.Amoy), "Chain name")

func main() {
	flag.Parse()
	if *contractAddress == "" {
		log.Fatalln("contract_address is required flag")
	}
	if *network == "" {
		log.Fatalln("network is required flag")
	}
	if *chain == "" {
		log.Fatalln("chain is required flag")
	}

	ethAddrHex := strings.TrimPrefix(*contractAddress, "0x")

	const didMethod = core.DIDMethodIden3
	genesis := genFromHex("00000000000000" + ethAddrHex)
	tp, err := core.BuildDIDType(
		didMethod,
		core.Blockchain(*network),
		core.NetworkID(*chain))
	if err != nil {
		log.Fatalf("failed to build DID type: %v", err)
	}
	id0 := core.NewID(tp, genesis)

	s := fmt.Sprintf("did:%s:%s:%s:%v",
		didMethod, *network, *chain, id0.String())
	fmt.Println("did:", s)
}

func genFromHex(gh string) [27]byte {
	genBytes, err := hex.DecodeString(gh)
	if err != nil {
		panic(err)
	}
	var gen [27]byte
	copy(gen[:], genBytes)
	return gen
}
