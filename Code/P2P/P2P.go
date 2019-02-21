package P2P

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	fmt.Println("Hello, World!")
}

type RPC int
type void struct{}

// Dummy methods for testing the RPC calls. Use as examples for how to
// make RPC methods
func sendGreeting(addr net.Addr, msg string) string {
	_, myPort, _ := net.SplitHostPort(addr.String())
	// Hardcoded localhost for testing purposes. Normally you just give
	// addr.String() as second parameter for some Net.Addr object
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:"+myPort)
	var response string
	if err != nil {
		log.Fatal("DialHTTP failed: ", err)
	} else {
		errRPC := client.Call("RPC.Greet", msg, &response)
		if errRPC != nil {
			log.Fatal("Client.Call failed: ", errRPC)
		}
	}
	return response
}
func (r *RPC) Greet(msg string, returnAddress *string) error {
	if msg != "Hello, m'lady" {
		*returnAddress = "That's no way to greet a lady!"
	} else {
		*returnAddress = "Hello, sir"
	}
	return nil
}

// Call this method to register all your RPC methods and listen for calls
// to them
func listenForRPC() (myAddress net.Addr) {
	ln, _ := net.Listen("tcp", ":")
	myAddress = ln.Addr()

	rpc_ := new(RPC)
	err := rpc.Register(rpc_)
	if err != nil {
		log.Fatal("RPC registration failed: ", err)
	}
	rpc.HandleHTTP()
	go http.Serve(ln, nil)
	return
}

/*This method should distribute a transaction/smartcontract on the P2P network */
func SendTransaction() {
	/* TODO */
}

// This method should be called from the consensus layer to distribute a Block on the network
func SendBlock() {
	/* TODO */
}

//This method should handle user input to allow transactions to be sent
func handleUserInput() {
	/* TODO */
}

//This method should deliver a received transaction or block to the consensus layer
func received() {
	/*TODO */
}
