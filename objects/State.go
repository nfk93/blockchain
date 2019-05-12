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
	Balance int
	Prepaid int
}

type State struct {
	Ledger     map[string]int
	ConLedger  map[string]ContractAccount
	ParentHash string
	TotalStake int
}

func NewInitialState(key PublicKey) State {
	initialStake := 1000000 // 1 mil
	ledger := make(map[string]int)
	conledger := make(map[string]ContractAccount)
	ledger[key.String()] = initialStake
	return State{ledger, conledger, "", initialStake}
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

//func (s *State)initContract(amount int) {
//
//	if !checkTransaction(*s, t, amountWithFees){
//		return false
//	}
//
//}
//
//func (s *State)refundUser(amount int ) {
//
//}
