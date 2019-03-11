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
)

var networkList map[string]bool
var nLock sync.RWMutex
var peers []string
var peersLock sync.RWMutex
var myHostPort string
var myIp string
var deliverBlock chan objects.Block
var deliverTrans chan objects.Transaction

func StartP2P(connectTo string, hostPort string) {
	networkList = make(map[string]bool)
	myIp = getIP().String()
	myHostPort = hostPort

	if connectTo == "" {
		fmt.Println("STARTING OWN NETWORK!")
		networkList[myIp+":"+myHostPort] = true
		determinePeers()
		go listenForRPC(myHostPort)

	} else {
		fmt.Println("CONNECTING TO EXISTING NETWORK AT ", connectTo)
		connectToNetwork(connectTo)
		go listenForRPC(myHostPort)
	}
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
	er := http.Serve(ln, nil)
	if er != nil {
		log.Fatal("Error serving: ", err)
	}
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

func (r *RPCHandler) NewConnection(newAddr string, reply *struct{}) error {
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
			fmt.Println("ERROR broadcastNewConnection: can't broadcast new connection to " + peer)
		} else {
			void := struct{}{}
			client.Call(RPC_NEW_CONNECTION, newAddr, &void)
		}
	}
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
	// TODO
	peersLock.Lock()
	defer peersLock.Unlock()
	connections := setAsList(networkList)
	sort.Strings(connections)
	networkSize := len(connections)
	peersSize := min(networkSize, 10)
	myIndex, err := indexOf(myIp+":"+myHostPort, connections)
	if err != nil {
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
