package ast

type BinOper int

const (
	PLUS BinOper = iota
	MINUS
	TIMES
	DIVIDE
	EQ
	NEQ
	GEQ
	LEQ
	LT
	GT
	AND
	OR
)

type UnOper int

const (
	UNARYMINUS UnOper = iota
	NOT
)

func binOperToString(op BinOper) string {
	switch op {
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case TIMES:
		return "TIMES"
	case DIVIDE:
		return "DIVIDE"
	case EQ:
		return "EQ"
	case NEQ:
		return "NEQ"
	case GEQ:
		return "GEQ"
	case LEQ:
		return "LEQ"
	case LT:
		return "LT"
	case GT:
		return "GT"
	case AND:
		return "AND"
	case OR:
		return "OR"
	}
	return "ERROR"
}

func unOperToString(op UnOper) string {
	switch op {
	case UNARYMINUS:
		return "MINUS"
	case NOT:
		return "NOT"
	default:
		return "ERROR"
	}
}
