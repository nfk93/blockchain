package P2P

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

func main() {
	fmt.Println("Hello, World!")
}

type void struct{}

type P2PLayer struct {
	p2PRPC      P2P
	rpcListener net.Listener
}

type P2P struct {
	myAddr             net.Addr
	Networklist        map[net.Addr]bool
	networkLock        sync.RWMutex
	connectedPeers     map[net.Addr]bool
	connectedPeersLock sync.RWMutex
}

type P2PError struct {
	Msg string
}

func (e *P2PError) Error() string {
	return e.Msg
}

func NewP2P(connect_to net.Addr) *P2PLayer {
	p2p := new(P2PLayer)
	p2p.rpcListener = p2p.listenForRPC()
	_, rpc_port, err := net.SplitHostPort(p2p.rpcListener.Addr().String())
	if err != nil {
		log.Fatal("Error determining port for RPC: ", err)
	}
	fmt.Println(rpc_port)
	return p2p
}

// Call this method to register all your RPC methods and listen for calls
// to them
func (p *P2PLayer) listenForRPC() (rpcListener net.Listener) {
	rpcListener, _ = net.Listen("tcp", ":")
	rpcListener.Addr()

	err := rpc.Register(p)
	if err != nil {
		log.Fatal("RPC registration failed: ", err)
	}
	rpc.HandleHTTP()
	go http.Serve(rpcListener, nil)
	return
}

/*This method should distribute a transaction/smartcontract on the P2P network */
func (p *P2PLayer) SendTransaction() {
	/* TODO */
}

// This method should be called from the consensus layer to distribute a Block on the network
func (p *P2PLayer) SendBlock() {
	/* TODO */
}

//This method should handle user input to allow transactions to be sent
func (p *P2PLayer) handleUserInput() {
	/* TODO */
}

//This method should deliver a received transaction or block to the consensus layer
func (p *P2PLayer) received() {
	/*TODO */
}

func (p *P2P) broadcastExistence(myAddr net.Addr, _ void) error {
	p.networkLock.Lock()
	defer p.networkLock.Unlock()
	p.connectedPeersLock.Lock()
	defer p.connectedPeersLock.Unlock()

	p.Networklist[myAddr] = true
	p.connectedPeers = determineConnectedPeers(p.Networklist, myAddr)

	return nil
}

// This is a slightly hacky way to obtain your own IPv4 IP address
func getIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		log.Fatal("Can't determine own IP, check internet connection... ", err)
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr).IP.String()
	return addr
}

func makeAddrObject(network string, ip string, port string) net.Addr {
	var addr net.Addr
	addr.String() = ip + ":" + port
	addr.Network() = network
	return addr
}

func determineConnectedPeers(network map[net.Addr]bool, myAddr net.Addr) map[net.Addr]bool {
	networkSize := len(network)
	keys := make([]string, 0, networkSize)
	for k, _ := range network {
		keys = append(keys, k.String())
	}
	resultSize := min(networkSize-1, 10)
	myIndex := indexOf(myAddr.String(), keys)
	result := make(map[net.Addr]bool, resultSize)
	for i := 0; i < resultSize; i++ {
		host, port, _ := net.SplitHostPort(keys[(myIndex+i+1)%networkSize])
		netaddrobj := makeAddrObject("tcp", host, port)
		result[netaddrobj] = true
	}
	return result
}

func makeMap(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func indexOf(element string, slice []string) int {
	for i, v := range slice {
		if element == v {
			return i
		}
	}
	return -1
}

// helper function to get a simple minimum function that returns an integer
func min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
