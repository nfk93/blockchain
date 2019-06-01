package p2p

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"sync"
	"time"
)

const (
	RPC_REQUEST_NETWORK_LIST string = "RPCHandler.RequestNetworkList"
	RPC_NEW_CONNECTION       string = "RPCHandler.NewConnection"
	RPC_SEND_BLOCK           string = "RPCHandler.SendBlock"
	RPC_SEND_TRANSDATA       string = "RPCHandler.SendTransData"

	NUMBER_OF_PEERS int = 5
)

// TODO use stringSet for networklist aswell
var networkList map[string]bool
var nLock sync.RWMutex
var peers []rpc.Client
var peersLock sync.RWMutex
var blocksSeen stringSet
var transSeen stringSet
var myHostPort string
var myIp string
var deliverBlock chan objects.Block
var deliverTrans chan objects.TransData
var inputBlock chan objects.Block
var inputTrans chan objects.TransData
var myKey crypto.PublicKey
var publicKeys map[crypto.PublicKey]bool
var pkLock sync.RWMutex

type stringSet struct {
	m map[string]bool
	l sync.RWMutex
}

func newStringSet() *stringSet {
	result := new(stringSet)
	result.m = make(map[string]bool)
	return result
}

func (b *stringSet) add(s string) {
	b.m[s] = true
}

func (b *stringSet) lock() {
	b.l.Lock()
}

func (b *stringSet) unlock() {
	b.l.Unlock()
}

func (b *stringSet) rlock() {
	b.l.RLock()
}

func (b *stringSet) runlock() {
	b.l.RUnlock()
}

func (b *stringSet) contains(s string) bool {
	return b.m[s]
}

func StartP2P(connectTo string, hostPort string, mypk crypto.PublicKey, channels objects.ChannelStruct) {
	networkList = make(map[string]bool)
	blocksSeen = *newStringSet()
	transSeen = *newStringSet()
	myIp = getIP()
	myHostPort = hostPort
	deliverBlock = channels.BlockFromP2P
	deliverTrans = channels.TransFromP2P
	inputBlock = channels.BlockToP2P
	inputTrans = channels.TransClientInput
	myKey = mypk
	publicKeys = make(map[crypto.PublicKey]bool)

	if connectTo == "" {
		fmt.Println("STARTING OWN NETWORK!")
		networkList[myIp+":"+myHostPort] = true
		publicKeys[myKey] = true
		determinePeers()
		go listenForRPC(myHostPort)
		fmt.Printf("Listening on %v:%v \n", myIp, myHostPort)

	} else {
		fmt.Println("CONNECTING TO EXISTING NETWORK AT ", connectTo)
		go listenForRPC(myHostPort)
		time.Sleep(1 * time.Second)
		connectToNetwork(connectTo)
	}

	// Pull user-input transactions and send via p2p
	go func() {
		for {
			trans := <-inputTrans
			go handleTransData(trans)
		}
	}()
	// Send blocks coming from the Consensus layer via p2p
	go func() {
		for {
			block := <-inputBlock
			go handleBlock(block)
		}
	}()
}

func PrintNetworkList() {
	nLock.RLock()
	defer nLock.RUnlock()
	for _, k := range setAsList(networkList) {
		fmt.Println(k)
	}
}

func PrintTransHashList() {
	transSeen.rlock()
	defer transSeen.runlock()
	for _, k := range setAsList(transSeen.m) {
		fmt.Println(k)
	}
}

func GetPublicKeys() []crypto.PublicKey {
	var pkList []crypto.PublicKey
	for pk := range publicKeys {
		pkList = append(pkList, pk)
	}
	return pkList
}

func PrintPublicKeys() {
	pkLock.RLock()
	defer pkLock.RUnlock()
	list := make([]crypto.PublicKey, 0, len(publicKeys))
	for k := range publicKeys {
		if publicKeys[k] == true {
			list = append(list, k)
		}
	}
	for _, k := range list {
		fmt.Println(k.String())
	}
}

func PrintPeers() {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, v := range peers {
		fmt.Println(v)
	}
}

func listenForRPC(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("error starting network listener: ", err.Error())
		log.Fatal("error starting network listener: ", err.Error())
	}
	rpcObj := new(RPCHandler)
	rpc.HandleHTTP()
	err = rpc.Register(rpcObj)
	if err != nil {
		log.Fatal("RPCHandler can't be registered, ", err)
	}
	er := http.Serve(ln, nil)
	if er != nil {
		log.Fatal("Error serving: ", err)
	}
}

// -----------------------------------------------------------
// NETWORK METHODS
// -----------------------------------------------------------
type RPCHandler int

type RequestNetworkListReply struct {
	NetworkList map[string]bool
	PublicKeys  map[crypto.PublicKey]bool
}

func (r *RPCHandler) RequestNetworkList(_ struct{}, reply *RequestNetworkListReply) error {
	nLock.RLock()
	defer nLock.RUnlock()
	*reply = RequestNetworkListReply{networkList, publicKeys}
	return nil
}

type NewConnectionData struct {
	NewAddr string
	Pk      crypto.PublicKey
}

func (r *RPCHandler) NewConnection(data NewConnectionData, _ *struct{}) error {
	// Check if we know the peer, and exit early if we do.
	alreadyKnown := false
	func() {
		nLock.RLock()
		defer nLock.RUnlock()
		if networkList[data.NewAddr] {
			alreadyKnown = true
		}
	}()
	if alreadyKnown {
		// Early exit
		return nil
	}

	nLock.Lock()
	defer nLock.Unlock()
	// We must check list again, because we can't upgrade locks (in GOs default rwlock implementation)
	if networkList[data.NewAddr] != true {
		networkList[data.NewAddr] = true
		determinePeers()

		go broadcastNewConnection(data)
		pkLock.Lock()
		defer pkLock.Unlock()
		publicKeys[data.Pk] = true
	}
	return nil
}

func broadcastNewConnection(data NewConnectionData) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range peers {
		void := struct{}{}
		err := peer.Call(RPC_NEW_CONNECTION, data, &void)
		if err != nil {
			peersLock.RUnlock()
			determinePeers()
			broadcastNewConnection(data)
			break
		}
	}
}

func (r *RPCHandler) SendBlock(block objects.Block, _ *struct{}) error {
	// Check if we know the peer, and exit early if we do.
	alreadyKnown := false
	func() {
		blocksSeen.rlock()
		defer blocksSeen.runlock()
		alreadyKnown = blocksSeen.contains(block.CalculateBlockHash())
	}()
	if alreadyKnown {
		// Early exit
		return nil
	}

	handleBlock(block)
	return nil
}

func handleBlock(block objects.Block) {

	blocksSeen.lock()
	defer blocksSeen.unlock()
	// We must check list again, because we can't upgrade locks (in GOs default rwlock implementation)
	if blocksSeen.contains(block.CalculateBlockHash()) != true {
		blocksSeen.add(block.CalculateBlockHash())

		go func() { deliverBlock <- block }()
		go broadcastBlock(block)
	}
}

func broadcastBlock(block objects.Block) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range peers {
		void := struct{}{}
		err := peer.Call(RPC_SEND_BLOCK, block, &void)
		if err != nil {
			peersLock.RUnlock()
			determinePeers()
			broadcastBlock(block)
			break
		}
	}
}

func (r *RPCHandler) SendTransData(trans objects.TransData, _ *struct{}) error {
	// Check if we know the peer, and exit early if we do.
	alreadyKnown := false
	func() {
		transSeen.rlock()
		defer transSeen.runlock()
		alreadyKnown = transSeen.contains(trans.Hash())
	}()
	if alreadyKnown {
		// Early exit
		return nil
	}

	handleTransData(trans)
	return nil
}

func handleTransData(trans objects.TransData) {
	transSeen.lock()
	defer transSeen.unlock()
	// We must check list again, because we can't upgrade locks (in GOs default rwlock implementation)
	transHash := trans.Hash()
	if transSeen.contains(transHash) != true {
		transSeen.add(transHash)

		go func() { deliverTrans <- trans }()
		go broadcastTrans(trans)
	}
}

func broadcastTrans(trans objects.TransData) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range peers {
		void := struct{}{}
		err := peer.Call(RPC_SEND_TRANSDATA, trans, &void)
		if err != nil {
			peersLock.RUnlock()
			determinePeers()
			broadcastTrans(trans)
			break
		}
	}
}

func connectToNetwork(addr string) {
	client, err := rpc.DialHTTP("tcp", addr)
	defer client.Close()

	if err != nil {
		log.Fatal(err)
	} else {
		var reply RequestNetworkListReply
		e := client.Call(RPC_REQUEST_NETWORK_LIST, struct{}{}, &reply)
		if e != nil {
			log.Fatal(e)
		}
		networkList = reply.NetworkList
		publicKeys = reply.PublicKeys
		networkList[myIp+":"+myHostPort] = true
		publicKeys[myKey] = true
		determinePeers()
		broadcastNewConnection(NewConnectionData{myIp + ":" + myHostPort, myKey})
	}
}

// -----------------------------------------------------------
// INTERNAL METHODS
// -----------------------------------------------------------

func determinePeers() {
	// determines your peers. Is run every time you receive a new connection
	peersLock.Lock()
	defer peersLock.Unlock()
	connections := setAsList(networkList)
	sort.Strings(connections)
	networkSize := len(connections)
	peersSize := min(networkSize-1, NUMBER_OF_PEERS)
	myIndex, err := indexOf(myIp+":"+myHostPort, connections)
	if err != nil {
		log.Fatal("FATAL ERROR, determinePeers: ", err)
	}
	peers = make([]rpc.Client, peersSize)
	j := 0
	for i := 0; i < peersSize; i++ {
		nextIndex := (myIndex + i + j + 1) % networkSize
		notConnected := true
		for notConnected && j < 20 {
			peerClient, err := rpc.DialHTTP("tcp", connections[nextIndex])
			if err != nil {
				j++
				fmt.Println(fmt.Sprintf("Cant connect to peer %s, error: %s\nTrying next peer... Failed attempts: %d",
					connections[nextIndex], err.Error(), j))
			} else {
				peers[i] = *peerClient
				notConnected = false
			}
		}
		if j >= 20 {
			log.Fatal("Too many failed peer connection attempts, exiting")
		}
	}
}

// This is a slightly hacky way to obtain your own IPv4 IP address
func getIP() string {
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("my ip is:", string(ip))
	return string("192.168.87.120")
}
