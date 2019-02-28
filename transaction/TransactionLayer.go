package transaction

import . "github.com/nfk93/blockchain/objects"

var ledger = map[string]int{}

func ReceiveBlock(b Block) {
	processBlock(b)
}

func processBlock(b Block) {
	for _, t := range b.BlockData.Trans {
		addTransaction(t)
	}
}

func addTransaction(t Transaction) {
	ledger[t.To.String()] += t.Amount
	ledger[t.From.String()] -= t.Amount
}

func GetShare(s string) int {
	return ledger[s]
}

func GetLedger() map[string]int {
	return ledger
}

func BeRich(s string) {
	ledger[s] += 100000
}
