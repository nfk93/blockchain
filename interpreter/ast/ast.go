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
)

var typeCodeToString = map[Type]string{
	STRING: "string",
	INT:    "int",
	FLOAT:  "float"}

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
