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
		TYPE, LIDENT, EQ, LBRACE, LIDENT, COLON, KEYHASH, SEMICOLON, LIDENT, COLON, TEZ, SEMICOLON, LIDENT, COLON, TEZ,
		SEMICOLON, LIDENT, COLON, TEZ, SEMICOLON, RBRACE, EOF}
	compare_tokens(t, strings, lex)
}

func TestLexInitStorage(t *testing.T) {
	bytes := read_file("tests/init_storage", t)
	lex := NewLexer(bytes)

	strings := []string{
		LET, PERC, LIDENT, LIDENT, EQ, LBRACE, LIDENT, EQ, HASH, SEMICOLON, LIDENT, EQ, TEZ_LIT, SEMICOLON, LIDENT, EQ, TEZ_LIT, SEMICOLON,
		LIDENT, EQ, TEZ_LIT, SEMICOLON, RBRACE, EOF}
	compare_tokens(t, strings, lex)
}

func TestLexSimpleEntry(t *testing.T) {
	bytes := read_file("tests/simple_entry", t)
	lex := NewLexer(bytes)

	strings := []string{
		LET, PERC, LIDENT, LIDENT, LPAREN, LIDENT, COLON, KEYHASH, RPAREN, LPAREN, LIDENT, COLON, LIDENT, RPAREN, EQ, IF, LIDENT, DOT, LIDENT, GEQ,
		LIDENT, DOT, LIDENT, THEN, LET, LIDENT, EQ, LIDENT, DOT, LIDENT, LARROW, LIDENT, IN, LPAREN, LPAREN, LBRACK, RBRACK, COLON, OPERATION,
		LIST, RPAREN, COMMA, LIDENT, RPAREN, ELSE, LPAREN, LPAREN, LBRACK, RBRACK, COLON, OPERATION, LIST, RPAREN, COMMA,
		LIDENT, RPAREN, EOF}
	compare_tokens(t, strings, lex)
}

func TestLexFloat(t *testing.T) {
	bytes := read_file("tests/float", t)
	lex := NewLexer(bytes)

	strings := []string{
		LET, LIDENT, EQ, FLOAT_LIT, IN,
		LET, LIDENT, EQ, FLOAT_LIT, IN,
		LET, LIDENT, EQ, FLOAT_LIT, IN,
		LET, LIDENT, EQ, INT_LIT, IN,
		LET, LIDENT, EQ, FLOAT_LIT, IN,
		LET, LIDENT, EQ, FLOAT_LIT, IN,
		LET, LIDENT, EQ, LIDENT, PLUS, LIDENT, MINUS, LIDENT, EOF}
	compare_tokens(t, strings, lex)
}

func TestNoInvalidsInFundMe(t *testing.T) {
	bytes := read_file("tests/fundme", t)
	lex := NewLexer(bytes)

	for {
		nextToken := lex.Scan()
		if nextToken.Type == 0 {
			t.Error("\nInvalid token:", nextToken.Type, token.TokMap.Id(nextToken.Type), nextToken.String())
		} else if nextToken.Type == token.TokMap.Type("$") {
			break
		}
	}
}

func TestLidUid(t *testing.T) {
	bytes := read_file("tests/lid_uid", t)
	lex := NewLexer(bytes)
	strings := []string{
		LIDENT, LIDENT, UIDENT, UIDENT, LIDENT, LIDENT, EOF}
	compare_tokens(t, strings, lex)
}

// TODO: make tests covering all of the below

const (
	COMMA      string = "comma"
	GEQ        string = "geq"
	GT         string = "gt"
	LEQ        string = "leq"
	LARROW     string = "larrow"
	NEQ        string = "neq"
	LT         string = "lt"
	RARROW     string = "rarrow"
	EQ         string = "eq"
	PLUS       string = "plus"
	MINUS      string = "minus"
	LBRACE     string = "lbrace"
	RBRACE     string = "rbrace"
	LBRACK     string = "lbrack"
	RBRACK     string = "rbrack"
	LPAREN     string = "lparen"
	RPAREN     string = "rparen"
	COLON      string = "colon"
	SEMICOLON  string = "semicolon"
	KEYHASH    string = "keyhash"
	OPERATION  string = "operation"
	LIST       string = "list"
	PERC       string = "percentage"
	LET        string = "let"
	IN         string = "in"
	IF         string = "if"
	THEN       string = "then"
	ELSE       string = "else"
	TYPE       string = "type"
	TEZ        string = "tez"
	HASH       string = "hash"
	LIDENT     string = "lident"
	UIDENT     string = "uident"
	STRING_LIT string = "string_lit"
	TEZ_LIT    string = "tez_lit"
	INT_LIT    string = "int_lit"
	FLOAT_LIT  string = "float_lit"
	DOT        string = "dot"
	AST        string = "ast"
	TRUE       string = "true"
	FALSE      string = "false"
	BOOL       string = "bool"
	STRING     string = "string"
	INT        string = "int"
	EOF        string = "$"
)
