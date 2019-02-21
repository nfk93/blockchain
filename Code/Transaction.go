package Code

type Transaction struct {
	From      PublicKey
	To        PublicKey
	Amount    int
	ID        string
	Signature string
}
