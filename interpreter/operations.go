package interpreter

type Operation interface {
	OperationID() OperationID
	OperationData() interface{}
}

type OperationID int

const (
	FAILWITH = iota
	TRANSFER
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
	amount int
}

func (o Transfer) OperationID() OperationID {
	return TRANSFER
}

func (o Transfer) OperationData() interface{} {
	return o.data
}
