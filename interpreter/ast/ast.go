package ast

import (
	"fmt"
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

/* Entry */
type Entry struct {
	id string
}

/* Top Level */
type TopLevel struct {
	Roots []Exp
}

func NewRoot(e interface{}) (Exp, error) {
	switch e.(type) {
	case SimpleTypeDecl:
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
