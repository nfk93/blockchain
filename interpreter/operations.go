package interpreter

type Operation interface {
	OperationID() OperationID
	OperationData() interface{}
}

type OperationID int

const (
	FAILWITH = iota
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
