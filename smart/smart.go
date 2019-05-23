package smart

import (
	"crypto/sha256"
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter"
	"github.com/nfk93/blockchain/smart/interpreter/ast"
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"github.com/nfk93/blockchain/smart/paramparser"
	"log"
)

type state struct {
	contractStates map[string]contractState
	slot           uint64
	parenthash     string
}

type contractState struct {
	balance        uint64
	prepaidStorage uint64
	storage        value.Value
	storagecap     uint64
}

var contracts = make(map[string]contract)
var stateTree = make(map[string]state)
var newBlockState state
var newBlockContracts = make(map[string]contract)

func StartSmartContractLayer(genesishash string) {
	contractStates := make(map[string]contractState)
	stateTree[genesishash] = state{contractStates, 0, ""}
}

/*
 * Precondition: parenthash points to an existing state, i.e. _, exists := stateTree[parenthash] is always true
 */
func NewBlockTreeNode(blockhash, parenthash string, slot uint64) (expiringContracts []string, storagereward uint64) {
	expiring, newstate, reward := getNewState(parenthash, slot)
	stateTree[blockhash] = newstate
	return expiring, reward
}

func getNewState(parenthash string, slot uint64) (expires []string, s state, storageReward uint64) {
	expiring := make([]string, 0)
	parentState := stateTree[parenthash]
	slots := slot - parentState.slot
	tempStates := make(map[string]contractState)
	reward := uint64(0)
	for k, v := range parentState.contractStates {
		if v.prepaidStorage < slots*v.storagecap {
			reward += v.prepaidStorage
			expiring = append(expiring, k)
		} else {
			contractReward := slots * v.storagecap
			v.prepaidStorage -= contractReward // this is positive, and so doesn't panic
			tempStates[k] = v
			reward += contractReward
		}
	}
	return expiring, state{tempStates, slot, parenthash}, reward
}

/*
 * Precondition: blockhash points to an existing state, i.e. _, exists := stateTree[blockhash] is always true
 */
func CallContract(
	address string,
	entry string,
	params string,
	amount uint64,
	gas_ uint64,
	blockhash string,
) (resultLedger map[string]uint64, transfers []ContractTransaction, remainingGas uint64, callError error) {
	blockstate, exists := stateTree[blockhash]
	if !exists {
		// should never happen, because of precondition
		errstring := fmt.Sprintf("blockhash node does not exist for hash: %s", blockhash)
		log.Fatal(errstring)
		return nil, nil, 0, nil
	}

	newstate, transfers, remainingGas, err := handleContractCall(blockstate, contracts, amount, gas_, address, entry, params)
	if err != nil {
		return nil, nil, remainingGas, callError
	} else {
		stateTree[blockhash] = newstate
		return getContractBalances(newstate.contractStates), transfers, remainingGas, nil
	}
}

/*
 * Precondition: blockhash points to an existing state, i.e. _, exists := stateTree[blockhash] is always true
 */
func InitiateContract(
	contractCode []byte,
	gas uint64,
	prepaid uint64,
	storageLimit uint64,
	blockhash string,
) (addr string, remainingGas uint64, err error) {

	blockstate, exists := stateTree[blockhash]
	if !exists {
		// should never happen, because of precondition
		errstring := fmt.Sprintf("blockhash node does not exist for hash: %s", blockhash)
		log.Fatal(errstring)
		return "", 0, nil
	}

	address := getAddress(contractCode)
	texp, newstate, remainingGas, err := initiateContract(contractCode, address, gas, prepaid, storageLimit, blockstate)
	if err != nil {
		return "", remainingGas, err
	} else {
		contracts[address] = contract{string(contractCode), texp}
		stateTree[blockhash] = newstate
		return address, remainingGas, nil
	}
}

func FinalizeBlock(blockHash string) {
	// TODO Use for deleting old contracts
	// do this by going through the state of each contract in the blockstate and deleting all contracts for which its
	// expiration slot is lower than the slot of the block
}

/*
 * Precondition: parenthash points to an existing state, i.e. _, exists := stateTree[parenthash] is always true
 */
func ResetAndSetNewBlockStartPoint(parenthash string, slot uint64) (expiring []string, err error) {
	newBlockContracts = make(map[string]contract)
	expires, newstate, _ := getNewState(parenthash, slot)
	newBlockState = newstate
	return expires, nil
}

func CallContractOnNewBlock(
	address string,
	entry string,
	params string,
	amount uint64,
	gas_ uint64,
) (resultLedger map[string]uint64, transfers []ContractTransaction, remainingGas uint64, callError error) {

	allcontracts := make(map[string]contract)
	for k, v := range contracts {
		allcontracts[k] = v
	}
	for k, v := range newBlockContracts {
		allcontracts[k] = v
	}

	newstate, transfers, remainingGas, err := handleContractCall(newBlockState, allcontracts, amount, gas_, address, entry, params)
	if err != nil {
		return nil, nil, remainingGas, callError
	} else {
		newBlockState = newstate
		return getContractBalances(newstate.contractStates), transfers, remainingGas, nil
	}
}

/*
 * Precondition: newBlockState is defined
 */
func InitiateContractOnNewBlock(
	contractCode []byte,
	gas uint64,
	prepaid uint64,
	storageLimit uint64,
) (addr string, remainingGas uint64, err error) {
	address := getAddress(contractCode)
	texp, newstate, remainingGas, err := initiateContract(contractCode, address, gas, prepaid, storageLimit, newBlockState)
	if err != nil {
		return "", remainingGas, err
	} else {
		newBlockContracts[address] = contract{string(contractCode), texp}
		newBlockState = newstate
		return address, remainingGas, nil
	}
}

func DoneCreatingNewBlock() {
	newBlockContracts = nil
	newBlockState = state{}
}

type ContractTransaction struct {
	To     string
	Amount uint64
}

func initiateContract(
	contractCode []byte,
	address string,
	gas uint64,
	prepaid uint64,
	storageLimit uint64,
	blockstate state,
) (texp ast.TypedExp, s state, remainingGas uint64, err error) {
	if storageLimit == 0 {
		return ast.TypedExp{}, state{}, gas, fmt.Errorf("storagelimit can't be 0")
	}
	if prepaid == 0 {
		return ast.TypedExp{}, state{}, gas, fmt.Errorf("initial storage exceeds storage cap")
	}

	texp, initstor, remainingGas, returnErr := interpreter.InitiateContract(contractCode, gas)

	if returnErr != nil {
		return ast.TypedExp{}, state{}, remainingGas, returnErr
	} else {
		if initstor.Size() > storageLimit {
			return ast.TypedExp{}, state{}, remainingGas, fmt.Errorf("initial storage exceeds storage cap")
		}

		tempStates := make(map[string]contractState)
		for k, v := range blockstate.contractStates {
			tempStates[k] = v
		}

		tempStates[address] = contractState{0, prepaid, initstor,
			storageLimit}
		newState := state{tempStates, blockstate.slot, blockstate.parenthash}
		return texp, newState, remainingGas, nil
	}
}

func handleContractCall(
	blockstate state,
	contracts_ map[string]contract,
	amount, gas_ uint64,
	address, entry, params string,
) (newstate state, transfers []ContractTransaction, remainingGas uint64, err error) {
	// initial cost
	gas := gas_
	if int64(gas)-10000 < 0 {
		gas = 0
		return state{}, nil, gas, fmt.Errorf("not enough gas. calling a contract has a minimum cost of 0.1kn")
	} else {
		gas = gas - 10000
	}

	// decode parameters
	paramval, paramErr := decodeParameters(params)
	if paramErr != nil {
		return state{}, nil, gas, fmt.Errorf("syntax error in parameters:, %s", paramErr.Error())
	}

	tempStates := make(map[string]contractState)
	for k, v := range blockstate.contractStates {
		tempStates[k] = v
	}

	newStates, transfers, gas, callError := interpretContract(address, entry, paramval, amount, gas, tempStates, contracts_)
	if callError != nil {
		return state{}, nil, gas, callError
	} else {
		newState := state{newStates, blockstate.slot, blockstate.parenthash}
		return newState, transfers, gas, nil
	}
}

func interpretContract(
	address string,
	entry string,
	params value.Value,
	amount uint64,
	gas_ uint64,
	states map[string]contractState,
	contracts_ map[string]contract,
) (contractStates map[string]contractState, transfers []ContractTransaction, remainingGas uint64, callError error) {
	gas := gas_
	contract, exist1 := contracts_[address]
	state, exist2 := states[address]
	if !exist1 || !exist2 {
		return nil, nil, gas, fmt.Errorf("attempted to call non-existing contract at address %s", address)
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
	transfers, err, gas := handleOpList(oplist, states, contracts_, gas)
	if err != nil {
		return nil, nil, gas, err
	} else {
		return states, transfers, gas, nil
	}
}

func handleOpList(
	operations []value.Operation,
	tempStates map[string]contractState,
	contracts_ map[string]contract,
	gas uint64,
) ([]ContractTransaction, error, uint64) {

	transfers := make([]ContractTransaction, 0)
	for _, op := range operations {
		switch op.(type) {
		case value.ContractCall:
			callop := op.(value.ContractCall)
			tempStates_, trans, remainingGas, callError :=
				interpretContract(callop.Address, callop.Entry, callop.Params, callop.Amount, gas, tempStates, contracts_)
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
