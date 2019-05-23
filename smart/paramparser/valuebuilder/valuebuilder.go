package valuebuilder

import (
	"fmt"
	. "github.com/nfk93/blockchain/smart/interpreter/value"
	"github.com/nfk93/blockchain/smart/paramparser/token"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"
)

func NewBoolVal(val bool) (Value, error) {
	return BoolVal{val}, nil
}

func NewKeyVal(v interface{}) (Value, error) {
	key, err := ParseKey(v)
	return KeyVal{key}, err
}

func NewAddressVal(v interface{}) (Value, error) {
	add, err := ParseAddress(v)
	return AddressVal{add}, err
}

func NewNatVal(v interface{}) (Value, error) {
	val, err := ParseNat(v)
	return NatVal{val}, err
}

func NewIntVal(v interface{}) (Value, error) {
	i, err := ParseInt(v)
	return IntVal{i}, err
}

func NewStringVal(v interface{}) (Value, error) {
	str := ParseString(v)
	return StringVal{str}, nil
}

func NewKoinVal(v interface{}) (Value, error) {
	val, err := ParseKoin(v)
	return KoinVal{val}, err
}

func NewUnitVal() (Value, error) {
	return UnitVal{}, nil
}

func NewEmptyListVal() (Value, error) {
	vals := make([]Value, 0)
	return ListVal{vals}, nil
}

func NewListVal(v interface{}) (Value, error) {
	val, ok := v.(Value)
	if !ok {
		return nil, fmt.Errorf("encountered error when parsing values in list")
	}
	return ListVal{[]Value{val}}, nil
}

func ConcatListVal(list_, val_ interface{}) (Value, error) {
	list, ok1 := list_.(ListVal)
	if !ok1 {
		return nil, fmt.Errorf("can't parse list")
	}
	val, ok2 := val_.(Value)
	if !ok2 {
		return nil, fmt.Errorf("can't parse value to concatenate")
	}
	if !checkValueTypesEqual(list.Values[0], val) {
		return nil, fmt.Errorf("list types doesn't match type of value to be concatenated")
	} else {
		return ListVal{append(list.Values, val)}, nil
	}
}

func NewStructVal(ident, val interface{}) (Value, error) {
	id := ParseId(ident)
	m := make(map[string]Value)
	v, ok := val.(Value)
	if !ok {
		return nil, fmt.Errorf("struct field %s has unknown value", id)
	} else {
		m[id] = v
		return StructVal{m}, nil
	}
}

func AddStructField(struc, ident, val interface{}) (Value, error) {
	s, ok := struc.(StructVal)
	if !ok {
		return nil, fmt.Errorf("trying to add field to a non-struct value")
	}
	id := ParseId(ident)
	v, ok := val.(Value)
	if !ok {
		return nil, fmt.Errorf("struct field %s has unknown value", id)
	} else if _, exists := s.Field[id]; exists {
		return nil, fmt.Errorf("struct field %s is already defined", id)
	} else {
		s.Field[id] = v
		return s, nil
	}
}

func NewTupleVal(v1, v2 interface{}) (Value, error) {
	val1, ok1 := v1.(Value)
	val2, ok2 := v2.(Value)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("unknown value encountered in tuple")
	}
	return TupleVal{[]Value{val1, val2}}, nil
}

func AddTupleEntry(tup, v interface{}) (Value, error) {
	tuple, ok := tup.(TupleVal)
	if !ok {
		return nil, fmt.Errorf("trying to add entry to non-tuple value")
	}
	val, ok := v.(Value)
	if !ok {
		return nil, fmt.Errorf("unknown value encountered in tuple")
	}
	return TupleVal{append(tuple.Values, val)}, nil
}

func ParseId(id interface{}) string {
	return string(id.(*token.Token).Lit)
}

func ParseKey(key interface{}) (string, error) {
	return removePrefixAndCheckLen(key)
}

func ParseAddress(add interface{}) (string, error) {
	return removePrefixAndCheckLen(add)
}

func removePrefixAndCheckLen(str interface{}) (string, error) {
	str = string(str.(*token.Token).Lit)[3:]
	if utf8.RuneCountInString(str.(string)) != 32 {
		return "", fmt.Errorf("keys and addresses must have length 32")
	} else {
		return str.(string), nil
	}
}

func ParseInt(i interface{}) (int64, error) {
	integer, err := strconv.ParseInt(string(i.(*token.Token).Lit), 10, 64)
	return integer, err
}

func ParseString(str interface{}) string {
	s := string(str.(*token.Token).Lit)[1:]
	s = s[:(len(s) - 1)]
	return s
}

func ParseKoin(kn interface{}) (uint64, error) {
	knstr := string(kn.(*token.Token).Lit)
	knstr = knstr[:len(knstr)-2]
	i := strings.Index(knstr, ".")
	if i > -1 {
		decimals := len(knstr) - i - 1
		if decimals > 5 {
			return 0, fmt.Errorf("too many decimals in koin literal, max is 5")
		} else {
			knstr = knstr + strings.Repeat("0", 5-decimals)
			val, _ := strconv.ParseUint(strings.Replace(knstr, ".", "", -1), 10, 64)
			return val, nil
		}
	} else {
		val, _ := strconv.ParseUint(knstr, 10, 64)
		return val * 100000, nil
	}
}

func ParseNat(i interface{}) (uint64, error) {
	str := string(i.(*token.Token).Lit)
	val, err := strconv.ParseUint(str[:(len(str)-1)], 10, 64)
	return val, err
}

func checkValueTypesEqual(val1, val2 Value) bool {
	code1 := GetTypeCode(val1)
	code2 := GetTypeCode(val2)
	switch code1 {
	case STRING, INT, KEY, BOOL, KOIN, UNIT, NAT, ADDRESS:
		return code1 == code2
	case LIST:
		switch code2 {
		case LIST:
			val1 := val1.(ListVal)
			val2 := val2.(ListVal)
			return checkValueTypesEqual(val1.Values[0], val2.Values[0])
		default:
			return false
		}
	case TUPLE:
		switch code2 {
		case TUPLE:
			equal := true
			val1 := val1.(TupleVal)
			val2 := val2.(TupleVal)
			for i, v := range val1.Values {
				equal = equal && checkValueTypesEqual(v, val2.Values[i])
			}
			return equal
		default:
			return false
		}
	case STRUCT:
		switch code2 {
		case STRUCT:
			val1 := val1.(StructVal)
			val2 := val2.(StructVal)
			if len(val1.Field) != len(val2.Field) {
				return false
			}
			equal := true
			for k1, v1 := range val1.Field {
				v2, exists := val2.Field[k1]
				if !exists {
					return false
				}
				equal = equal && checkValueTypesEqual(v1, v2)
			}
			return equal
		default:
			return false
		}
	default:
		log.Println("checkTypesEqual case not matched")
		return false
	}
}
