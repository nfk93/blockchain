package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/nfk93/blockchain/consensus"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/p2p"
	"github.com/nfk93/blockchain/transaction"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"
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
	consensus.StartConsensus(channels, publicKey, secretKey, false, *saveLogFile)
	if *addr == "" {
		fmt.Println("When all other clients are ready, use -start to begin the Blockchain protocol or -h for help with further commands!")
	} else {
		fmt.Println("This client is ready for the Blockchain protocol to start. Use -h for further commands!")
	}
	cliLoop()
}

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("login"),
	readline.PcItem("say",
		readline.PcItemDynamic(listFiles("./"),
			readline.PcItem("with",
				readline.PcItem("following"),
				readline.PcItem("items"),
			),
		),
		readline.PcItem("hello"),
		readline.PcItem("bye"),
	),
	readline.PcItem("setprompt"),
	readline.PcItem("setpassword"),
	readline.PcItem("bye"),
	readline.PcItem("help"),
	readline.PcItem("go",
		readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
		readline.PcItem("install",
			readline.PcItem("-v"),
			readline.PcItem("-vv"),
			readline.PcItem("-vvv"),
		),
		readline.PcItem("test"),
	),
	readline.PcItem("sleep"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

func cliLoop() {

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case line == "-h":
			prettyPrintHelpMessage("-q", "Exit this program")
			//		prettyPrintHelpMessage("-start", "Begins the blockchain protocol")
			//		prettyPrintHelpMessage("-verbose", "Initially active. Switches verbose mode between on and off")
			//		prettyPrintHelpMessage("-id", "Show your public key in short and which port you are listening on")
			//		prettyPrintHelpMessage("-n", "Print out the network list of who you are connected to")
			//		prettyPrintHelpMessage("-trans", "Print list of seen transactions")
			//		prettyPrintHelpMessage("-peers", "Print list of all peers in the network")
			//		prettyPrintHelpMessage("-public-keys", "Print list of know Public keys in network")
			//		prettyPrintHelpMessage("-ledger", "Print the current ledger")
			//		prettyPrintHelpMessage("-final", "Print the last finalized ledger")
			//		prettyPrintHelpMessage("-trans1000", "Transfer 1000k to all known public keys in the network")
			//		prettyPrintHelpMessage("-trans2", "Transfer an even share of your current stake to everyone in the network including yourself")
			//		prettyPrintHelpMessage("-trans5", "Transfer 5 random amounts to random accounts")
			//		prettyPrintHelpMessage("-autotrans", "Initially active. Starts or stops automatic transfers to random accounts every 5 second")

		case line == "help":
			usage(l.Stderr())

		case line == "-q":
			return
		case line == "-n":
			p2p.PrintNetworkList()
		case line == "-trans":
			p2p.PrintTransHashList()
		case line == "-peers":
			p2p.PrintPeers()
		case line == "-public-keys":
			p2p.PrintPublicKeys()
		case line == "-ledger":
			transaction.PrintCurrentLedger()
		case line == "-final":
			consensus.PrintCurrentStake()
		case line == "-start": //"-start_network":
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
				log.Println("Only the network founder can start the network!")
			}
		case line == "-trans1000":
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
		case line == "-trans2":
			/*currentStake := consensus.GetLastFinalState()[publicKey.Hash()]
			for i, p := range p2p.GetPublicKeys() {

				trans := objects.CreateTransaction(publicKey,
					p,
					currentStake/uint64(len(p2p.GetPublicKeys())),
					strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- objects.TransData{Transaction: trans}

			} */
		case line == "-trans5":
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
		case line == "-autotrans":
			autoTransStatus = !autoTransStatus
			if autoTransStatus {
				println("AutoTrans is now on")
			} else {
				println("AutoTrans is now off")
			}
			go autoTrans()
		case line == "-id":
			log.Printf("Your ID is: \n    Public Key hash: %v\n    Port: %v\n", publicKey.Hash(), *port)
		case line == "-verbose":
			consensus.SwitchVerbose()

		case strings.HasPrefix(line, "transaction "):
			params := strings.Fields(line[12:])
			noOfParams := 4

			var amount uint64
			var receiver string

			if len(params) == noOfParams {
				for i := 0; i < noOfParams; i += 2 {
					if params[i] == "-amount" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as transaction amount")
						}
						amount = am
					} else if params[i] == "-receiver" {
						receiver = params[i+1]
					}
				}
				if amount > 0 && receiver != "" {
					newTrans := objects.CreateTransaction(publicKey, publicKey, amount, "", secretKey)
					log.Printf("Transaction with id: %v has been created!", newTrans.ID)
					channels.TransClientInput <- objects.TransData{Transaction: newTrans}
					goto exit

				}
			}

			log.Println("Bad input! Use syntax: transaction -amount [uint] -receiver [string]")

		case strings.HasPrefix(line, "call "):
			params := strings.Fields(line[5:])
			noOfParams := 12
			var call string
			var entry string
			var callParams string
			var amount uint64
			var gas uint64
			var conAddr string

			if len(params) == noOfParams {
				for i := 0; i < noOfParams; i += 2 {
					if params[i] == "-call" {
						call = params[i+1]
					} else if params[i] == "-entry" {
						entry = params[i+1]
					} else if params[i] == "-callParams" {
						callParams = params[i+1]
					} else if params[i] == "-amount" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as contract amount")
						}
						amount = am
					} else if params[i] == "-gas" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as contract gas")
						}
						gas = am
					} else if params[i] == "-conAddr" {
						conAddr = params[i+1]
					}
				}
				if call != "" && entry != "" && callParams != "" && amount > 0 && gas > 0 && conAddr != "" {
					conCall := objects.CreateContractCall(call, entry, callParams, amount, gas, conAddr, publicKey, secretKey)
					log.Printf("Contract Call to %v has been created!", conCall.Address)
					channels.TransClientInput <- objects.TransData{ContractCall: conCall}
					goto exit

				}
			}

			log.Println("Bad input! Use syntax: call  init -call [string] -entry [string] -callParams [string] -amount [uint] -gas [uint] -conAddr [string]")

		case strings.HasPrefix(line, "init "):
			params := strings.Fields(line[5:])
			noOfParams := 8
			var code []byte
			var gas uint64
			var prepaid uint64
			var storageLimit uint64

			if len(params) == noOfParams {
				fmt.Println(params)
				for i := 0; i < noOfParams; i += 2 {
					if params[i] == "-code" {
						code = readFromFile(params[i+1])
						fmt.Println(string(code))
						goto exit
					} else if params[i] == "-gas" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as gas amount")
						}
						gas = am
					} else if params[i] == "-prepaid" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as prepaid amount")
						}
						prepaid = am
					} else if params[i] == "-storageLimit" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as Storage limit")
						}
						storageLimit = am
					}
				}

				if code != nil && gas > 0 && prepaid > 0 && storageLimit > 0 {
					conInit := objects.CreateContractInit(publicKey, code, gas, prepaid, storageLimit, secretKey)
					log.Println("The Contract init has been created!")
					channels.TransClientInput <- objects.TransData{ContractInit: conInit}
					goto exit
				}
			} else {
				log.Printf("Not %v parameters supplied!", noOfParams)
			}

			log.Println("Bad input! Use syntax: init -code [Path] -gas [uint] -prepaid [uint] -storageLimit [uint]")

		default:
			println(line, "is not a known command. Type -h for help")

		}
	exit:
	}
}
func readFromFile(s string) []byte {
	bytemsg, err := ioutil.ReadFile(s)
	if err != nil {
		panic(err)
	}
	return bytemsg
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

	log.Println(buf.String())
}

func autoTrans() {

	for {
		// checks if autotrans should stop
		if !autoTransStatus {
			return
		}
		currentStake := transaction.GetCurrentLedger()[publicKey.Hash()]
		if int(currentStake)/50 <= 0 {
			time.Sleep(time.Second * 5)
		} else {
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
}
