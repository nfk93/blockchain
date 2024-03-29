package smart

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"io/ioutil"
	"os"
	"testing"
)

var _, pk = crypto.KeyGen(2048)

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

func TestInitiateContract(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	addr, remaining, err := InitiateContract(pk, "nonce", fundme, 1000000, 100000, 10000, "1")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if remaining >= 900000 {
		t.Errorf("should have less remaining gas. Had: %d", remaining)
	}
	if c, exists := contracts[addr]; exists {
		if c.CreatedAtSlot != 5 {
			t.Errorf("CreatedAtSlot has wrong value")
		}
	} else {
		t.Errorf("contract doesn't exist")
	}
	contstate, exists := stateTree["1"].contractStates[addr]
	if !exists {
		t.Errorf("contract state doesn't exist")
	} else {
		if contstate.Storagecap != 10000 {
			t.Errorf("")
		}
		if !value.Equals(contstate.Storage, getFundmeStorage("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1100000, 0)) {
			t.Errorf("Storage has wrong value of %s", contstate.Storage)
		}
	}
}

func TestInitiateContract2(t *testing.T) {
	// not high enough Storage limit
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	fundme := getFundMeCode(t)
	_, remaining, err := InitiateContract(pk, "nonce", fundme, 1000000, 100000, 100, "1")
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
	_, remaining, err := InitiateContract(pk, "nonce", fundme, 1000000, 9000, 10000, "1")
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
	_, remaining, err := InitiateContract(pk, "nonce", fundme, 200000, 100000, 10000, "1")
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
	addr, _, _ := InitiateContract(pk, "nonce", fundme, 400000, 100000, 10000, "1")
	ledger, trans, remainingGas, err := CallContract(addr, "main", "kn1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		100000, 40000, "1")

	if err != nil {
		t.Errorf("error in contractcall: %s", err.Error())
		return
	}

	fundmestate, exists := stateTree["1"].contractStates[addr]
	if !exists {
		t.Errorf("contract state doesn't exist1")
	} else {
		if fundmestate.Storagecap != 10000 {
			t.Errorf("")
		}
		if !value.Equals(fundmestate.Storage, getFundmeStorage("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1100000, 100000)) {
			t.Errorf("Storage has wrong value of %s", fundmestate.Storage)
		}
		if fundmestate.Balance != 100000 {
			t.Errorf("")
		}
	}

	params := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab"
	ledger, trans, remainingGas, err = CallContract(addr, "main", "kn1"+params,
		1100000, 40000, "1")
	fundmestate, exists = stateTree["1"].contractStates[addr]
	if !exists {
		t.Errorf("contract state doesn't exist2")
	} else {
		if fundmestate.Storagecap != 10000 {
			t.Errorf("")
		}
		if !value.Equals(fundmestate.Storage, getFundmeStorage("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1100000, 1100000)) {
			t.Errorf("Storage has wrong value of %s", fundmestate.Storage)
		}
		if fundmestate.Balance != 1100000 {
			t.Errorf("")
		}
	}
	if remainingGas >= 40000 {
		t.Errorf("")
	}
	if len(trans) != 1 {
		t.Errorf("")
	} else {
		transaction := trans[0]
		if transaction.Amount != 100000 || transaction.To != params {
			t.Errorf("")
		}
	}
	if v, exists := ledger[addr]; !exists || v != 1100000 {
		t.Errorf("")
	}

	ledger, trans, remainingGas, err = CallContract(addr, "main", "kn1"+params,
		0, 40000, "1")
	fundmestate, exists = stateTree["1"].contractStates[addr]
	if !exists {
		t.Errorf("contract state doesn't exist2")
	} else {
		if fundmestate.Storagecap != 10000 {
			t.Errorf("")
		}
		if !value.Equals(fundmestate.Storage, getFundmeStorage("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1100000, 1100000)) {
			t.Errorf("Storage has wrong value of %s", fundmestate.Storage)
		}
		if fundmestate.Balance != 0 {
			t.Errorf("")
		}
	}
	if remainingGas >= 40000 {
		t.Errorf("")
	}
	if len(trans) != 1 {
		t.Errorf("")
	} else {
		transaction := trans[0]
		if transaction.Amount != 1100000 || transaction.To != "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" {
			t.Errorf("")
		}
	}
	if v, exists := ledger[addr]; !exists || v != 0 {
		t.Errorf("")
	}
}

func TestBranching(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	code := getSimpleIntStorage(t)
	addr, _, err := InitiateContract(pk, "nonce", code, 130000, 10000, 64, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	_, _, _, err = CallContract(addr, "main", "1", 0, 20000, "1")
	_, _ = NewBlockTreeNode("2", "1", 8)
	_, _, _, err = CallContract(addr, "main", "1", 0, 20000, "2")
	_, _ = NewBlockTreeNode("3", "1", 9)
	_, _, _, err = CallContract(addr, "main", "4", 1, 20000, "3")

	block1state := stateTree["1"].contractStates[addr]
	if !value.Equals(block1state.Storage, value.IntVal{1}) {
		t.Errorf("")
	}
	if block1state.Balance != 0 {
		t.Errorf("")
	}

	block2state := stateTree["2"].contractStates[addr]
	if !value.Equals(block2state.Storage, value.IntVal{2}) {
		t.Errorf("")
	}
	if block2state.PrepaidStorage != 10000-3*64 {
		t.Errorf("")
	}
	if block2state.Balance != 0 {
		t.Errorf("")
	}

	block3state := stateTree["3"].contractStates[addr]
	if !value.Equals(block3state.Storage, value.IntVal{5}) {
		t.Errorf("")
	}
	if block3state.PrepaidStorage != 10000-4*64 {
		t.Errorf("")
	}
	if block3state.Balance != 1 {
		t.Errorf("")
	}
}

func TestStorageSizeIncrease(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	code := getIntListStorage(t)
	addr, _, err := InitiateContract(pk, "nonce", code, 150000, 10000, 64, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	_, _, _, err = CallContract(addr, "main", "1", 0, 20000, "1")
	if err == nil || err.Error() != "Storage cap exceeded" {
		t.Errorf("")
	}
}

func TestChainCalls(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	code1 := getContract1(t)
	code2 := getContract2(t)
	addr1, _, err := InitiateContract(pk, "nonce", code1, 200000, 10000, 64, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	addr2, _, err := InitiateContract(pk, "nonce", code2, 200000, 10000, 64, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	_, transfers, _, err := CallContract(addr1, "main", fmt.Sprintf("kn2%s", addr2), 33, 100000, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(transfers) != 2 {
		t.Errorf("")
	} else {
		trans1 := transfers[0]
		if trans1.To != "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff" || trans1.Amount != 11 {
			t.Errorf("%s, %d", trans1.To, trans1.Amount)
		}
		trans2 := transfers[1]
		if trans2.To != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" || trans2.Amount != 5 {
			t.Errorf("%s, %d", trans2.To, trans2.Amount)
		}
	}

	blockstate := stateTree["1"]
	contractstate1 := blockstate.contractStates[addr1]
	contractstate2 := blockstate.contractStates[addr2]
	if contractstate1.Balance != 11 {
		t.Errorf("")
	}
	if contractstate2.Balance != 6 {
		t.Errorf("")
	}
}

func TestExpiringContract(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	code := getIntListStorage(t)
	addr1, _, err := InitiateContract(pk, "nonce", code, 150000, 1500, 150, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	code = getSimpleIntStorage(t)
	addr2, _, err := InitiateContract(pk, "nonce", code, 150000, 1100, 100, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	expiring, reward := NewBlockTreeNode("2", "1", 14)
	if len(expiring) != 0 {
		t.Errorf("")
	}
	if reward != (14-5)*(100+150) {
		t.Errorf("")
	}

	expiring, reward = NewBlockTreeNode("3", "1", 30)
	if len(expiring) != 2 {
		t.Errorf("")
	}
	if reward != 1100+1500 {
		t.Errorf("")
	}

	expiring, reward = NewBlockTreeNode("4", "2", 16)
	if len(expiring) != 1 {
		t.Errorf("")
	}
	if reward != 200+150 {
		t.Errorf("")
	}

	_, _, _, err = CallContract(addr1, "main", "1", 0, 100000, "4")
	if err == nil {
		t.Errorf("")
	}
	_, _, _, err = CallContract(addr2, "main", "1", 0, 100000, "4")
	if err != nil {
		t.Errorf("")
	}
}

func TestFinalizeBlock1(t *testing.T) {
	// check that contracts that are expired in all branches descending branches from a finalization node is deleted
	// from the contract map
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	code := getIntListStorage(t)
	addr, _, err := InitiateContract(pk, "nonce", code, 150000, 1000, 100, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	_, _ = NewBlockTreeNode("2", "1", 16)
	_, _ = NewBlockTreeNode("3", "2", 20)
	FinalizeBlock("2")
	if _, exists := contracts[addr]; exists {
		t.Errorf("")
	}
}

func TestFinalizeBlock2(t *testing.T) {
	// check that contracts that have expired in a finalizing block but the later reinitiated is not deleted on
	// finalizing said block
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	code := getIntListStorage(t)
	_, _, err := InitiateContract(pk, "nonce", code, 150000, 1000, 100, "1")
	if err != nil {
		t.Errorf(err.Error())
	}
	_, _ = NewBlockTreeNode("2", "1", 16)
	_, _ = NewBlockTreeNode("3", "2", 20)
	addr, _, err := InitiateContract(pk, "nonce", code, 150000, 1000, 64*2, "3")
	FinalizeBlock("2")
	_, _, _, err = CallContract(addr, "main", "1", 0, 100000, "3")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestNewBlock(t *testing.T) {
	reset()
	_, _ = NewBlockTreeNode("1", "genesis", 5)
	_, _ = SetStartingPointForNewBlock("1", 11)
	fundme := getFundMeCode(t)
	addr, _, _ := InitiateContractOnNewBlock(pk, "nonce", fundme, 400000, 100000, 10000)
	_, _, _, err := CallContractOnNewBlock(addr, "main", "kn1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		100000, 40000)
	if err != nil {
		t.Errorf("error in contractcall: %s", err.Error())
		return
	}
	fundmestate, exists := newBlockState.contractStates[addr]
	if !exists {
		t.Errorf("contract state doesn't exist1")
	} else {
		if fundmestate.Storagecap != 10000 {
			t.Errorf("")
		}
		if !value.Equals(fundmestate.Storage, getFundmeStorage("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1100000, 100000)) {
			t.Errorf("Storage has wrong value of %s", fundmestate.Storage)
		}
		if fundmestate.Balance != 100000 {
			t.Errorf("")
		}
	}

	// check that real tree is not mutated
	if len(stateTree) != 2 {
		t.Errorf("")
	}

	// check that finishing creating new block deletes the data
	DoneCreatingNewBlock()
	if newBlockContracts != nil {
		t.Errorf("")
	}
	if newBlockState.contractStates != nil || newBlockState.slot != 0 || newBlockState.parenthash != "" {
		t.Errorf("")
	}

	// testing that calling new contract doesn't mute the parent block
	addr, _, _ = InitiateContract(pk, "nonce", fundme, 400000, 100000, 10000, "1")
	previous := stateTree["1"].contractStates[addr]
	prevBal := previous.Balance
	prevPre := previous.PrepaidStorage
	fmt.Println("old prepaid:     ", prevPre)
	prevSto := value.Copy(previous.Storage)
	prevCap := previous.Storagecap
	_, _ = SetStartingPointForNewBlock("1", 12)
	_, _, _, err = CallContractOnNewBlock(addr, "main", "kn1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		100000, 40000)
	if err != nil {
		t.Errorf("error in contractcall: %s", err.Error())
		return
	}
	fmt.Println(newBlockState.contractStates[addr].PrepaidStorage)
	if newBlockState.contractStates[addr].PrepaidStorage >= 100000 {
		t.Errorf("prepaid should be lower")
	}

	DoneCreatingNewBlock()
	currentstate := stateTree["1"].contractStates[addr]
	fmt.Println("current prepaid: ", currentstate.PrepaidStorage)
	fmt.Println(currentstate.Storage)
	if prevBal != currentstate.Balance {
		t.Errorf("parent balance mutated")
	}
	if prevPre != currentstate.PrepaidStorage {
		t.Errorf("parent prepaid mutated")
	}
	if prevCap != currentstate.Storagecap {
		t.Errorf("parent storage cap mutated")
	}
	if !value.Equals(prevSto, currentstate.Storage) {
		t.Errorf("parent storage mutated")
	}
}

func getCodeBytes(t *testing.T, filepath string) ([]byte, error) {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Errorf("Error reading testfile %s", filepath)
		panic(err)
		return nil, err
	}
	return dat, nil
}

func getFundMeCode(t *testing.T) []byte {
	dat, _ := getCodeBytes(t, os.Getenv("GOPATH")+"/src/github.com/nfk93/blockchain/usecases/fundme")
	return dat
}

func getSimpleIntStorage(t *testing.T) []byte {
	dat, _ := getCodeBytes(t, "testcases/simple_int_storage")
	return dat
}

func getIntListStorage(t *testing.T) []byte {
	dat, _ := getCodeBytes(t, "testcases/intlist_storage")
	return dat
}

func getContract1(t *testing.T) []byte {
	dat, _ := getCodeBytes(t, "testcases/contract1")
	return dat
}

func getContract2(t *testing.T) []byte {
	dat, _ := getCodeBytes(t, "testcases/contract2")
	return dat
}

func getFundmeStorage(owner string, fundgoal uint64, amountrsd uint64) value.Value {
	return value.StructVal{Field: map[string]value.Value{
		"owner":         value.KeyVal{owner},
		"funding_goal":  value.KoinVal{fundgoal},
		"amount_raised": value.KoinVal{amountrsd}}}
}

func reset() {
	contracts = make(map[string]contract)
	stateTree = make(map[string]state)
	StartSmartContractLayer("genesis", false)
	DoneCreatingNewBlock()
}
