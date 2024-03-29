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
		"key_lit",
		"add_lit",
		"true",
		"false",
		"nat_lit",
		"koin_lit",
		"int_lit",
		"string_lit",
		"()",
		"(",
		")",
		",",
		"[",
		"]",
		"[]",
		";",
		"{",
		"}",
		"lident",
		"=",
	},

	idMap: map[string]Type{
		"INVALID":    0,
		"$":          1,
		"key_lit":    2,
		"add_lit":    3,
		"true":       4,
		"false":      5,
		"nat_lit":    6,
		"koin_lit":   7,
		"int_lit":    8,
		"string_lit": 9,
		"()":         10,
		"(":          11,
		")":          12,
		",":          13,
		"[":          14,
		"]":          15,
		"[]":         16,
		";":          17,
		"{":          18,
		"}":          19,
		"lident":     20,
		"=":          21,
	},
}
