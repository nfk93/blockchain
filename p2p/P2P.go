package p2p

import (
	"fmt"
	"github.com/nfk93/blockchain/objects"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"sync"
)

const (
	RPC_REQUEST_NETWORK_LIST string = "RPCHandler.RequestNetworkList"
	RPC_NEW_CONNECTION       string = "RPCHandler.NewConnection"
	RPC_SEND_BLOCK           string = "RPCHandler.SendBlock"
	RPC_SEND_TRANSACTION     string = "RPCHandler.SendTransaction"
)

var networkList map[string]bool
var nLock sync.RWMutex
var peers []string
var peersLock sync.RWMutex
var blocksSeen stringSet
var transSeen stringSet
var myHostPort string
var myIp string
var deliverBlock chan objects.Block
var deliverTrans chan objects.Transaction
var inputBlock chan objects.Block
var inputTrans chan objects.Transaction

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

func StartP2P(connectTo string, hostPort string, blockIn chan objects.Block, blockOut chan objects.Block,
	transIn chan objects.Transaction, transOut chan objects.Transaction) {
	networkList = make(map[string]bool)
	blocksSeen = *newStringSet()
	transSeen = *newStringSet()
	myIp = getIP().String()
	myHostPort = hostPort
	deliverBlock = blockOut
	deliverTrans = transOut
	inputBlock = blockIn
	inputTrans = transIn

	if connectTo == "" {
		fmt.Println("STARTING OWN NETWORK!")
		networkList[myIp+":"+myHostPort] = true
		determinePeers()
		listenForRPC(myHostPort)

	} else {
		fmt.Println("CONNECTING TO EXISTING NETWORK AT ", connectTo)
		connectToNetwork(connectTo)
		listenForRPC(myHostPort)
	}

	// Pull user-input transactions and send via p2p
	go func() {
		for {
			trans := <-inputTrans
			go handleTransaction(trans)
		}
	}()
	// Send blocks coming from the Consensus layer via p2p, without delivering them back to consensuslayer
	go func() {
		for {
			block := <-inputBlock
			go handleBlockWithoutDelivering(block)
		}
	}()
}

func PrintNetworkList() {
	for _, k := range setAsList(networkList) {
		fmt.Println(k)
	}
}

func listenForRPC(port string) {
	ln, _ := net.Listen("tcp", ":"+port)
	rpcObj := new(RPCHandler)
	err := rpc.Register(rpcObj)
	if err != nil {
		log.Fatal("RPCHandler can't be registered, ", err)
	}
	rpc.HandleHTTP()
	go func() {
		er := http.Serve(ln, nil)
		if er != nil {
			log.Fatal("Error serving: ", err)
		}
	}()
}

// -----------------------------------------------------------
// NETWORK METHODS
// -----------------------------------------------------------
type RPCHandler int

func (r *RPCHandler) RequestNetworkList(_ struct{}, reply *map[string]bool) error {
	nLock.RLock()
	defer nLock.RUnlock()
	*reply = networkList
	return nil
}

func (r *RPCHandler) NewConnection(newAddr string, _ *struct{}) error {
	// Check if we know the peer, and exit early if we do.
	alreadyKnown := false
	func() {
		nLock.RLock()
		defer nLock.RUnlock()
		if networkList[newAddr] {
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
	if networkList[newAddr] != true {
		networkList[newAddr] = true
		determinePeers()

		go broadcastNewConnection(newAddr)
	}
	return nil
}

func broadcastNewConnection(newAddr string) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range peers {
		client, err := rpc.DialHTTP("tcp", peer)
		if err != nil {
			fmt.Println("ERROR broadcastNewConnection: can't broadcast new connection to "+peer+"\n\tError: ", err)
		} else {
			void := struct{}{}
			client.Call(RPC_NEW_CONNECTION, newAddr, &void)
		}
	}
}

func (r *RPCHandler) SendBlock(block objects.Block, _ *struct{}) error {
	// Check if we know the peer, and exit early if we do.
	alreadyKnown := false
	func() {
		blocksSeen.rlock()
		defer blocksSeen.runlock()
		alreadyKnown = blocksSeen.contains(block.BlockHash)
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
	if blocksSeen.contains(block.BlockHash) != true {
		blocksSeen.add(block.BlockHash)

		// TODO: handle the block more?
		go func() { deliverBlock <- block }()
		go broadcastBlock(block)
	}
}

func handleBlockWithoutDelivering(block objects.Block) {
	blocksSeen.lock()
	defer blocksSeen.unlock()
	// We must check list again, because we can't upgrade locks (in GOs default rwlock implementation)
	if blocksSeen.contains(block.BlockHash) != true {
		blocksSeen.add(block.BlockHash)
		go broadcastBlock(block)
	}
}

func broadcastBlock(block objects.Block) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range peers {
		client, err := rpc.DialHTTP("tcp", peer)
		if err != nil {
			fmt.Println("ERROR broadcastBlock: can't broadcast block to "+peer+"\n\tError: ", err)
		} else {
			void := struct{}{}
			client.Call(RPC_SEND_BLOCK, block, &void)
		}
	}
}

func (r *RPCHandler) SendTransaction(trans objects.Transaction, _ *struct{}) error {
	// Check if we know the peer, and exit early if we do.
	alreadyKnown := false
	func() {
		transSeen.rlock()
		defer transSeen.runlock()
		alreadyKnown = transSeen.contains(transHash(trans))
	}()
	if alreadyKnown {
		// Early exit
		return nil
	}

	handleTransaction(trans)
	return nil
}

func handleTransaction(trans objects.Transaction) {
	transSeen.lock()
	defer transSeen.unlock()
	// We must check list again, because we can't upgrade locks (in GOs default rwlock implementation)
	if transSeen.contains(transHash(trans)) != true {
		transSeen.add(transHash(trans))

		// TODO: handle the block more?
		go func() { deliverTrans <- trans }()
		go broadcastTrans(trans)
	}
}

func broadcastTrans(trans objects.Transaction) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range peers {
		client, err := rpc.DialHTTP("tcp", peer)
		if err != nil {
			fmt.Println("ERROR broadcastTrans: can't broadcast transaction to "+peer+"\n\tError: ", err)
		} else {
			void := struct{}{}
			client.Call(RPC_SEND_TRANSACTION, trans, &void)
		}
	}
}

func transHash(t objects.Transaction) string {
	return t.From.String() + t.ID
}

func connectToNetwork(addr string) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		// TODO: handle error
		log.Fatal(err)
	} else {
		var reply map[string]bool
		client.Call(RPC_REQUEST_NETWORK_LIST, struct{}{}, &reply)
		networkList = reply
		networkList[myIp+":"+myHostPort] = true
		determinePeers()
		broadcastNewConnection(myIp + ":" + myHostPort)
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
	peersSize := min(networkSize, 10) //TODO use dynamic parameter rather than 10
	myIndex, err := indexOf(myIp+":"+myHostPort, connections)
	if err != nil {
		// TODO: handle gracefully
		log.Fatal("FATAL ERROR, determinePeers: ", err)
	}
	peers = make([]string, peersSize)
	for i := 0; i < peersSize; i++ {
		peers[i] = connections[(myIndex+i)%networkSize]
	}
}

// This is a slightly hacky way to obtain your own IPv4 IP address
func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr).IP
	return addr
}
