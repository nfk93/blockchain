package P2P

import (
	"testing"
)

var rpc_ RPC

// Method for failing a test and reporting a meaningful error
func fail(t *testing.T, expected string, actual string) {
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

func TestGreetAsRPC1(t *testing.T) {
	myAddress := listenForRPC()
	response := sendGreeting(myAddress, "Hello, m'lady")

	expected := "Hello, sir"
	fail(t, expected, response)
}

func TestGreetAsRPC2(t *testing.T) {
	myAddress := listenForRPC()
	response := sendGreeting(myAddress, "yo")

	expected := "That's no way to greet a lady!"
	fail(t, expected, response)
}
