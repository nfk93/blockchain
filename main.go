package main

import (
	"flag"
	"fmt"
	"github.com/nfk93/blockchain/consensus"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/p2p"
)

var p2p_blockIn chan objects.Block
var p2p_blockOut chan objects.Block
var p2p_transactionIn chan objects.Transaction
var p2p_transactionOut chan objects.Transaction
var secretKey crypto.SecretKey
var publicKey crypto.PublicKey

func main() {
	var addr = flag.String("a", "", "address to connect to (if not set, start own network)")
	var port = flag.String("p", "65000", "port to be used for p2p (default=65000)")
	flag.Parse()

	secretKey, publicKey = crypto.KeyGen(2048)

	p2p_blockIn = make(chan objects.Block)
	p2p_blockOut = make(chan objects.Block)
	p2p_transactionIn = make(chan objects.Transaction)
	p2p_transactionOut = make(chan objects.Transaction)

	p2p.StartP2P(*addr, *port, p2p_blockIn, p2p_blockOut, p2p_transactionIn, p2p_transactionOut)
	consensus.StartConsensus(p2p_transactionOut, p2p_blockOut, p2p_blockIn)
	cliLoop()
}

func cliLoop() {
	for {
		// TODO use a variable to track latest printed cmdline entry, to always have '>' as the latest line printed
		var commandline string
		fmt.Print(">")

		fmt.Scanln(&commandline)
		if commandline == "-n" {
			p2p.PrintNetworkList()
		} else if commandline == "-send-test-block" {
			fmt.Println("NOT IMPLEMENTED")
		} else if commandline == "-send-test-trans" {
			p2p_transactionIn <- objects.Transaction{publicKey, publicKey, 123, "id1", "sign1"}
		} else if commandline == "-trans" {
			p2p.PrintTransHashList()
		} else if commandline == "-peers" {
			p2p.PrintPeers()
		}
	}
}
