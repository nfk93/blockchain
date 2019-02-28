package objects

import . "github.com/nfk93/blockchain/crypto"

type Transaction struct {
	From      PublicKey
	To        PublicKey
	Amount    int
	ID        string
	Signature string
}
