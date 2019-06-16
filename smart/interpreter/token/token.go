// Code generated by gocc; DO NOT EDIT.

package token

import (
	"fmt"
)

type Token struct {
	Type
	Lit []byte
	Pos
}

type Type int

const (
	INVALID Type = iota
	EOF
)

type Pos struct {
	Offset int
	Line   int
	Column int
}

func (p Pos) String() string {
	return fmt.Sprintf("Pos(offset=%d, line=%d, column=%d)", p.Offset, p.Line, p.Column)
}

type TokenMap struct {
	typeMap []string
	idMap   map[string]Type
}

func (m TokenMap) Id(tok Type) string {
	if int(tok) < len(m.typeMap) {
		return m.typeMap[tok]
	}
	return "unknown"
}

func (m TokenMap) Type(tok string) Type {
	if typ, exist := m.idMap[tok]; exist {
		return typ
	}
	return INVALID
}

func (m TokenMap) TokenString(tok *Token) string {
	//TODO: refactor to print pos & token string properly
	return fmt.Sprintf("%s(%d,%s)", m.Id(tok.Type), tok.Type, tok.Lit)
}

func (m TokenMap) StringType(typ Type) string {
	return fmt.Sprintf("%s(%d)", m.Id(typ), typ)
}

var TokMap = TokenMap{
	typeMap: []string{
		"INVALID",
		"$",
		"letinit",
		"lident",
		"eq",
		"letentry",
		"type",
		"lbrace",
		"rbrace",
		"colon",
		"semicolon",
		"if",
		"then",
		"else",
		"concat",
		"let",
		"in",
		"uident",
		"dot",
		"lparen",
		"rparen",
		"larrow",
		"or",
		"and",
		"plus",
		"minus",
		"ast",
		"slash",
		"neq",
		"geq",
		"leq",
		"lt",
		"gt",
		"unminus",
		"not",
		"comma",
		"bool",
		"int",
		"nat",
		"unit",
		"koin",
		"string",
		"key",
		"operation",
		"address",
		"option",
		"list",
		"key_lit",
		"address_lit",
		"true",
		"false",
		"int_lit",
		"nat_lit",
		"koin_lit",
		"string_lit",
		"lbrack",
		"rbrack",
	},

	idMap: map[string]Type{
		"INVALID":     0,
		"$":           1,
		"letinit":     2,
		"lident":      3,
		"eq":          4,
		"letentry":    5,
		"type":        6,
		"lbrace":      7,
		"rbrace":      8,
		"colon":       9,
		"semicolon":   10,
		"if":          11,
		"then":        12,
		"else":        13,
		"concat":      14,
		"let":         15,
		"in":          16,
		"uident":      17,
		"dot":         18,
		"lparen":      19,
		"rparen":      20,
		"larrow":      21,
		"or":          22,
		"and":         23,
		"plus":        24,
		"minus":       25,
		"ast":         26,
		"slash":       27,
		"neq":         28,
		"geq":         29,
		"leq":         30,
		"lt":          31,
		"gt":          32,
		"unminus":     33,
		"not":         34,
		"comma":       35,
		"bool":        36,
		"int":         37,
		"nat":         38,
		"unit":        39,
		"koin":        40,
		"string":      41,
		"key":         42,
		"operation":   43,
		"address":     44,
		"option":      45,
		"list":        46,
		"key_lit":     47,
		"address_lit": 48,
		"true":        49,
		"false":       50,
		"int_lit":     51,
		"nat_lit":     52,
		"koin_lit":    53,
		"string_lit":  54,
		"lbrack":      55,
		"rbrack":      56,
	},
}
