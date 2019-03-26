package lexer

import (
	"github.com/nfk93/blockchain/interpreter/token"
	"io/ioutil"
	"testing"
)

func read_file(filepath string, t *testing.T) []byte {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Error("Error reading testfile:", filepath)
	}
	return dat
}

func compare_tokens(t *testing.T, strings []string, lex *Lexer) {
	l := len(strings)
	correct_tokens := make([]token.Type, l)
	for idx, str := range strings {
		correct_tokens[idx] = token.TokMap.Type(str)
	}
	i := 0
	for {
		nextToken := lex.Scan()
		if i >= l {
			t.Error("\nWrong token at index", i,
				"\n\tGot token:", nextToken.Type, token.TokMap.Id(nextToken.Type), nextToken.String(),
				"\n\tExpected: ", "no more tokens")
			break
		}
		if nextToken.Type == token.TokMap.Type("$") {
			if correct_tokens[i] != token.TokMap.Type("$") {
				t.Error("\nWrong token at index", i,
					"\n\tGot token:", nextToken.Type, token.TokMap.Id(nextToken.Type), nextToken.String(),
					"\n\tExpected: ", correct_tokens[i], token.TokMap.Id(correct_tokens[i]))
			}
			break
		} else if nextToken.Type != correct_tokens[i] {
			t.Error("\nWrong token at index", i,
				"\n\tGot token:", nextToken.Type, token.TokMap.Id(nextToken.Type), nextToken.String(),
				"\n\tExpected: ", correct_tokens[i], token.TokMap.Id(correct_tokens[i]))
		}
		i++
	}
}

func TestLexStorage(t *testing.T) {
	bytes := read_file("tests/storage_type", t)
	lex := NewLexer(bytes)

	strings := []string{
		TYPE, ID, EQ, LBRACE, ID, COLON, KEYHASH, SEMICOLON, ID, COLON, TEZ, SEMICOLON, ID, COLON, TEZ,
		SEMICOLON, ID, COLON, TEZ, SEMICOLON, RBRACE, EOF}
	compare_tokens(t, strings, lex)
}

func TestLexInitStorage(t *testing.T) {
	bytes := read_file("tests/init_storage", t)
	lex := NewLexer(bytes)

	strings := []string{
		LET, PERC, ID, ID, EQ, LBRACE, ID, EQ, HASH, SEMICOLON, ID, EQ, TEZ_LIT, SEMICOLON, ID, EQ, TEZ_LIT, SEMICOLON,
		ID, EQ, TEZ_LIT, SEMICOLON, RBRACE, EOF}
	compare_tokens(t, strings, lex)
}

func TestLexSimpleEntry(t *testing.T) {
	bytes := read_file("tests/simple_entry", t)
	lex := NewLexer(bytes)

	strings := []string{
		LET, PERC, ID, ID, LPAREN, ID, COLON, KEYHASH, RPAREN, LPAREN, ID, COLON, ID, RPAREN, EQ, IF, ID, DOT, ID, GEQ,
		ID, DOT, ID, THEN, LET, ID, EQ, ID, DOT, ID, LARROW, ID, IN, LPAREN, LPAREN, LBRACK, RBRACK, COLON, OPERATION,
		LIST, RPAREN, COMMA, ID, RPAREN, ELSE, LPAREN, LPAREN, LBRACK, RBRACK, COLON, OPERATION, LIST, RPAREN, COMMA,
		ID, RPAREN, EOF}
	compare_tokens(t, strings, lex)
}

func TestLexFloat(t *testing.T) {
	bytes := read_file("tests/float", t)
	lex := NewLexer(bytes)

	strings := []string{
		LET, ID, EQ, FLOAT, IN,
		LET, ID, EQ, FLOAT, IN,
		LET, ID, EQ, FLOAT, IN,
		LET, ID, EQ, INT, IN,
		LET, ID, EQ, FLOAT, IN,
		LET, ID, EQ, FLOAT, IN,
		LET, ID, EQ, ID, PLUS, ID, MINUS, ID, EOF}
	compare_tokens(t, strings, lex)
}

// TODO: make tests covering all of the below

const (
	COMMA     string = "comma"
	GEQ       string = "geq"
	GT        string = "gt"
	LEQ       string = "leq"
	LARROW    string = "larrow"
	NEQ       string = "neq"
	LT        string = "lt"
	RARROW    string = "rarrow"
	EQ        string = "eq"
	PLUS      string = "plus"
	MINUS     string = "minus"
	LBRACE    string = "lbrace"
	RBRACE    string = "rbrace"
	LBRACK    string = "lbrack"
	RBRACK    string = "rbrack"
	LPAREN    string = "lparen"
	RPAREN    string = "rparen"
	COLON     string = "colon"
	SEMICOLON string = "semicolon"
	KEYHASH   string = "keyhash"
	OPERATION string = "operation"
	LIST      string = "list"
	PERC      string = "percentage"
	LET       string = "let"
	IN        string = "in"
	IF        string = "if"
	THEN      string = "then"
	ELSE      string = "else"
	TYPE      string = "type"
	TEZ       string = "tez"
	HASH      string = "hash"
	ID        string = "id"
	STRING    string = "string"
	TEZ_LIT   string = "tez_lit"
	INT       string = "int"
	FLOAT     string = "float"
	DOT       string = "dot"
	EOF       string = "$"
)
