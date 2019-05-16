package smart

import (
	"github.com/nfk93/blockchain/smart/interpreter"
	"io/ioutil"
	"os"
	"testing"
)

func TestCallContract(t *testing.T) {
	resetMaps()
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
	amount := uint64(500000)
	gas := uint64(100000)

	_, transfers, _, err := CallContract(address, entry, params, amount, gas)
	if contracts["A"].storage.(interpreter.StructVal).Field["amount_raised"].(interpreter.KoinVal).Value != 500000 {
		t.Errorf("contract A storage has wrong value in amount_raised")
	}
	_, transfers, _, err = CallContract(address, entry, params, amount, gas)
	_, transfers, _, err = CallContract(address, entry, params, amount, gas)
	if contracts["A"].storage.(interpreter.StructVal).Field["amount_raised"].(interpreter.KoinVal).Value != 1100000 {
		t.Errorf("contract A storage has wrong value in amount_raised")
	}
	if len(transfers) != 1 {
		t.Errorf("transferlist not long enough")
	} else {
		trans := transfers[0]
		if trans.To != "asdasda" {
			t.Errorf("wrong to in transaction: %s", trans.To)
		}
		if trans.Amount != 400000 {
			t.Errorf("wrong amount in transaction: %d", trans.Amount)
		}
	}
}

func TestCallContract2(t *testing.T) {
	resetMaps()
	// testing chain of calls
	dat, err := ioutil.ReadFile("testcases/contract1")
	if err != nil {
		t.Error("Error reading testfile")
	}
	texp, stor, _, err := interpreter.InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}
	contracts["contract1"] = Contract{string(dat), texp, stor}

	dat, err = ioutil.ReadFile("testcases/contract2")
	if err != nil {
		t.Error("Error reading testfile")
	}
	texp, stor, _, err = interpreter.InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}
	contracts["contract2"] = Contract{string(dat), texp, stor}

	address := "contract1"
	entry := "main"
	params := interpreter.UnitVal{}
	amount := uint64(500000)
	gas := uint64(100000)
	ledger, transfers, remainingGas, err := CallContract(address, entry, params, amount, gas)

	if ledger["contract1"] != uint64(500000-(2*(int(500000/19)))) {
		t.Errorf("contract1 has wrong balance: %d, expected %d", ledger["contract1"], 500000-(2*(int(500000/19))))
	}
	contract2balance := ledger["contract2"]
	if contract2balance != 500000/19-10 {
		t.Errorf("contract2 has wrong balance: %d", contract2balance)
	}
	if len(transfers) != 2 {
		t.Errorf("transfers list has wrong list: %d", len(transfers))
	} else {
		trans1 := transfers[0]
		trans2 := transfers[1]
		if trans1.Amount != 500000/19 || trans1.To != "key1" {
			t.Errorf("transaction 1 is wrong: %v", trans1)
		}
		if trans2.Amount != 10 || trans2.To != "key2" {
			t.Errorf("transaction 2 is wrong: %v", trans2)
		}
	}
	if err != nil {
		t.Errorf(err.Error())
	}
	if remainingGas > 100000 {
		t.Errorf("too much gas remaining")
	}
}

func TestInitiateContract(t *testing.T) {
	resetMaps()
	dat, err := ioutil.ReadFile("testcases/contract1")
	if err != nil {
		t.Error("Error reading testfile")
	}
	_, _, _, err = InitiateContract(dat, 1000000)
	if err != nil {
		t.Fail()
	}
	if contract, exists := contracts[getAddress(dat)]; !exists || contract.storage.(interpreter.IntVal).Value != 0 {
		t.Fail()
	}
}

func TestExpireContract(t *testing.T) {
	resetMaps()
	dat1, err := ioutil.ReadFile("testcases/contract1_altaddress")
	if err != nil {
		t.Error("Error reading testfile")
	}
	dat2, err := ioutil.ReadFile("testcases/contract2")
	if err != nil {
		t.Error("Error reading testfile")
	}

	add1, _, _, err := InitiateContract(dat1, 99999999999999)
	if err != nil {
		t.Errorf("error initiating contract1: %s", err.Error())
	}
	add2, _, _, err := InitiateContract(dat2, 99999999999999999)
	if err != nil {
		t.Errorf("error initiating contract2: %s", err.Error())
	}

	ledger, transfers, remainingGas, err := CallContract(add1, "main", interpreter.UnitVal{}, 10000, 100000)
	if ledger[add1] != uint64(10000-(2*(int(10000/19)))) {
		t.Errorf("contract1 has wrong balance: %d, expected %d", ledger[add1], 10000-(2*(int(500000/19))))
	}
	contract2balance := ledger[add2]
	if contract2balance != 10000/19-10 {
		t.Errorf("contract2 has wrong balance: %d", contract2balance)
	}
	if len(transfers) != 2 {
		t.Errorf("transfers list has wrong list: %d", len(transfers))
	} else {
		trans1 := transfers[0]
		trans2 := transfers[1]
		if trans1.Amount != 10000/19 || trans1.To != "key1" {
			t.Errorf("transaction 1 is wrong: %v", trans1)
		}
		if trans2.Amount != 10 || trans2.To != "key2" {
			t.Errorf("transaction 2 is wrong: %v", trans2)
		}
	}
	if err != nil {
		t.Errorf(err.Error())
	}
	if remainingGas > 100000 {
		t.Errorf("too much gas remaining")
	}

	ExpireContract(add2)
	ledger, transfers, remainingGas, err = CallContract(add1, "main", interpreter.UnitVal{}, 10000, 100000)
	if ledger[add2] != 0 {
		t.Errorf("contract 2 balance not empty")
	}
	if _, exists := contracts[add2]; exists {
		t.Errorf("contract 2 still exists2")
	}
	if len(transfers) != 0 {
		t.Errorf("transferlist should be empty")
	}
	if remainingGas >= 100000 {
		t.Errorf("no gas was used")
	}
	if err == nil {
		t.Errorf("no error resulted from call")
	}
}

func TestOutOfGas(t *testing.T) {
	resetMaps()
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile")
	}
	addr, _, _, _ := InitiateContract(dat, 999999999)
	_, _, _, err = CallContract(addr, "main", interpreter.KeyVal{""}, 100, 1000)
	if err == nil {
		t.Errorf("this should return an error from running out of gas")
	}
}

func TestInsufficientFunds(t *testing.T) {
	resetMaps()
	dat, err := ioutil.ReadFile("testcases/expensive")
	if err != nil {
		t.Error("Error reading testfile")
	}
	addr, _, _, _ := InitiateContract(dat, 999999999)
	_, _, _, err = CallContract(addr, "main", interpreter.UnitVal{}, 100, 100000000000)
	if err == nil {
		t.Errorf("this should return an error from having insufficient funds")
	}
}

func resetMaps() {
	contracts = make(map[string]Contract)
	contractBalances = make(map[string]uint64)
}
