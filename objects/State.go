package objects

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	"sort"
	"strconv"
)

type ContractAccount struct {
	Owner   PublicKey
	Prepaid int
}

type State struct {
	Ledger      map[string]int
	ConStake    map[string]int
	ConAccounts map[string]ContractAccount
	ParentHash  string
	TotalStake  int
}

func NewInitialState(key PublicKey) State {
	initialStake := 1000000 // 1 mil
	ledger := make(map[string]int)
	conStake := make(map[string]int)
	conledger := make(map[string]ContractAccount)
	ledger[key.String()] = initialStake
	return State{ledger, conStake, conledger, "", initialStake}
}

func (s *State) AddTransaction(t Transaction, transFee int) {
	//TODO: Handle checks of legal transactions

	amountWithFees := t.Amount + transFee

	if !t.VerifyTransaction() {
		fmt.Println("The transactions didn't verify", t)
		return
	}
	if t.Amount <= 0 {
		fmt.Println("Invalid transaction Amount! Amount should be positive!", t.Amount)
		return
	}

	// Sender has to be able to pay both the amount and the fee
	if s.Ledger[t.From.String()] < amountWithFees {
		fmt.Println("Not enough money on senders account")
		return
	}

	s.Ledger[t.From.String()] -= amountWithFees
	s.Ledger[t.To.String()] += t.Amount
	s.TotalStake -= transFee // Take the fee out of the system

}

func (s *State) AddBlockReward(pk PublicKey, reward int) {
	s.Ledger[pk.String()] += reward
	s.TotalStake += reward // putting back the fees and an block reward if anyone claim it
}

func (s State) toString() string {
	var buf bytes.Buffer

	sortedLedger := make(map[string]int)

	keys := make([]string, 0, len(s.Ledger))
	for k := range s.Ledger {
		sortedLedger[k] = s.Ledger[k]
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(strconv.Itoa(sortedLedger[k]))
	}
	buf.WriteString(s.ParentHash)
	buf.WriteString(strconv.Itoa(s.TotalStake))

	return buf.String()
}

func (s *State) SignHashedState(sk SecretKey) string {
	return Sign(HashSHA(s.toString()), sk)
}

func (s State) VerifyHashedState(sig string, pk PublicKey) bool {
	return Verify(HashSHA(s.toString()), sig, pk)
}

func (s *State) InitializeContract(addr string, owner PublicKey, prepaid int) {
	s.ConAccounts[addr] = ContractAccount{owner, prepaid}
}

// Used for putting more prepaid on an contract account
func (s *State) PrepayContracts(addr string, amount int) {
	newBalance := s.ConAccounts[addr]
	newBalance.Prepaid += amount
	s.ConAccounts[addr] = newBalance
}

// Used for handling contract layer transaction to users
func (s *State) AddContractTransaction(t ContractTransaction) {
	s.Ledger[t.To.String()] += t.Amount
}

// Returns true if caller has enough funds on account to pay for call
func (s *State) FundContractCall(callerAccount PublicKey, amount int) bool {
	if s.Ledger[callerAccount.String()] >= amount {
		s.Ledger[callerAccount.String()] -= amount
		return true
	}
	return false
}

//Used to refund money from contracts back into the original user ledger
func (s *State) RefundContractCall(callerAccount PublicKey, amount int) {
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
			s.RefundContractCall(contract.Owner, s.ConStake[c])
			delete(s.ConAccounts, c)
		}
	}
	return expiredContracts
}

// Switches depending on type of Trans. Returns amount of gas used
func (s *State) HandleTransData(td TransData, transactionFee int) int {
	switch td.(type) {

	case Transaction:
		td := td.(Transaction)
		s.AddTransaction(td, transactionFee)
		return transactionFee

	case ContractCall:
		td := td.(ContractCall)
		// Transfer funds from caller to contract
		if !s.FundContractCall(td.Caller, td.Amount+td.Gas) {
			return transactionFee // TODO price for funding contract? essentially just a transfer of money -> Transfee
		}

		// Runs contract at contract layer
		callSuccess, newContractStake, transactionList, remainingGas := CallAtConLayer(td)

		// Calc how much gas used and refund not used gas to caller
		gasUsed := td.Gas - remainingGas
		s.RefundContractCall(td.Caller, remainingGas)

		// If contract not successful, then return amount to caller
		if !callSuccess {
			s.RefundContractCall(td.Caller, td.Amount)
			return gasUsed
		}

		// If contract succeeded, execute the transactions from the contract layer
		s.ConStake = newContractStake
		for _, t := range transactionList {
			s.AddContractTransaction(t)
		}
		return gasUsed

	case ContractInitialize:
		td := td.(ContractInitialize)
		addr, remainGas, success := InitContractAtConLayer(td.Code, td.Gas)
		s.RefundContractCall(td.Owner, remainGas)
		if success {
			s.InitializeContract(addr, td.Owner, td.Prepaid)
		}
		return td.Gas - remainGas
	}
	return 0
}

// Returns the new State, cost of the Contract call and true if contract executed successful
func (s *State) handleContractCall(contract ContractCall) (int, bool) {

	// Transfer funds from caller to contract
	if !s.FundContractCall(contract.Caller, contract.Amount+contract.Gas) {
		return 0, false
	}

	// Runs contract at contract layer
	callSuccess, newContractStake, transactionList, remainingGas := CallAtConLayer(contract)
	gasUsed := contract.Gas - remainingGas

	// If contract succeeded, execute the transactions from the contract layer
	if callSuccess {
		s.ConStake = newContractStake
		for _, t := range transactionList {
			s.AddContractTransaction(t)
		}
		s.RefundContractCall(contract.Caller, remainingGas)
		return gasUsed, callSuccess
	}

	// If contract not successful, then return remaining funds to caller
	s.RefundContractCall(contract.Caller, contract.Amount+remainingGas)
	return gasUsed, callSuccess

}
