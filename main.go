package main

import (
	"flag"
	"fmt"
	"github.com/nfk93/blockchain/consensus"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/p2p"
	"log"
	"time"
)

var channels objects.ChannelStruct
var secretKey crypto.SecretKey
var publicKey crypto.PublicKey

var slotduration *int
var hardness *float64
var newNetwork *bool
var genesisBlock objects.Block

func main() {
	var addr = flag.String("a", "", "Address to connect to (if not set, start own network)")
	var port = flag.String("p", "65000", "Port to be used for p2p (default=65000)")
	slotduration = flag.Int("slot_duration", int(time.Second*10), "Specify the slot length (default=10sec)")
	hardness = flag.Float64("hardness", 0.90, "Specify hardness (default=0.90)")
	newNetwork = flag.Bool("new_network", true, "Set this flag to true if you want to start a new network")
	flag.Parse()

	secretKey, publicKey = crypto.KeyGen(2048)
	channels = objects.CreateChannelStruct()
	p2p.StartP2P(*addr, *port, channels.BlockToP2P, channels.BlockFromP2P, channels.TransClientInput, channels.TransFromP2P)
	consensus.StartConsensus(channels, publicKey, secretKey, true)

	cliLoop()
}

func cliLoop() {
	for {
		// TODO use a variable to track latest printed cmdline entry, to always have '>' as the latest line printed
		var commandline string
		fmt.Print(">")

		fmt.Scanln(&commandline)
		switch commandline {
		case "-h":
			fmt.Println("NOT IMPLEMENTED")
		case "-n":
			p2p.PrintNetworkList()
		case "-send-test-block":
			channels.BlockToP2P <- objects.Block{1, genesisBlock.CalculateBlockHash(), publicKey,
				"", "", genesisBlock.CalculateBlockHash(), objects.Data{}, ""}
		case "-send-test-trans":
			channels.TransClientInput <- objects.Transaction{publicKey, publicKey, 123, "id1", "sign1"}
		case "-trans":
			p2p.PrintTransHashList()
		case "-peers":
			p2p.PrintPeers()
		case "-start_network":
			if *newNetwork {
				genesisdata, err := objects.NewGenesisData(publicKey, secretKey, time.Duration(*slotduration), *hardness)
				if err != nil {
					log.Fatal(err)
				}
				genesisblock := objects.Block{Slot: 0, BlockData: objects.Data{GenesisData: genesisdata}}
				channels.BlockFromP2P <- genesisblock
				channels.BlockToP2P <- genesisblock
			} else {
				fmt.Println("Only the network founder can start the network!")
			}
		default:
			fmt.Println(commandline, "is not a known command. Type -h for help")
		}
	}
}
