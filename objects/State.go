package objects

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	"sort"
	"strconv"
)

type State struct {
	Ledger     map[PublicKey]int
	ParentHash string
}

func NewInitialState(key PublicKey) State {
	ledger := make(map[PublicKey]int)
	ledger[key] = 1000000 // 1 mil
	return State{ledger, ""}
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
	if s.Ledger[t.From] < amountWithFees {
		fmt.Println("Not enough money on senders account")
		return false
	}

	s.Ledger[t.To] += t.Amount
	s.Ledger[t.From] -= amountWithFees
	return true
}

func (s *State) AddBlockRewardAndTransFees(pk PublicKey, reward int) {
	s.Ledger[pk] += reward
}

func (s State) StateAsString() string {
	var buf bytes.Buffer

	sortedLedger := make(map[string]int)

	keys := make([]string, 0, len(s.Ledger))
	for k := range s.Ledger {
		sortedLedger[k.String()] = s.Ledger[k]
		keys = append(keys, k.String())
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(strconv.Itoa(sortedLedger[k]))
	}
	buf.WriteString(s.ParentHash)

	return buf.String()
}

func (s *State) CreateStateHash(sk SecretKey) string {

	return Sign(HashSHA(s.StateAsString()), sk)

}

func (s State) VerifyStateHash(sig string, pk PublicKey) bool {

	return Verify(HashSHA(s.StateAsString()), sig, pk)
}
