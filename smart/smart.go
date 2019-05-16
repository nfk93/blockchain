package smart

import (
	"crypto/sha256"
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter"
)

var contracts = make(map[string]Contract)
var contractBalances = make(map[string]uint64)

func CallContract(
	address string,
	entry string,
	params interpreter.Value,
	amount uint64,
	gas_ uint64,
	slot uint64, // TODO use slot to check if a contract has expired. Don't expire contracts yet
) (resultLedger map[string]uint64, transfers []ContractTransaction, remainingGas uint64, callError error) {
	tempBalances := make(map[string]uint64)

	gas := gas_

	for k, v := range contractBalances {
		tempBalances[k] = v
	}
	tempContracts := make(map[string]Contract)
	for k, v := range contracts {
		tempContracts[k] = v
	}

	contract, exist := tempContracts[address]
	if !exist {
		if int64(gas)-10000 < 0 {
			gas = 0
		} else {
			gas = gas - 10000
		}
		return nil, nil, remainingGas, fmt.Errorf("attempted to call non-existing contract at address %s", address)
	}

	oplist, sto, spent, gas := interpreter.InterpretContractCall(contract.tabs, params, entry,
		contract.storage, amount, tempBalances[address], gas)

	contract.storage = sto
	tempContracts[address] = contract
	tempBalances[address] = tempBalances[address] + amount - spent

	// handle operation list
	transfers, err, gas := handleOpList(oplist, tempBalances, tempContracts, gas)
	if err != nil {
		return contractBalances, nil, gas, err
	} else {
		contracts = tempContracts
		contractBalances = tempBalances
		return contractBalances, transfers, gas, nil
	}
}

func ExpiringContract(slot uint64) []string {
	return nil
}

func InitiateContract(
	contractCode []byte,
	gas uint64,
	prepaid uint64,
	storageLimit uint64,
	blockhash, parenthash string,
	slot uint64,
) (addr string, remainingGas uint64, err error) {
	address := getAddress(contractCode)

	// TODO: IMPORTANT!!! calculate storage size
	texp, initstor, remainingGas, returnErr := interpreter.InitiateContract(contractCode, gas)
	if returnErr != nil {
		return "", remainingGas, 0, returnErr
	} else {
		contracts[address] = Contract{string(contractCode), texp, initstor}
		contractBalances[address] = 0
		return address, remainingGas, 0, nil
	}
}

func StorageCost(blockhash string) (reward uint64) {
	return 0 // TODO BIG FAT TODO
}

type ContractTransaction struct {
	To     string
	Amount uint64
}

func handleOpList(operations []interpreter.Operation, tempBalances map[string]uint64,
	tempContracts map[string]Contract, gas uint64) ([]ContractTransaction, error, uint64) {
	transfers := make([]ContractTransaction, 0)
	for _, op := range operations {
		switch op.(type) {
		case interpreter.ContractCall:
			callop := op.(interpreter.ContractCall)

			contract, exist := tempContracts[callop.Address]
			if !exist {
				return nil, fmt.Errorf("attempted to call non-existing contract at address %s", callop.Address), gas
			}

			oplist, storage, spent, remainingGas := interpreter.InterpretContractCall(contract.tabs, callop.Params,
				callop.Entry, contract.storage, callop.Amount, tempBalances[callop.Address], gas)

			contract.storage = storage
			tempContracts[callop.Address] = contract
			tempBalances[callop.Address] = tempBalances[callop.Address] + callop.Amount - spent

			trans, err, remainingGas := handleOpList(oplist, tempBalances, tempContracts, remainingGas)
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
