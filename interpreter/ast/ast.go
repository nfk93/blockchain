package ast

import "fmt"

type Expression interface {
	children() []Expression
}

type Texpression interface {
	exp() Expression
	typ() Type
}

type Type int

const (
	// Types
	INT Type = iota
	STRING
	FLOAT
)

type Operation int

const (
	PLUS Operation = iota
)

type BinOPExp struct {
	c    []Expression
	oper Operation
}

func NewBinOpExp(children []Expression, oper Operation) BinOPExp {
	return BinOPExp{
		children,
		oper}
}

func (e BinOPExp) children() []Expression {
	return e.c
}

type TodoExp struct{}

func NewTodoExp() TodoExp {
	return TodoExp{}
}

func (e TodoExp) children() []Expression {
	return []Expression{}
}

func PrintExpression(exp Expression) {
	switch exp.(type) {
	case BinOPExp:
		exp := exp.(BinOPExp)
		fmt.Println(exp.oper)
	}
}
