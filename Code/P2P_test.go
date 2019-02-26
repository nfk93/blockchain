package Code

import (
	"testing"
)

// Method for failing a test and reporting a meaningful error
func fail(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Error("FAILURE: expected: ", expected, ", actual: ", actual)
	}
}

func TestSendingBlocks(t *testing.T) {
	go serveOnPort("8080")
	payload := make([]byte, 10)
	payload[5] = 1
	broadcastBlock("localhost:8080", Block{123})
}

func TestKeyset(t *testing.T) {
	var m map[string]bool = make(map[string]bool)
	m["hey"] = true
	m["yo"] = true
	m["hey"] = false

	keys := keyset(m)

	fail(t, 1, len(keys))
}
