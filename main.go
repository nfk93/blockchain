package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/nfk93/blockchain/consensus"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/p2p"
	"github.com/nfk93/blockchain/transaction"
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
	hardness = flag.Float64("hardness", 0.30, "Specify hardness (default=0.90)")
	newNetwork = flag.Bool("new_network", true, "Set this flag to true if you want to start a new network")
	flag.Parse()
	secretKey, publicKey = crypto.KeyGen(2048)
	_, pk2 = crypto.KeyGen(2048)
	channels = objects.CreateChannelStruct()
	p2p.StartP2P(*addr, *port, publicKey, channels)
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
			prettyPrint("-q", "Exit this program")
			prettyPrint("-start", "Begins the blockchain protocol")
			prettyPrint("-n", "Print out the network list of who you are connected to")
			prettyPrint("-trans", "Print list of seen transactions")
			prettyPrint("-peers", "Print list of all peers in the network")
			prettyPrint("-public-keys", "Print list of know Public keys in network")
			prettyPrint("-ledger", "Print the current ledger")
			prettyPrint("-trans1000", "Transfer 1000k to all known public keys in the network")
			prettyPrint("-trans2", "Transfer an even share of your current stake to everyone in the network including yourself")
			prettyPrint("-trans5", "Transfer 5 random amounts to random accounts")
			prettyPrint("-autotrans", "Starts a routine that makes to random transfers every 10 second as long as you have money.")
		case "-q":
			return
		case "-n":
			p2p.PrintNetworkList()
		case "-trans":
			p2p.PrintTransHashList()
		case "-peers":
			p2p.PrintPeers()
		case "-public-keys":
			p2p.PrintPublicKeys()
		case "-ledger":
			ledger := transaction.GetCurrentLedger()
			for l := range ledger {
				fmt.Printf("Amount %v is owned by %v\n", ledger[l], l[4:14])
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
		case "-trans1000":
			for _, p := range p2p.GetPublicKeys() {
				if p.String() != publicKey.String() {

					trans := objects.CreateTransaction(publicKey,
						p,
						1000,
						publicKey.String()+time.Now().String(),
						secretKey)
					channels.TransClientInput <- trans
				}
			}
		case "-trans2":
			currentStake := consensus.GetLastFinalState()[publicKey.String()]
			for _, p := range p2p.GetPublicKeys() {

				trans := objects.CreateTransaction(publicKey,
					p,
					currentStake/len(p2p.GetPublicKeys()),
					publicKey.String()+time.Now().String(), //+strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- trans

			}
		case "-trans5":
			currentStake := transaction.GetCurrentLedger()[publicKey.String()]
			pkList := p2p.GetPublicKeys()

			if currentStake == 0 {
				continue
			}
			amount := rand.Intn(currentStake / 20)

			for i := 0; i < 5; i++ {
				receiverPK := pkList[rand.Intn(len(pkList))]
				trans := objects.CreateTransaction(publicKey,
					receiverPK,
					amount,
					publicKey.String()+time.Now().String(), //+strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- trans
			}

			// starts Go routine that randomly keeps making transactions to others
		case "-autotrans":
			go func() {
				for {
					currentStake := transaction.GetCurrentLedger()[publicKey.String()]
					pkList := p2p.GetPublicKeys()

					if currentStake == 0 {
						continue
					}
					//noTrans := rand.Intn(5)
					for i := 0; i < 2; i++ {
						receiverPK := pkList[rand.Intn(len(pkList))]
						trans := objects.CreateTransaction(publicKey,
							receiverPK,
							rand.Intn(currentStake/50),
							publicKey.String()+time.Now().String(),
							secretKey)
						channels.TransClientInput <- trans
					}
					time.Sleep(time.Second * 10)
				}
			}()

		default:
			fmt.Println(commandline, "is not a known command. Type -h for help")
		}
	}
}

func prettyPrint(command string, explain string) {
	var buf bytes.Buffer
	buf.WriteString("Command: ")
	buf.WriteString(command)
	for i := 0; i < 15-len(command); i++ {
		buf.WriteString(" ")
	}
	buf.WriteString("-> Action: ")
	buf.WriteString(explain)
	buf.WriteString("\n")

	fmt.Println(buf.String())
}
