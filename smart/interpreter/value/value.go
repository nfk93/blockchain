package value

import "log"

const bitcost = uint64(1)
const charcost = uint64(32)
const addresscost = uint64(32) * charcost * bitcost

type Value = interface {
	Size() uint64
}

type StringVal struct {
	Value string
}

func (v StringVal) Size() uint64 {
	return stringSizeVal(v.Value)
}

type IntVal struct {
	Value int64
}

func (v IntVal) Size() uint64 {
	return 64 * bitcost
}

type NatVal struct {
	Value uint64
}

func (v NatVal) Size() uint64 {
	return 64 * bitcost
}

type AddressVal struct {
	Value string
}

func (v AddressVal) Size() uint64 {
	return addresscost
}

type KeyVal struct {
	Value string
}

func (v KeyVal) Size() uint64 {
	return addresscost
}

type BoolVal struct {
	Value bool
}

func (v BoolVal) Size() uint64 {
	return bitcost
}

type KoinVal struct {
	Value uint64
}

func (v KoinVal) Size() uint64 {
	return 64 * bitcost
}

type OperationVal struct {
	Value interface{}
}

func (v OperationVal) Size() uint64 {
	switch v.Value.(type) {
	case FailWith:
		op := v.Value.(FailWith)
		return stringSizeVal(op.Msg)
	case Transfer:
		return 64*bitcost + addresscost
	case ContractCall:
		op := v.Value.(ContractCall)
		return addresscost + 64*bitcost + stringSizeVal(op.Entry) + op.Params.Size()
	}
	return 0 // TODO
}

type OptionVal struct {
	Value Value
	Opt   bool
}

func (v OptionVal) Size() uint64 {
	return bitcost + v.Value.Size()
}

type ListVal struct {
	Values []Value
}

func (v ListVal) Size() uint64 {
	cost := 64 * bitcost
	if len(v.Values) > 0 {
		itemcost := v.Values[0].Size()
		cost += uint64(len(v.Values) * int(itemcost))
	}
	return cost
}

type UnitVal struct{}

func (v UnitVal) Size() uint64 {
	return bitcost
}

type TupleVal struct {
	Values []Value
}

func (v TupleVal) Size() uint64 {
	sum := 64 * bitcost
	for _, v := range v.Values {
		sum += v.Size()
	}
	return sum
}

type StructVal struct {
	Field map[string]Value
}

func (v StructVal) Size() uint64 {
	sum := 64 * bitcost
	for _, v := range v.Field {
		sum += v.Size()
	}
	return sum
}

type LambdaVal struct {
	Value ModuleLookup
}

func (v LambdaVal) Size() uint64 {
	return 0 // this can't occur outside the interpreter, and thus can't be in a storage
}

func stringSizeVal(s string) uint64 {
	return uint64(len(s)) * 32 * bitcost
}

type ModuleLookup int

const (
	CURRENT_BALANCE = iota
	CURRENT_AMOUNT
	CURRENT_GAS
	CURRENT_FAILWITH
	CONTRACT_CALL
	ACCOUNT_TRANSFER
	ACCOUNT_DEFAULT
)

type Code int

const (
	STRING Code = iota
	INT
	KEY
	ADDRESS
	NAT
	BOOL
	KOIN
	LIST
	TUPLE
	STRUCT
	UNIT
	ERROR
)

func GetTypeCode(val Value) Code {
	switch val.(type) {
	case StringVal:
		return STRING
	case IntVal:
		return INT
	case KeyVal:
		return KEY
	case AddressVal:
		return ADDRESS
	case NatVal:
		return NAT
	case KoinVal:
		return KOIN
	case BoolVal:
		return BOOL
	case ListVal:
		return LIST
	case TupleVal:
		return TUPLE
	case StructVal:
		return STRUCT
	case UnitVal:
		return UNIT
	default:
		return ERROR
	}
}

func Equals(val1, val2 Value) bool {
	code1 := GetTypeCode(val1)
	code2 := GetTypeCode(val2)
	if code1 != code2 {
		return false
	}
	switch code1 {
	case STRING:
		val1 := val1.(StringVal)
		val2 := val2.(StringVal)
		return val1.Value == val2.Value
	case INT:
		val1 := val1.(IntVal)
		val2 := val2.(IntVal)
		return val1.Value == val2.Value
	case KEY:
		val1 := val1.(KeyVal)
		val2 := val2.(KeyVal)
		return val1.Value == val2.Value
	case BOOL:
		val1 := val1.(BoolVal)
		val2 := val2.(BoolVal)
		return val1.Value == val2.Value
	case KOIN:
		val1 := val1.(KoinVal)
		val2 := val2.(KoinVal)
		return val1.Value == val2.Value
	case UNIT:
		return true
	case NAT:
		val1 := val1.(NatVal)
		val2 := val2.(NatVal)
		return val1.Value == val2.Value
	case ADDRESS:
		val1 := val1.(AddressVal)
		val2 := val2.(AddressVal)
		return val1.Value == val2.Value
	case LIST:
		val1 := val1.(ListVal)
		val2 := val2.(ListVal)
		if len(val1.Values) != len(val2.Values) {
			return false
		}
		for i, v := range val1.Values {
			if !Equals(v, val2.Values[i]) {
				return false
			}
		}
		return true
	case TUPLE:
		val1 := val1.(TupleVal)
		val2 := val2.(TupleVal)
		if len(val1.Values) != len(val2.Values) {
			return false
		}
		for i, v := range val1.Values {
			if !Equals(v, val2.Values[i]) {
				return false
			}
		}
		return true
	case STRUCT:
		val1 := val1.(StructVal)
		val2 := val2.(StructVal)
		if len(val1.Field) != len(val2.Field) {
			return false
		}
		for k1, v1 := range val1.Field {
			v2, exists := val2.Field[k1]
			if !exists {
				return false
			} else if !Equals(v1, v2) {
				return false
			}
		}
		return true
	default:
		log.Printf("can't compare values %s and %s", val1, val2)
		return false
	}
}
