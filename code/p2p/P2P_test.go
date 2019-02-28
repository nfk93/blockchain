package p2p

import (
	"testing"
	"time"
)

// Method for failing a test and reporting a meaningful error
func fail(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Error("FAILURE: expected: ", expected, ", actual: ", actual)
	}
}

/*func TestSendingBlocks(t *testing.T) {
	ln, err := serveOnPort("8080")
	defer ln.Close()
	if err != nil {
		t.Error("Can't serve on port 8080")
	}
	payload := make([]byte, 10)
	payload[5] = 1
	broadcastBlock("localhost:8080", Block{123})


}*/

func TestSendNewConnectionTo(t *testing.T) {
	initNetworkList()
	ln, err := serveOnPort("8080")
	defer ln.Close()
	if err != nil {
		t.Error("Can't serve on port 8080")
	}

	if len(networkList) != 0 {
		t.Error("Networklist is nonempty on startup")
	}
	newAddr := "127.0.0.1:8081"
	sendNewConnectionTo("127.0.0.1:8080", newConnectionData{newAddr})

	for t := 0; t < 20; t++ {
		if networkList[newAddr] == true {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if len(networkList) != 1 {
		t.Error("Networklist has wrong amount of keys. Should be 1, but is ", len(networkList))
	} else if networkList[newAddr] != true {
		t.Error("New address was not added to networklist")
	}
}

func initNetworkList() {
	networkList = make(map[string]bool)
}
