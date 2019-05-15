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
	gas uint64,
) (resultLedger map[string]uint64, transfers []Transfer, remainingGas uint64, callError error) {
	tempBalances := make(map[string]uint64)
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
			remainingGas = 0
		} else {
			remainingGas = gas - 10000
		}
		return nil, nil, remainingGas, fmt.Errorf("attempted to call non-existing contract at address %s", address)
	}

	// check if contract exists
	oplist, sto, spent, remainingGas := interpreter.InterpretContractCall(contract.tabs, params, entry,
		contract.storage, amount, tempBalances[address], gas)

	contract.storage = sto
	tempContracts[address] = contract
	tempBalances[address] = tempBalances[address] + amount - spent

	// handle operation list
	transfers, err, remainingGas := handleOpList(oplist, address, tempBalances, tempContracts, gas)
	if err != nil {
		return contractBalances, nil, remainingGas, err
	} else {
		contracts = tempContracts
		contractBalances = tempBalances
		return contractBalances, transfers, remainingGas, nil
	}
}

func ExpireContract(address string) {
	delete(contracts, address)
	delete(contractBalances, address)
}

func InitiateContract(
	contractCode []byte,
	gas uint64,
) (remainingGas uint64, err error) {
	address := getAddress(contractCode)

	texp, initstor, remainingGas, returnErr := interpreter.InitiateContract(contractCode, gas)
	if returnErr != nil {
		return remainingGas, returnErr
	} else {
		contracts[address] = Contract{string(contractCode), texp, initstor}
		contractBalances[address] = 0
		return remainingGas, nil
	}
}

func handleOpList(operations []interpreter.Operation, caller string, tempBalances map[string]uint64,
	tempContracts map[string]Contract, gas uint64) ([]Transfer, error, uint64) {
	transfers := make([]Transfer, 0)
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

			trans, err, remainingGas := handleOpList(oplist, caller, tempBalances, tempContracts, remainingGas)
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
			transfers = append(transfers, Transfer{transferop.Key, caller, transferop.Amount})
		}
	}
	return transfers, nil, gas
}

func getAddress(contractCode []byte) string {
	bytes := sha256.Sum256(contractCode)
	return fmt.Sprintf("%x", bytes)
}
