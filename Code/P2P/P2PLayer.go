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

func (p *P2P) connect(my_port string, _ void) error {
	p.networkLock.Lock()
	defer p.networkLock.Unlock()

	return nil
}

//func (p *P2P) addPeer(addr net.Addr)
