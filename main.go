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
	"github.com/nfk93/blockchain/smart"
	"github.com/nfk93/blockchain/transaction"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
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
var finalizeGap *uint64
var epochLength *uint64
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
	finalizeGap = flag.Uint64("finalize_gap", 100, "Specify the finalization gap (default=50")
	epochLength = flag.Uint64("epoch_length", 20, "Specify the epoch length (default=50")
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
		fmt.Println("When all other clients are ready, use start to begin the Blockchain protocol or -h or --help for help with further commands!")
	} else {
		fmt.Println("This client is ready for the Blockchain protocol to start. Use -h or --help for further commands!")
	}
	cliLoop()
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func cliLoop() {

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
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
			printHelpMenu()
		case line == "--help":
			printHelpMenu()
		case line == "exit":
			return
		case line == "network":
			p2p.PrintNetworkList()
		case line == "seenTrans":
			p2p.PrintTransHashList()
		case line == "peers":
			p2p.PrintPeers()
		case line == "publicKeys":
			p2p.PrintPublicKeys()
		case line == "ledger":
			ledger := transaction.GetCurrentLedger()
			var keyList []string
			for k := range ledger {
				keyList = append(keyList, k)
			}
			sort.Strings(keyList)

			for _, k := range keyList {
				log.Printf("Amount %v is owned by %v\n", ledger[k], k[:10])
			}

		case line == "contracts":
			contractMap := smart.GetContracts()
			var keyList []string
			for k := range contractMap {
				keyList = append(keyList, k)
			}
			sort.Strings(keyList)

			for _, k := range keyList {
				log.Printf("Contract %v was created in slot %v.\n", k, contractMap[k].CreatedAtSlot)
			}

		case strings.HasPrefix(line, "contractInfo "):
			params := strings.Fields(line[13:])
			conAddr := params[0]
			conState := smart.GetContractState(conAddr)

			log.Printf(" Contract: %v \n Balance: %v \n Prepaid: %v \n Storage Limit: %v\n Storage: %v \n",
				conAddr, conState.Balance, conState.PrepaidStorage, conState.Storagecap, conState.Storage)

			if len(params) == 3 && params[1] == "-o" { // with output file
				var targetFileName string
				targetFileName = params[2]
				contract := smart.GetContract(conAddr)

				func() {
					f, err := os.Create(targetFileName)
					if err != nil {
						log.Println(err)
						return
					}
					_, err = f.WriteString(contract.Code)
					if err != nil {
						log.Println(err)
					}
					err = f.Close()
					if err != nil {
						log.Println(err)
					}
				}()

			}

		case line == "final":
			consensus.PrintCurrentStake()
		case line == "start": //"-start_network":
			if *newNetwork {
				genesisdata, err := objects.NewGenesisData(publicKey, time.Second*time.Duration(*slotduration), *hardness, *finalizeGap, *epochLength)
				if err != nil {
					log.Println(err)
					goto exit
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

		case line == "id":
			log.Printf("Your ID is: \n    Short Public Key hash: %v\n    Full Public Key hash: %v\n    Port: %v\n", publicKey.Hash()[:10], publicKey.Hash(), *port)
		case line == "verbose":
			consensus.SwitchVerbose()

		case strings.HasPrefix(line, "transaction "):
			params := strings.Fields(line[12:])
			noOfParams := 2

			var amount uint64
			var receiver string

			if len(params) == noOfParams {
				receiver = params[0]
				amountUint, err := strconv.ParseUint(params[1], 10, 64)
				if err != nil {
					log.Println("Bad number as transaction amount")
					goto exit
				}
				amount = amountUint

				if amount > 0 && receiver != "" {
					recPk, success := getPK(receiver)
					if !success {
						log.Printf("Public Key for %v did not exist", receiver)
						goto exit
					}
					newTrans := objects.CreateTransaction(publicKey, recPk, amount, "", secretKey)
					log.Printf("Transaction with id: %v has been created!", newTrans.ID)
					channels.TransClientInput <- objects.TransData{Transaction: newTrans}
					goto exit

				}
			}

			log.Println("Bad input! Use -h or --help for help menu!")

		case strings.HasPrefix(line, "call "):
			params := strings.Fields(line[5:])
			noOfParams := 2
			entry := "main"     //default
			callParams := "()"  //default
			amount := uint64(0) //default
			var gas uint64
			var conAddr string

			if len(params) >= noOfParams {
				conAddr = params[0]
				gasUint, err := strconv.ParseUint(params[1], 10, 64)
				if err != nil {
					log.Println("Bad number as contract gas")
					goto exit
				}
				gas = gasUint

				//Flags
				for i := noOfParams; i < len(params); i += 2 {
					if params[i] == "-entry" {
						entry = params[i+1]
					} else if params[i] == "-params" {
						callParams = params[i+1]
					} else if params[i] == "-amount" {
						am, err := strconv.ParseUint(params[i+1], 10, 64)
						if err != nil {
							log.Println("Bad number as contract amount")
							goto exit
						}
						amount = am
					}
				}
				if gas > 0 && conAddr != "" {
					conCall := objects.CreateContractCall("CALL", entry, callParams, amount, gas, conAddr, publicKey, secretKey)
					log.Printf("Contract Call to %v has been created!", conCall.Address)
					channels.TransClientInput <- objects.TransData{ContractCall: conCall}
					goto exit

				}
			}

			log.Println("Bad input! Use -h or --help for help menu!")

		case strings.HasPrefix(line, "init "):
			params := strings.Fields(line[5:])
			noOfParams := 4
			var code []byte
			var gas uint64
			var prepaid uint64
			var storageLimit uint64

			if len(params) == noOfParams {
				code, err = readFromFile(params[0])
				if err != nil {
					log.Println("Error in reading file: " + err.Error())
					goto exit
				}

				gasUint, err := strconv.ParseUint(params[1], 10, 64)
				if err != nil {
					log.Println("Bad number as gas amount")
					goto exit
				}
				gas = gasUint
				prepaidUint, err := strconv.ParseUint(params[2], 10, 64)
				if err != nil {
					log.Println("Bad number as prepaid amount")
					goto exit
				}
				prepaid = prepaidUint

				storageUint, err := strconv.ParseUint(params[3], 10, 64)
				if err != nil {
					log.Println("Bad number as Storage limit")
					goto exit
				}
				storageLimit = storageUint

				if code != nil && gas > 0 && prepaid > 0 && storageLimit > 0 {
					conInit := objects.CreateContractInit(publicKey, code, gas, prepaid, storageLimit, secretKey)
					log.Println("The Contract init has been created!")
					channels.TransClientInput <- objects.TransData{ContractInit: conInit}
					goto exit
				}
			} else {
				log.Printf("Not %v parameters supplied!", noOfParams)
			}

			log.Println("Bad input! Use -h or --help for help menu!")

		default:
			println(line, "is not a known command. Type -h or --help for help menu!")

		}
	exit:
		/* //Test Code
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
			currentStake := consensus.GetLastFinalState()[publicKey.Hash()]
			for i, p := range p2p.GetPublicKeys() {

				trans := objects.CreateTransaction(publicKey,
					p,
					currentStake/uint64(len(p2p.GetPublicKeys())),
					strconv.Itoa(i),
					secretKey)
				channels.TransClientInput <- objects.TransData{Transaction: trans}

			}
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
			go autoTrans() */

	}
}
func printHelpMenu() {
	println("Command", []string{"Description"})
	prettyPrintHelpMessage("exit", []string{"Exit this program"})
	prettyPrintHelpMessage("start", []string{"Begins the blockchain protocol"})
	prettyPrintHelpMessage("verbose", []string{"Initially active. Switches verbose mode between on and off"})
	prettyPrintHelpMessage("id", []string{"Show your public key in short and which port you are listening on"})
	prettyPrintHelpMessage("networkList", []string{"Print out the network list of who you are connected to"})
	prettyPrintHelpMessage("seenTrans", []string{"Print list of seen transactions"})
	prettyPrintHelpMessage("peers", []string{"Print list of all peers in the network"})
	prettyPrintHelpMessage("publicKeys", []string{"Print list of know Public keys in network"})
	prettyPrintHelpMessage("ledger", []string{"Print the current ledger"})
	prettyPrintHelpMessage("final", []string{"Print the last finalized ledger"})
	prettyPrintHelpMessage("contracts", []string{"Prints a list of all currently active contracts"})
	prettyPrintHelpMessage("contractInfo ADDRESS", []string{"Prints info of a given contract",
		"", "ADDRESS: The address of a given contract",
		"", "",
		"-o <string>", "Path with filename of output file. Path should be without file extension.",
	})
	prettyPrintHelpMessage("transaction RECEIVER AMOUNT", []string{"Send Amount to the Receiver",
		"", "RECEIVER: A 10 digit prefix of senders publicKey hash",
		"", "AMOUNT: Positive integer of amount to transfer"})
	prettyPrintHelpMessage("call ADDRESS GAS ", []string{"Makes a contract call to the specified contract",
		"", "ADDRESS: The address of the contract",
		"", "GAS: Positive integer of how much gas to include",
		"", "",
		"-entry <string>", "Default: main. Used to specify the entry in the contract to call",
		"-amount <uint>", "Default: 0. Non negative integer of amount included in a contract call",
		"-params <string>", "Default: \"()\". Used to specify parameters to include in contract call "})
	prettyPrintHelpMessage("init CODE GAS PREPAID STORAGE", []string{"",
		"", "CODE: path to code file",
		"", "GAS: Positive integer of how much gas to include",
		"", "PREPAID: Positive integer of how much prepaid money to attached at contract",
		"", "STORAGE: Positive integer of max storage usage for a contract"})
}
func readFromFile(s string) ([]byte, error) {
	bytemsg, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return bytemsg, nil
}

func prettyPrintHelpMessage(command string, explain []string) {
	var buf bytes.Buffer
	buf.WriteString("  " + command)
	for i := 0; i < 36-len(command); i++ {
		buf.WriteString(" ")
	}
	buf.WriteString(explain[0] + "\n")

	for i := 1; i < len(explain); i += 2 {
		buf.WriteString("        " + explain[i])
		for p := 0; p < 30-len(explain[i]); p++ {
			buf.WriteString(" ")
		}
		buf.WriteString(explain[i+1] + "\n")

	}

	println(buf.String())
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

func getPK(prefix string) (crypto.PublicKey, bool) {
	pkList := p2p.GetPublicKeys()
	if len(prefix) < 10 {
		log.Println("Receivers key is to short...")
	}
	for _, pk := range pkList {
		if strings.HasPrefix(pk.Hash(), prefix) {
			return pk, true
		}
	}
	return crypto.PublicKey{}, false
}
