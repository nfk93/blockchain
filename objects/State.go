package objects

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
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

func (s *State) AddTransaction(t Transaction) {
	//TODO: Handle checks of legal transactions

	if !t.VerifyTransaction() {
		fmt.Println("The transactions didn't verify", t)
		return
	}
	if t.Amount <= 0 {
		fmt.Println("Invalid transaction Amount! Amount should be positive!", t.Amount)
		return
	}

	//if s.ledger[t.From] < t.Amount { //TODO: remove comment such that it checks the balance
	//	fmt.Println("Not enough money on senders account")
	//	return
	//}
	s.Ledger[t.To] += t.Amount
	s.Ledger[t.From] -= t.Amount
}

func (s State) StateAsString() string {
	var buf bytes.Buffer
	for _, account := range s.Ledger {
		buf.WriteString(string(account))
	}
	buf.WriteString(s.ParentHash)

	return buf.String()
}
