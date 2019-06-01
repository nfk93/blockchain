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
	"strconv"
	"time"
)

var channels objects.ChannelStruct
var secretKey crypto.SecretKey
var publicKey crypto.PublicKey
var pk2 crypto.PublicKey

var slotduration *int
var hardness *float64
var newNetwork *bool
var saveLogFile *bool
var addr *string
var port *string
var autoTransStatus bool

func main() {
	addr = flag.String("a", "", "Address to connect to (if not set, start own network)")
	port = flag.String("p", "65000", "Port to be used for p2p (default=65000)")
	slotduration = flag.Int("slot_duration", 1, "Specify the slot length (default=1sec)")
	hardness = flag.Float64("hardness", 0.2, "Specify hardness (default=0.2)")
	newNetwork = flag.Bool("new_network", true, "Set this flag to true if you want to start a new network")
	saveLogFile = flag.Bool("log", false, "will save logs of all transactions and blocks if true (default=false)")
	flag.Parse()
	secretKey, publicKey = crypto.KeyGen(2048)
	_, pk2 = crypto.KeyGen(2048)
	channels = objects.CreateChannelStruct()
	autoTransStatus = false
	p2p.StartP2P(*addr, *port, publicKey, channels)
	consensus.StartConsensus(channels, publicKey, secretKey, true, *saveLogFile)
	if *addr == "" {
		fmt.Println("When all other clients are ready, use -start to begin the Blockchain protocol or -h for help with further commands!")
	} else {
		fmt.Println("This client is ready for the Blockchain protocol to start. Use -h for further commands!")
	}
	cliLoop()
}

func cliLoop() {
	for {
		// TODO: make cli pretty
		var commandline string
		fmt.Print(">")

		fmt.Scanln(&commandline)
		switch commandline {
		case "-h":
			prettyPrintHelpMessage("-q", "Exit this program")
			prettyPrintHelpMessage("-start", "Begins the blockchain protocol")
			prettyPrintHelpMessage("-verbose", "Initially active. Switches verbose mode between on and off")
			prettyPrintHelpMessage("-id", "Show your public key in short and which port you are listening on")
			prettyPrintHelpMessage("-n", "Print out the network list of who you are connected to")
			prettyPrintHelpMessage("-trans", "Print list of seen transactions")
			prettyPrintHelpMessage("-peers", "Print list of all peers in the network")
			prettyPrintHelpMessage("-public-keys", "Print list of know Public keys in network")
			prettyPrintHelpMessage("-ledger", "Print the current ledger")
			prettyPrintHelpMessage("-final", "Print the last finalized ledger")
			prettyPrintHelpMessage("-trans1000", "Transfer 1000k to all known public keys in the network")
			prettyPrintHelpMessage("-trans2", "Transfer an even share of your current stake to everyone in the network including yourself")
			prettyPrintHelpMessage("-trans5", "Transfer 5 random amounts to random accounts")
			prettyPrintHelpMessage("-autotrans", "Initially active. Starts or stops automatic transfers to random accounts every 5 second")
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
			transaction.PrintCurrentLedger()
		case "-final":
			consensus.PrintFinalizedLedger()
		case "-start": //"-start_network":
			if *newNetwork {
				genesisdata, err := objects.NewGenesisData(publicKey, time.Second*time.Duration(*slotduration), *hardness)
				if err != nil {
					log.Fatal(err)
				}
				genesisblock := objects.Block{Slot: 0, BlockData: objects.BlockData{GenesisData: genesisdata}}
				channels.BlockToP2P <- genesisblock
				publicKeys := p2p.GetPublicKeys()
				for i, v := range publicKeys {
					if i > 4 {
						break
					} else {
						trans := objects.CreateTransaction(publicKey,
							v,
							uint64(10000000000000),
							strconv.Itoa(i),
							secretKey)
						channels.TransClientInput <- objects.TransData{Transaction: trans}
					}
				}
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
					channels.TransClientInput <- objects.TransData{Transaction: trans}
				}
			}
		case "-trans2":
			currentStake := consensus.GetLastFinalState()[publicKey.Hash()]
			for i, p := range p2p.GetPublicKeys() {

				trans := objects.CreateTransaction(publicKey,
					p,
					currentStake/uint64(len(p2p.GetPublicKeys())),
					strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- objects.TransData{Transaction: trans}

			}
		case "-trans5":
			currentStake := transaction.GetCurrentLedger()[publicKey.Hash()]
			pkList := p2p.GetPublicKeys()

			if currentStake == 0 {
				continue
			}
			amount := uint64(rand.Intn(int(currentStake) / 20))

			for i := 0; i < 5; i++ {
				receiverPK := pkList[rand.Intn(len(pkList))]
				trans := objects.CreateTransaction(publicKey,
					receiverPK,
					amount,
					strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- objects.TransData{Transaction: trans}
			}

			// starts Go routine that randomly keeps making transactions to others
		case "-autotrans":
			autoTransStatus = !autoTransStatus
			go autoTrans()
		case "-id":
			fmt.Printf("Your ID is: \n    Public Key short: %v\n    Port: %v\n", publicKey.Hash()[:6], *port)
		case "-verbose":
			consensus.SwitchVerbose()
		default:
			fmt.Println(commandline, "is not a known command. Type -h for help")
		}
	}
}

func prettyPrintHelpMessage(command string, explain string) {
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

func autoTrans() {

	for {
		// checks if autotrans should stop
		if !autoTransStatus {
			return
		}
		currentStake := transaction.GetCurrentLedger()[publicKey.Hash()]
		if currentStake == 0 {
			continue
		}

		pkList := p2p.GetPublicKeys()
		for i := 0; i < 2; i++ {
			receiverPK := pkList[rand.Intn(len(pkList))]
			trans := objects.CreateTransaction(publicKey,
				receiverPK,
				uint64(rand.Intn(int(currentStake)/50)),
				strconv.Itoa(i),
				secretKey)
			channels.TransClientInput <- objects.TransData{Transaction: trans}
		}
		time.Sleep(time.Second * 5)
	}
}
