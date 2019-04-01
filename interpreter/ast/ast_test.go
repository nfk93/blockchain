package ast

import "testing"

func TestSmoke(t *testing.T) {
	binopexp := NewBinOpExp([]Expression{}, PLUS)
	PrintExpression(binopexp)
}
