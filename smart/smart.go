package smart

import (
	"crypto/sha256"
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter"
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"github.com/nfk93/blockchain/smart/paramparser"
	"log"
)

type state struct {
	contractStates map[string]contractState
	slot           uint64
	parentHash     string
}

type contractState struct {
	balance        uint64
	prepaidStorage uint64
	storage        value.Value
	storagecap     uint64
	expirationslot uint64
}

var contracts = make(map[string]Contract)
var stateTree = make(map[string]state)

/*
 * Precondition: parenthash points to an existing state, i.e. _, exists := stateTree[parenthash] is always true
 */
func CallContract(
	address string,
	entry string,
	params string,
	amount uint64,
	gas_ uint64,
	blockhash, parenthash string,
	slot uint64, // TODO use slot and parent slot to update prepaidstorage and expire contracts
) (resultLedger map[string]uint64, transfers []ContractTransaction, remainingGas uint64, callError error) {
	if blockhash == "" {
		// TODO making new block
	}

	blockstate, exists := stateTree[blockhash]
	if !exists {
		blockstate, exists = stateTree[parenthash]
	}
	if !exists {
		// should never happen
		errstring := fmt.Sprintf("parenthash node does not exist for hash: %s", parenthash)
		log.Fatal(errstring)
		return nil, nil, 0, nil
	}

	// add contractcall to state for given blockhash
	contractstates := blockstate.contractStates

	tempStates := make(map[string]contractState)
	for k, v := range contractstates {
		tempStates[k] = v
	}

	// initial cost
	gas := gas_
	if int64(gas)-10000 < 0 {
		gas = 0
		return nil, nil, gas, fmt.Errorf("not enough gas. calling a contract has a minimum cost of 0.1kn")
	} else {
		gas = gas - 10000
	}

	// decode parameters
	paramval, paramErr := decodeParameters(params)
	if paramErr != nil {
		return nil, nil, gas, fmt.Errorf("syntax error in parameters:, %s", paramErr.Error())
	}

	newStates, transfers, gas, callError := interpretContract(address, entry, paramval, amount, gas, tempStates, slot)
	if callError != nil {
		return getContractBalances(contractstates), nil, gas, callError
	} else {
		newState := state{newStates, slot, parenthash}
		stateTree[blockhash] = newState
		return getContractBalances(newStates), transfers, gas, nil
	}
}

/*
 * Precondition: parenthash points to an existing state, i.e. _, exists := stateTree[parenthash] is always true
 */
func ExpiringContract(slot uint64, parenthash string) []string {
	result := make([]string, 0)
	for k, v := range stateTree[parenthash].contractStates {
		if v.expirationslot < slot {
			result = append(result, k)
		}
	}
	return result
}

func FinalizeBlock(blockHash string) {
	// TODO Use for deleting old contracts
	// do this by going through the state of each contract in the blockstate and deleting all contracts for which its
	// expiration slot is lower than the slot of the block
}

/*
 * Precondition: parenthash points to an existing state, i.e. _, exists := stateTree[parenthash] is always true
 */
func InitiateContract(
	contractCode []byte,
	gas uint64,
	prepaid uint64,
	storageLimit uint64,
	blockhash, parenthash string,
	slot uint64, // TODO use slot and parent slot to update prepaidstorage and expire contracts
) (addr string, remainingGas uint64, err error) {

	if storageLimit == 0 {
		return "", remainingGas, fmt.Errorf("storagelimit can't be 0")
	}
	if prepaid == 0 {
		return "", remainingGas, fmt.Errorf("initial storage exceeds storage cap")
	}

	texp, initstor, remainingGas, returnErr := interpreter.InitiateContract(contractCode, gas)

	if returnErr != nil {
		return "", remainingGas, returnErr
	} else {
		if initstor.Size() > storageLimit {
			return "", remainingGas, fmt.Errorf("initial storage exceeds storage cap")
		}

		address := getAddress(contractCode)
		contracts[address] = Contract{string(contractCode), texp}

		blockstate, exists := stateTree[blockhash]
		if !exists {
			blockstate, exists = stateTree[parenthash]
		}
		if !exists {
			errstring := fmt.Sprintf("parenthash node does not exist for hash: %s", parenthash)
			log.Fatal(errstring)
			return "", 0, nil
		}

		tempStates := make(map[string]contractState)
		for k, v := range blockstate.contractStates {
			tempStates[k] = v
		}

		expiration := slot + (prepaid / storageLimit)
		tempStates[address] = contractState{0, prepaid, initstor,
			storageLimit, expiration}
		newState := state{tempStates, slot, parenthash}
		stateTree[blockhash] = newState
		return address, remainingGas, nil
	}
}

/*
 * Precondition: blockhash is a hash for a block that exists in the statetree
 */
func StorageReward(blockhash string) (reward uint64) {
	block := stateTree[blockhash]
	parent := stateTree[block.parentHash]
	slots := block.slot - parent.slot

	reward = uint64(0)

	for _, k := range parent.contractStates {
		reward += min(k.prepaidStorage, slots*k.storagecap)
	}

	return reward
}

type ContractTransaction struct {
	To     string
	Amount uint64
}

func interpretContract(
	address string,
	entry string,
	params value.Value,
	amount uint64,
	gas_ uint64,
	states map[string]contractState,
	slot uint64,
) (contractStates map[string]contractState, transfers []ContractTransaction, remainingGas uint64, callError error) {
	gas := gas_
	contract, exist1 := contracts[address]
	state, exist2 := states[address]
	if !exist1 || !exist2 {
		return nil, nil, gas, fmt.Errorf("attempted to call non-existing contract at address %s", address)
	}

	// check if contract has expired
	if slot > state.expirationslot {
		return nil, nil, gas, fmt.Errorf("attempted to call expired contract")
	}

	oplist, sto, spent, gas := interpreter.InterpretContractCall(contract.tabs, params, entry, state.storage, amount,
		state.balance, gas)

	if sto.Size() > state.storagecap {
		return nil, nil, gas, fmt.Errorf("storage cap exceeded")
	}

	state.storage = sto
	state.balance = state.balance + amount - spent // it is checked in the interpreter that this value isn't negative
	states[address] = state

	// handle operation list
	transfers, err, gas := handleOpList(oplist, states, gas, slot)
	if err != nil {
		return nil, nil, gas, err
	} else {
		return states, transfers, gas, nil
	}
}

func handleOpList(
	operations []value.Operation,
	tempStates map[string]contractState,
	gas, slot uint64,
) ([]ContractTransaction, error, uint64) {

	transfers := make([]ContractTransaction, 0)
	for _, op := range operations {
		switch op.(type) {
		case value.ContractCall:
			callop := op.(value.ContractCall)
			tempStates_, trans, remainingGas, callError :=
				interpretContract(callop.Address, callop.Entry, callop.Params, callop.Amount, gas, tempStates, slot)
			if callError != nil {
				return nil, callError, remainingGas
			} else {
				tempStates = tempStates_
				gas = remainingGas
				transfers = append(transfers, trans...)
			}
		case value.FailWith:
			return nil, fmt.Errorf(op.(value.FailWith).Msg), gas
		case value.Transfer:
			transferop := op.(value.Transfer)
			transfers = append(transfers, ContractTransaction{transferop.Key, transferop.Amount})
		}
	}
	return transfers, nil, gas
}

func decodeParameters(params string) (value.Value, error) {
	return paramparser.ParseParams(params)
}

func getAddress(contractCode []byte) string {
	bytes := sha256.Sum256(contractCode)
	return fmt.Sprintf("%x", bytes)
}

func getContractBalances(states map[string]contractState) map[string]uint64 {
	result := make(map[string]uint64)
	for k, v := range states {
		result[k] = v.balance
	}
	return result
}

func min(a, b uint64) uint64 {
	if a <= b {
		return a
	} else {
		return b
	}
}
