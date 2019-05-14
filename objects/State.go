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

func (s *State) AddTransaction(t Transaction, transFee int) bool {
	//TODO: Handle checks of legal transactions

	amountWithFees := t.Amount + transFee

	if !t.VerifyTransaction() {
		fmt.Println("The transactions didn't verify", t)
		return false
	}
	if t.Amount <= 0 {
		fmt.Println("Invalid transaction Amount! Amount should be positive!", t.Amount)
		return false
	}

	// Sender has to be able to pay both the amount and the fee
	if s.Ledger[t.From.String()] < amountWithFees {
		fmt.Println("Not enough money on senders account")
		return false
	}

	s.Ledger[t.From.String()] -= amountWithFees
	s.Ledger[t.To.String()] += t.Amount
	s.TotalStake -= transFee // Take the fee out of the system
	return true
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
