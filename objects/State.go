package objects

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/smart"
	"sort"
	"strconv"
)

type ContractAccount struct {
	Owner       PublicKey
	Prepaid     uint64
	StorageCost uint64
}

type State struct {
	Ledger      map[string]uint64
	ConStake    map[string]uint64
	ConAccounts map[string]ContractAccount
	ParentHash  string
	TotalStake  uint64
}

func NewInitialState(key PublicKey) State {
	initialStake := uint64(1000000) // 1 mil
	ledger := make(map[string]uint64)
	conStake := make(map[string]uint64)
	conledger := make(map[string]ContractAccount)
	ledger[key.String()] = initialStake
	return State{ledger, conStake, conledger, "", initialStake}
}

//Returns gasCost
func (s *State) AddTransaction(t Transaction, gasCost uint64) uint64 {
	//TODO: Handle checks of legal transactions
	amountWithFees := t.Amount + gasCost

	if !t.VerifyTransaction() {
		fmt.Println("The transactions didn't verify", t)
		return 0
	}

	// Sender has to be able to pay both the amount and the fee
	if s.Ledger[t.From.String()] < amountWithFees {
		fmt.Println("Not enough money on senders account")
		return 0
	}

	s.Ledger[t.From.String()] -= amountWithFees
	s.Ledger[t.To.String()] += t.Amount
	s.TotalStake -= gasCost // Take the fee out of the system
	return gasCost
}

func (s *State) PayBlockRewardOrRemainGas(pk PublicKey, reward uint64) {
	s.Ledger[pk.String()] += reward
	s.TotalStake += reward // putting back the fees and an block reward if anyone claim it
}

func (s State) toString() string {
	var buf bytes.Buffer

	sortedLedger := make(map[string]uint64)

	keys := make([]string, 0, len(s.Ledger))
	for k := range s.Ledger {
		sortedLedger[k] = s.Ledger[k]
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(strconv.Itoa(int(sortedLedger[k])))
	}
	buf.WriteString(s.ParentHash)
	buf.WriteString(strconv.Itoa(int(s.TotalStake)))

	return buf.String()
}

func (s *State) SignHashedState(sk SecretKey) string {
	return Sign(HashSHA(s.toString()), sk)
}

func (s State) VerifyHashedState(sig string, pk PublicKey) bool {
	return Verify(HashSHA(s.toString()), sig, pk)
}

// Opens account for contract and moves prepaid to its account
func (s *State) InitializeContractAccount(addr string, owner PublicKey, prepaid uint64, storageCost uint64) {
	s.TotalStake -= prepaid
	s.ConAccounts[addr] = ContractAccount{owner, prepaid, storageCost}

}

// Used for putting more prepaid on an contract account
func (s *State) PrepayContracts(addr string, amount uint64) {
	s.TotalStake -= amount
	newBalance := s.ConAccounts[addr]
	newBalance.Prepaid += amount
	s.ConAccounts[addr] = newBalance
}

// Used for handling contract layer transaction to users
func (s *State) AddContractTransaction(t smart.ContractTransaction) {
	s.Ledger[t.To] += t.Amount
}

// Returns true if caller has enough funds on account to pay for call
func (s *State) FundContractCall(callerAccount PublicKey, amount uint64, gas uint64) bool {
	if s.Ledger[callerAccount.String()] >= amount+gas {
		s.TotalStake -= gas
		s.Ledger[callerAccount.String()] -= amount + gas
		return true
	}
	return false
}

//Used to refund money from contracts back into the original user ledger
func (s *State) returnAmountFromContracts(callerAccount PublicKey, amount uint64) {
	s.Ledger[callerAccount.String()] += amount
}

// Goes through the contract accounts and check which is out of prepaid.
// These are deleted and a list of deleted contracts is returned
func (s *State) CleanContractLedger() []string {
	var expiredContracts []string
	for c := range s.ConAccounts {
		contract := s.ConAccounts[c]
		if contract.Prepaid <= 0 {
			expiredContracts = append(expiredContracts, c)
			s.returnAmountFromContracts(contract.Owner, s.ConStake[c])
			delete(s.ConAccounts, c)
		}
	}
	return expiredContracts
}

// Switches depending on type of Trans. Returns amount of gas used
func (s *State) HandleTransData(td TransData, transactionFee uint64) uint64 {
	switch td.getType() {

	case TRANSACTION:
		gasCost := s.AddTransaction(td.Transaction, transactionFee)
		return gasCost

	case CONTRACTCALL:
		contract := td.ContractCall
		// Transfer funds from caller to contract
		if !s.FundContractCall(contract.Caller, contract.Amount, contract.Gas) {
			return 0
		}

		// Runs contract at contract layer
		newContractLedger, transferList, remainingGas, callerr := smart.CallContract(contract.Address,
			contract.Entry, contract.Params, contract.Amount, contract.Gas)

		// Calc how much gas used and refund not used gas to caller
		gasUsed := contract.Gas - remainingGas
		s.PayBlockRewardOrRemainGas(contract.Caller, remainingGas)

		// If contract not successful, then return amount to caller
		if callerr != nil {
			s.returnAmountFromContracts(contract.Caller, contract.Amount)
			return gasUsed
		}
		// If contract succeeded, execute the transactions from the contract layer
		s.ConStake = newContractLedger
		for _, t := range transferList {
			s.AddContractTransaction(t)
		}
		return gasUsed

	case CONTRACTINIT:
		contractInit := td.ContractInit
		if s.Ledger[contractInit.Owner.String()] > contractInit.Prepaid+contractInit.Gas {
			addr, remainGas, storageCost, err := smart.InitiateContract(contractInit.Code, contractInit.Gas)
			s.PayBlockRewardOrRemainGas(contractInit.Owner, remainGas)
			if err != nil {
				return contractInit.Gas - remainGas
			} else {
				s.InitializeContractAccount(addr, contractInit.Owner, contractInit.Prepaid, storageCost)
				return contractInit.Gas - remainGas
			}
		}

	}
	return 0
}

// Returns the new State, cost of the Contract call and true if contract executed successful
func (s *State) handleContractCall(contract ContractCall) (uint64, error) {

	// Transfer funds from caller to contract
	if !s.FundContractCall(contract.Caller, contract.Amount, contract.Gas) {
		return 0, fmt.Errorf("not enough money to pay for contract call")
	}

	// Runs contract at contract layer
	newContractLedger, transactionList, remainingGas, err := smart.CallContract(contract.Address, contract.Entry, contract.Params,
		contract.Amount, contract.Gas)
	// this number can't be negative, it is checked in smart contract layer
	gasUsed := contract.Gas - remainingGas

	if err != nil {
		// If contract not successful, then return remaining funds to caller
		s.PayBlockRewardOrRemainGas(contract.Caller, contract.Amount+remainingGas)
		return gasUsed, err
	} else {
		// If contract succeeded, execute the transactions from the contract layer
		s.ConStake = newContractLedger
		for _, t := range transactionList {
			s.AddContractTransaction(t)
		}
		s.PayBlockRewardOrRemainGas(contract.Caller, remainingGas)
		return gasUsed, nil
	}

}

// checks all contract accounts and withdraw the storage cost from their prepaid.
// Takes as argument how many slots to pay for and then returns total amount of costs for all contracts
func (s *State) CollectStorageCost(slots uint64) uint64 {
	accumulatedStorageCosts := uint64(0)

	for acc := range s.ConAccounts {
		account := s.ConAccounts[acc]
		storageCost := min(account.StorageCost*slots, account.Prepaid)
		account.Prepaid -= storageCost
		s.ConAccounts[acc] = account
		accumulatedStorageCosts += storageCost
	}

	return accumulatedStorageCosts
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
