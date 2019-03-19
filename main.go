package main

import (
	"flag"
	"fmt"
	"github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/p2p"
)

func main() {
	var addr = flag.String("a", "", "address to connect to (if not set, start own network)")
	var port = flag.String("p", "65000", "port to be used for p2p (default=65000)")
	flag.Parse()

	p2p_blockIn := make(chan objects.Block)
	p2p_blockOut := make(chan objects.Block)
	p2p_transactionIn := make(chan objects.Transaction)
	p2p_transactionOut := make(chan objects.Transaction)

	p2p.StartP2P(*addr, *port, p2p_blockIn, p2p_blockOut, p2p_transactionIn, p2p_transactionOut)
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
		}
	}
}
