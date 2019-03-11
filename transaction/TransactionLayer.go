package transaction

import (
	"fmt"
	. "github.com/nfk93/blockchain/objects"
)

type TLNode struct {
	block Block
	state State
}

type State struct {
	ledger     map[string]int
	parentHash string
}

func StartTransactionLayer(blockInput chan Block, stateReturn chan State) {
	go func() {
		for {
			b := <-blockInput
			stateReturn <- processBlock(b)
		}
	}()
}

func processBlock(b Block) State {
	s := State{}
	s.ledger = map[string]int{}

	for _, t := range b.BlockData.Trans {
		s.addTransaction(t)
	}
	fmt.Println(s.ledger)
	return s
}

func (s *State) addTransaction(t Transaction) {
	s.ledger[t.To.String()] += t.Amount
	s.ledger[t.From.String()] -= t.Amount
}
