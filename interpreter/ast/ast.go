package ast

import (
	"fmt"
	"github.com/pkg/errors"
)

type Exp interface {
	String() string
}

type Type int

const (
	// Types
	STRING Type = iota
	INT
	FLOAT
	KEYHASH
)

var typeCodeToString = map[Type]string{
	STRING:  "string",
	INT:     "int",
	FLOAT:   "float",
	KEYHASH: "keyhash"}

type TypeOption struct {
	opt bool
	typ Type
}

type Operation int

const (
	PLUS Operation = iota
	MINUS
	TIMES
	DIVIDE
)

type BinOPExp struct {
	left  interface{}
	right interface{}
	oper  Operation
}

func NewBinOpExp(left, right, oper Operation) (BinOPExp, error) {
	return BinOPExp{left, right, oper}, nil
}

type TodoExp struct{}

func NewTodoExp() (Exp, error) {
	return TodoExp{}, nil
}

func (e TodoExp) String() string {
	return "TODOEXP"
}

/* Simple Type Declaration */
type SimpleTypeDecl struct {
	id  string
	typ Type
}

func NewSimpleTypeDecl(id string, typ interface{}) (Exp, error) {
	return SimpleTypeDecl{id, typ.(Type)}, nil
}

func (e SimpleTypeDecl) String() string {
	return fmt.Sprintf("SimpleTypeDecl(id: %s, typ: %s)", e.id, typeCodeToString[e.typ])
}

/* Top Level */
type TopLevel struct {
	Roots []Exp
}

func NewRoot(e interface{}) (Exp, error) {
	switch e.(type) {
	case SimpleTypeDecl, EntryExpression:
		return TopLevel{[]Exp{e.(Exp)}}, nil
	default:
		ex, _ := fail(fmt.Sprintf("Toplevel error, New root can't be type %T", e))
		return TopLevel{[]Exp{ex.(Exp)}}, nil
	}
}

func (e TopLevel) String() string {
	str := fmt.Sprint("TopLevel([")
	for _, v := range e.Roots {
		str = str + "\n\t" + v.String()
	}
	return str + "\n])"
}

func AppendRoots(e1, e2 interface{}) (Exp, error) {
	switch e2.(type) {
	case TopLevel:
		e2 := e2.(TopLevel)
		return TopLevel{append([]Exp{e1.(Exp)}, e2.Roots...)}, nil
	default:
		return fail(fmt.Sprintf("TopLevel is not a TopLevel expression"))
	}
}

/* New Entry */
type EntryExpression struct {
	Id      string
	Params  Pattern
	Storage Pattern
	Body    Exp
}

func (e EntryExpression) String() string {
	return fmt.Sprintf("EntryExpression(id: %s, params: %s, storage: %s, body: %s)", e.Id, e.Params.String(),
		e.Storage.String(), e.Body.String())
}

func NewEntryExpression(id string, params, pattern, body interface{}) (Exp, error) {
	return EntryExpression{id, params.(Pattern), pattern.(Pattern), body.(Exp)}, nil
}

/* Pattern */
type Param struct {
	id   string
	anno TypeOption
}

func (p Param) String() string {
	if p.anno.opt {
		return fmt.Sprintf("(%s: %s)", p.id, typeCodeToString[p.anno.typ])
	} else {
		return fmt.Sprintf("%s", p.id)
	}
}

func NewParam(id string) (Param, error) {
	return Param{id, TypeOption{false, -1}}, nil
}

func NewAnnoParam(id string, anno interface{}) (Param, error) {
	return Param{id, TypeOption{true, anno.(Type)}}, nil
}

func AppendParams(par, pars interface{}) ([]Param, error) {
	params := pars.([]Param)
	param := par.(Param)
	return append([]Param{param}, params...), nil
}

func NewParamList(param interface{}) ([]Param, error) {
	return []Param{param.(Param)}, nil
}

type Pattern struct {
	params []Param
}

func NewPattern(params interface{}) (Pattern, error) {
	switch params.(type) {
	case []Param:
		return Pattern{params.([]Param)}, nil
	case Param:
		return Pattern{[]Param{params.(Param)}}, nil
	default:
		return Pattern{}, errors.Errorf("Can't derive pattern")
	}
}

func NewEmptyPattern() (Pattern, error) {
	return Pattern{}, nil
}

func (p Pattern) String() string {
	if len(p.params) == 0 {
		return "Pattern()"
	} else {
		res := "Pattern("
		var par Param
		params := p.params
		for len(params) > 1 {
			par, params = params[0], params[1:]
			res = res + fmt.Sprintf("%s, ", par.String())
		}
		par = params[0]
		res = res + fmt.Sprintf("%s)", par.String())
		return res
	}
}

/* KeyLit */
type KeyLit struct {
	key string
}

func (k KeyLit) String() string {
	return fmt.Sprintf("KeyLit(key: %s)", k.key)
}

func NewKeyLit(key []byte) (Exp, error) {
	actualkey := string(key)[3:]
	if checkKey(key) {
		return KeyLit{actualkey}, nil
	} else {
		err := "key is not valid"
		return ErrorExpression{err}, errors.Errorf(err)
	}
}

func checkKey(key []byte) bool {
	// TODO: check if this is an actual base 58 key, i.e. has the right characters and such
	return true
}

/* BoolLit */
type BoolLit struct {
	val bool
}

func (b BoolLit) String() string {
	return fmt.Sprintf("BoolLit(val: %t)", b.val)
}

func NewBoolLit(val bool) (Exp, error) {
	return BoolLit{val}, nil
}

/* IntLit */
type IntLit struct {
	val int64
}

func (i IntLit) String() string {
	return fmt.Sprintf("IntLit(val: %d)", i.val)
}

func NewIntLit(val int64) (Exp, error) {
	return IntLit{val}, nil
}

/* FloatLit */
type FloatLit struct {
	val float64
}

func (f FloatLit) String() string {
	return fmt.Sprintf("FloatLit(val: %e)", f.val)
}

func NewFloatLit(val float64) (Exp, error) {
	return FloatLit{val}, nil
}

/* Koin Lit */
type KoinLit struct {
	val int64
}

func (k KoinLit) String() string {
	return fmt.Sprintf("KoinLit(val: %d)", k.val)
}

func NewKoinLit(koins int64) (Exp, error) {
	if koins <= 0 {
		err := "koin literal can't have negative value"
		return ErrorExpression{err}, nil
	} else {
		return KoinLit{koins}, nil
	}
}

/* String Lit */
type StringLit struct {
	val string
}

func (s StringLit) String() string {
	return fmt.Sprintf("StringLit(val: %s)", s.val)
}

func NewStringLit(str string) (Exp, error) {
	return StringLit{str}, nil
}

/* Unit Lit */
type UnitLit struct{}

func (u UnitLit) String() string {
	return fmt.Sprintf("UnitLit")
}

func NewUnitLit() (Exp, error) {
	return UnitLit{}, nil
}

/* List Lit TODO: Refactor to use head and tail instead? */
type ListLit struct {
	typ  Type
	list []Exp
}

func (l ListLit) String() string {
	if len(l.list) == 0 {
		return fmt.Sprintf("ListLit(list: [], typ: %s)", typeCodeToString[l.typ])
	} else {
		res := "ListLit(list: "
		var e Exp
		list := l.list
		for len(list) > 1 {
			e, list = list[0], list[1:]
			res = res + fmt.Sprintf("%s, ", e.String())
		}
		e = list[0]
		res = res + fmt.Sprintf("%s)", e.String())
		return res
	}
}

// TODO: derive types
func NewListLit(exp interface{}) (Exp, error) {
	return ListLit{-1, //TODO
		[]Exp{exp.(Exp)}}, nil
}

func NewEmptyList() (Exp, error) {
	return ListLit{-1, //TODO
		[]Exp{}}, nil
}

// TODO: check that listtype matches exp type
func AppendList(exp1, exp2 interface{}) (Exp, error) {
	lst1 := exp1.(ListLit)
	lst2 := exp2.(ListLit)
	return ListLit{lst1.typ, append(lst1.list, lst2.list...)}, nil
}

func PrependList(exp, list interface{}) (Exp, error) {
	return TodoExp{}, nil
}

// ---------------------------

func notImplemented() *notImplementedError { return &notImplementedError{} }

type notImplementedError struct{}

func (e *notImplementedError) Error() string {
	return fmt.Sprint("Not Implemented!")
}

type ErrorExpression struct {
	errormsg string
}

func (e ErrorExpression) String() string {
	return fmt.Sprintf("ErrorExpression(\"%s\")", e.errormsg)
}
func fail(str string) (Exp, error) {
	return ErrorExpression{str}, nil
}
