package smart

import (
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestNewBlockTreeNode(t *testing.T) {
	reset()
	expiring, reward := NewBlockTreeNode("newblock", "genesis", 5)
	if len(expiring) != 0 {
		t.Errorf("")
	}
	if reward != 0 {
		t.Errorf("")
	}
}

func TestNewBlockTreeNode1(t *testing.T) {
	reset()
	expiring, reward := NewBlockTreeNode("newblock", "genesis", 5)
	if len(expiring) != 0 {
		t.Errorf("")
	}
	if reward != 0 {
		t.Errorf("")
	}
}

func TestInitiateContract(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	addr, remaining, err := InitiateContract(fundme, 1000000, 100000, 10000, "1")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if remaining >= 900000 {
		t.Errorf("should have less remaining gas. Had: %d", remaining)
	}
	if c, exists := contracts[addr]; exists {
		if c.createdAtSlot != 5 {
			t.Errorf("createdAtSlot has wrong value")
		}
	} else {
		t.Errorf("contract doesn't exist")
	}
	contstate, exists := stateTree["1"].contractStates[addr]
	if !exists {
		t.Errorf("contract state doesn't exist")
	} else {
		if contstate.storagecap != 10000 {
			t.Errorf("")
		}
		if !value.Equals(contstate.storage, getFundmeStorage("1234567890abcdef1234567890abcdef", 1100000, 0)) {
			t.Errorf("storage has wrong value of %s", contstate.storage)
		}
	}
}

func TestInitiateContract2(t *testing.T) {
	// not high enough storage limit
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	_, remaining, err := InitiateContract(fundme, 1000000, 100000, 100, "1")
	if err == nil {
		t.Errorf("should have error")
		return
	}
	if remaining >= 900000 {
		t.Errorf("should have less remaining gas. Had: %d", remaining)
	}
}

func TestInitiateContract3(t *testing.T) {
	// not enough prepaid
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	_, remaining, err := InitiateContract(fundme, 1000000, 9000, 10000, "1")
	if err == nil {
		t.Errorf("should have error")
		return
	}
	if remaining >= 1000000 {
		t.Errorf("should have less remaining gas. Had: %d", remaining)
	}
}

func TestInitiateContract4(t *testing.T) {
	// not enough gas
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	_, remaining, err := InitiateContract(fundme, 200000, 100000, 10000, "1")
	if err == nil {
		t.Errorf("should have error")
		return
	}
	if remaining != 0 {
		t.Errorf("should have no remaining gas. Had: %d", remaining)
	}
}

func TestCallContract(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	addr, _, _ := InitiateContract(fundme, 400000, 100000, 10000, "1")
	ledger, trans, remainingGas, err := CallContract(addr, "main", "kn1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		100000, 10000, "1")

	if err != nil {
		t.Errorf("error in contractcall: %s", err.Error())
		return
	}
	fmt.Println(ledger)
	fmt.Println(trans)
	fmt.Println(remainingGas)
}

// updating contractstate 1 and 2
// branching contracts
// chain of calls
// previous state not mutated
// expiring contract
// storage reward

func getFundMeCode(t *testing.T) []byte {
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile_noerror")
		log.Fatal("can't read file")
		return nil
	}
	return dat
}

func getFundmeStorage(owner string, fundgoal uint64, amountrsd uint64) value.Value {
	return value.StructVal{Field: map[string]value.Value{
		"owner":         value.KeyVal{owner},
		"funding_goal":  value.KoinVal{fundgoal},
		"amount_raised": value.KoinVal{amountrsd}}}
}

/*
func TestCallContract(t *testing.T) {
	reset()
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}


	address := "A"
	entry := "main"
	params := value.KeyVal{"asdasda"}
	amount := uint64(500000)
	gas := uint64(100000)

	_, transfers, _, err := CallContract(address, entry, params, amount, gas)
	if contracts["A"].storage.(value.StructVal).Field["amount_raised"].(value.KoinVal).Value != 500000 {
		t.Errorf("contract A storage has wrong value in amount_raised")
	}
	_, transfers, _, err = CallContract(address, entry, params, amount, gas)
	_, transfers, _, err = CallContract(address, entry, params, amount, gas)
	if contracts["A"].storage.(value.StructVal).Field["amount_raised"].(value.KoinVal).Value != 1100000 {
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
*/
/*
func TestCallContract2(t *testing.T) {
	reset()
	// testing chain of calls
	dat, err := ioutil.ReadFile("testcases/contract1")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	texp, stor, _, err := interpreter.InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}
	contracts["contract1"] = contract{string(dat), texp, stor}

	dat, err = ioutil.ReadFile("testcases/contract2")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	texp, stor, _, err = interpreter.InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}
	contracts["contract2"] = contract{string(dat), texp, stor}

	address := "contract1"
	entry := "main"
	params := value.UnitVal{}
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
	reset()
	dat, err := ioutil.ReadFile("testcases/contract1")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	_, _, _, err = InitiateContract(dat, 1000000)
	if err != nil {
		t.Fail()
	}
	if contract, exists := contracts[getAddress(dat)]; !exists || contract.storage.(value.IntVal).Value != 0 {
		t.Fail()
	}
}

func TestExpireContract(t *testing.T) {
	reset()
	dat1, err := ioutil.ReadFile("testcases/contract1_altaddress")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	dat2, err := ioutil.ReadFile("testcases/contract2")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}

	add1, _, _, err := InitiateContract(dat1, 99999999999999)
	if err != nil {
		t.Errorf("error initiating contract1: %s", err.Error())
	}
	add2, _, _, err := InitiateContract(dat2, 99999999999999999)
	if err != nil {
		t.Errorf("error initiating contract2: %s", err.Error())
	}

	ledger, transfers, remainingGas, err := CallContract(add1, "main", value.UnitVal{}, 10000, 100000)
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
	ledger, transfers, remainingGas, err = CallContract(add1, "main", value.UnitVal{}, 10000, 100000)
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
	reset()
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	addr, _, _, _ := InitiateContract(dat, 999999999)
	_, _, _, err = CallContract(addr, "main", value.KeyVal{""}, 100, 1000)
	if err == nil {
		t.Errorf("this should return an error from running out of gas")
	}
}

func TestInsufficientFunds(t *testing.T) {
	reset()
	dat, err := ioutil.ReadFile("testcases/expensive")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	addr, _, _, _ := InitiateContract(dat, 999999999)
	_, _, _, err = CallContract(addr, "main", value.UnitVal{}, 100, 100000000000)
	if err == nil {
		t.Errorf("this should return an error from having insufficient funds")
	}
}
*/

func reset() {
	contracts = make(map[string]contract)
	stateTree = make(map[string]state)
	StartSmartContractLayer("genesis")
	DoneCreatingNewBlock()
}
