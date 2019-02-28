package main

import (
	"flag"
	"fmt"
	"github.com/nfk93/blockchain/code/p2p"
)

func main() {
	var addr = flag.String("a", "", "address to connect to (if not set, start own network)")
	var port = flag.String("p", "65000", "port to be used for p2p (default=65000)")
	flag.Parse()

	p2p.StartP2P(*addr, *port)
	cliLoop()
}

func cliLoop() {
	for {
		var commandline string
		fmt.Print(">")

		fmt.Scanln(&commandline)
		if commandline == "-n" {
			p2p.PrintNetworkList()
		}
	}
}
