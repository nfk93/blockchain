package p2p

import (
	"testing"
)

// Method for failing a test and reporting a meaningful error
func fail(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Error("FAILURE: expected: ", expected, ", actual: ", actual)
	}
}
