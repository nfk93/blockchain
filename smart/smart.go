package smart

import (
	"crypto/sha256"
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter"
	"log"
)

type state struct {
	contractStates map[string]contractState
	parentHash     string
}

type contractState struct {
	balance        uint64
	prepaidStorage uint64
	storage        interpreter.Value
	storagecap     uint64
}

var contracts = make(map[string]Contract)
var stateTree = make(map[string]state)

/*
 * Precondition: parenthash points to an existing state, i.e. _, exists := stateTree[parenthash] is always true
 */
func CallContract(
	address string,
	entry string,
	params interpreter.Value,
	amount uint64,
	gas uint64,
	blockhash, parenthash string,
	slot uint64, // TODO use slot to check if a contract has expired. Don't expire contracts yet
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
	newStates, transfers, gas, callError := interpretContract(address, entry, params, amount, gas, contractstates)
	if callError != nil {
		return getContractBalances(contractstates), nil, gas, callError
	} else {
		blockstate.contractStates = newStates
		stateTree[blockhash] = blockstate
		return getContractBalances(newStates), transfers, gas, nil
	}
}

func ExpiringContract(slot uint64) []string {
	return nil
}

func FinalizeBlock(blockHash string) {
	// TODO Use for deleting old contracts
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
	slot uint64,
) (addr string, remainingGas uint64, err error) {
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
		blockstate.contractStates[address] = contractState{0, prepaid, initstor, storageLimit}
		stateTree[blockhash] = blockstate
		return address, remainingGas, nil
	}
}

func StorageCost(blockhash string) (reward uint64) {
	return 0 // TODO BIG FAT TODO
}

type ContractTransaction struct {
	To     string
	Amount uint64
}

func interpretContract(
	address string,
	entry string,
	params interpreter.Value,
	amount uint64,
	gas_ uint64,
	states map[string]contractState,
) (contractStates map[string]contractState, transfers []ContractTransaction, remainingGas uint64, callError error) {

	tempStates := make(map[string]contractState)
	for k, v := range states {
		tempStates[k] = v
	}
	gas := gas_

	contract, exist1 := contracts[address]
	state, exist2 := tempStates[address]
	if !exist1 || !exist2 {
		if int64(gas)-10000 < 0 {
			gas = 0
		} else {
			gas = gas - 10000
		}
		return nil, nil, gas, fmt.Errorf("attempted to call non-existing contract at address %s", address)
	}

	oplist, sto, spent, gas := interpreter.InterpretContractCall(contract.tabs, params, entry, state.storage, amount,
		state.balance, gas)

	// TODO: check if storage limit is exceeded
	state.storage = sto
	state.balance = state.balance + amount - spent // it is checked in the interpreter that this value isn't negative
	tempStates[address] = state

	// handle operation list
	transfers, err, gas := handleOpList(oplist, tempStates, gas)
	if err != nil {
		return states, nil, gas, err
	} else {
		return tempStates, transfers, gas, nil
	}
}

func handleOpList(
	operations []interpreter.Operation,
	tempStates map[string]contractState,
	gas uint64,
) ([]ContractTransaction, error, uint64) {
	transfers := make([]ContractTransaction, 0)
	for _, op := range operations {
		switch op.(type) {
		case interpreter.ContractCall:
			callop := op.(interpreter.ContractCall)

			contract, exist1 := contracts[callop.Address]
			state, exist2 := tempStates[callop.Address]
			if !exist1 || !exist2 {
				return nil, fmt.Errorf("attempted to call non-existing contract at address %s", callop.Address), gas
			}

			oplist, storage, spent, remainingGas := interpreter.InterpretContractCall(contract.tabs, callop.Params,
				callop.Entry, state.storage, callop.Amount, state.balance, gas)

			state.storage = storage
			state.balance = state.balance + callop.Amount - spent
			tempStates[callop.Address] = state

			trans, err, remainingGas := handleOpList(oplist, tempStates, remainingGas)
			if err != nil {
				return nil, err, remainingGas
			} else {
				gas = remainingGas
				transfers = append(transfers, trans...)
			}
		case interpreter.FailWith:
			return nil, fmt.Errorf(op.(interpreter.FailWith).Msg), gas
		case interpreter.Transfer:
			transferop := op.(interpreter.Transfer)
			transfers = append(transfers, ContractTransaction{transferop.Key, transferop.Amount})
		}
	}
	return transfers, nil, gas
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
