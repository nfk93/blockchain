package objects

type ContractCall struct {
	Call      string
	Entry     string
	Params    Parameter
	Amount    int
	Gas       int
	Address   string
	Nonce     string
	Signature string
}

type Operation interface {
}

type Parameter interface {
}

type Storage interface {
}
