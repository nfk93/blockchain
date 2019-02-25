package P2P

import (
	"testing"
)

var rpc_ RPC

// Method for failing a test and reporting a meaningful error
func fail(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Errorf("FAILURE: expected: %s, actual: %s",
			expected, actual)
	}
}

// These tests serve as an example for how to write tests for networking code
// GoLand can't run them. Run from terminal with 'go test'

func TestGreet1(t *testing.T) {
	var response string
	err := rpc_.Greet("Hello, m'lady", &response)
	if err != nil {
		t.Fail()
	}

	expected := "Hello, sir"
	fail(t, expected, response)
}

func TestGreet2(t *testing.T) {
	var response string
	err := rpc_.Greet("Oi cunt", &response)
	if err != nil {
		t.Fail()
	}

	expected := "That's no way to greet a lady!"
	fail(t, expected, response)
}

func TestRPCListening(t *testing.T) {
	rpcListener := listenForRPC()
	defer rpcListener.Close()

	t.Run("TestGreetingRPC1", func(t *testing.T) {
		response := sendGreeting(rpcListener.Addr(), "Hello, m'lady")

		expected := "Hello, sir"
		fail(t, expected, response)
	})
	t.Run("TestGreetingRPC2", func(t *testing.T) {
		response := sendGreeting(rpcListener.Addr(), "yo")

		expected := "That's no way to greet a lady!"
		fail(t, expected, response)
	})
}

func TestMakingBlock(t *testing.T) {
	icanmakeablock()
}
