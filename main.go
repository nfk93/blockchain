package main

import (
	"flag"
	"fmt"
	"github.com/nfk93/blockchain/consensus"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/p2p"
	"log"
	"math/rand"
	"time"
)

var channels objects.ChannelStruct
var secretKey crypto.SecretKey
var publicKey crypto.PublicKey
var pk2 crypto.PublicKey

var slotduration *int
var hardness *float64
var newNetwork *bool
var genesisBlock objects.Block

func main() {
	var addr = flag.String("a", "", "Address to connect to (if not set, start own network)")
	var port = flag.String("p", "65000", "Port to be used for p2p (default=65000)")
	slotduration = flag.Int("slot_duration", int(time.Second*1), "Specify the slot length (default=10sec)")
	hardness = flag.Float64("hardness", 0.10, "Specify hardness (default=0.90)")
	newNetwork = flag.Bool("new_network", true, "Set this flag to true if you want to start a new network")
	flag.Parse()
	secretKey, publicKey = crypto.KeyGen(2048)
	_, pk2 = crypto.KeyGen(2048)
	channels = objects.CreateChannelStruct()
	p2p.StartP2P(*addr, *port, channels)
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
		case "-q":
			return
		case "-n":
			p2p.PrintNetworkList()
		case "-send-test-block":
			fmt.Println("Not Implemented")
		case "-send-test-trans":
			channels.TransClientInput <- objects.Transaction{publicKey, pk2, 123, "id1", "sign1"}
		case "-trans":
			p2p.PrintTransHashList()
		case "-peers":
			p2p.PrintPeers()
		case "-ledger":
			ledger := consensus.GetLastFinalState()
			for l := range ledger {
				fmt.Printf("Amount %v is owned by %v", ledger[l], l)
			}
		case "-start": //"-start_network":
			if *newNetwork {
				genesisdata, err := objects.NewGenesisData(publicKey, time.Duration(*slotduration), *hardness)
				if err != nil {
					log.Fatal(err)
				}
				genesisblock := objects.Block{Slot: 0, BlockData: objects.Data{GenesisData: genesisdata}}
				channels.BlockToP2P <- genesisblock
			} else {
				fmt.Println("Only the network founder can start the network!")
			}
		case "-test1":
			for p := range consensus.GetLastFinalState() {
				trans := objects.CreateTransaction(publicKey,
					p,
					100000,
					publicKey.String()+time.Now().String(), //+strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- trans
			}
		case "-test2":
			finalizedLedger := consensus.GetLastFinalState()
			currentStake := finalizedLedger[publicKey]
			var peerList []crypto.PublicKey
			for p := range finalizedLedger {
				peerList = append(peerList, p)
			}

			if currentStake == 0 {
				continue
			}
			amount := rand.Intn(currentStake / 20)

			for i := 0; i < 5; i++ {
				receiverPK := peerList[rand.Intn(len(peerList))]
				trans := objects.CreateTransaction(publicKey,
					receiverPK,
					amount,
					publicKey.String()+time.Now().String(), //+strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- trans
			}

		default:
			fmt.Println(commandline, "is not a known command. Type -h for help")
		}
	}
}
