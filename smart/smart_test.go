package smart

import (
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter"
	"io/ioutil"
	"os"
	"testing"
)

func TestCallContract(t *testing.T) {
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile")
	}
	texp, stor, _, err := interpreter.InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}
	contracts["A"] = Contract{string(dat), texp, stor}

	address := "A"
	entry := "main"
	params := interpreter.KeyVal{"asdasda"}
	amount := uint64(10000)
	gas := uint64(100000)

	_, _, _, err = CallContract(address, entry, params, amount, gas)

	fmt.Println(contracts["A"].storage)
	if contracts["A"].storage.(interpreter.StructVal).Field["amount_raised"].(interpreter.KoinVal).Value != 10000 {
		t.Errorf("contract A storage has wrong value in amount_raised")
	}
}
