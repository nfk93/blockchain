package interpreter

type Operation interface{}

type FailWith struct {
	Msg string
}

type Transfer struct {
	Key    string
	Amount uint64
}

type ContractCall struct {
	Address string
	Amount  uint64
	Entry   string
	Params  Value
}
