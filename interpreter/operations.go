package interpreter

type Operation interface {
	OperationID() OperationID
	OperationData() interface{}
}

type OperationID int

const (
	FAILWITH = iota
	TRANSFER
	CONTRACTCALL
)

// --------------------

type FailWith struct {
	msg string
}

func (o FailWith) OperationID() OperationID {
	return FAILWITH
}

func (o FailWith) OperationData() interface{} {
	return o.msg
}

type Transfer struct {
	data TransferData
}
type TransferData struct {
	key    string
	amount uint64
}

func (o Transfer) OperationID() OperationID {
	return TRANSFER
}

func (o Transfer) OperationData() interface{} {
	return o.data
}

type ContractCall struct {
	data CallData
}

type CallData struct {
	address string
	gas     float64
	params  Value
}

func (o ContractCall) OperationData() interface{} {
	return o.data
}

func (o ContractCall) OperationID() OperationID {
	return CONTRACTCALL
}
