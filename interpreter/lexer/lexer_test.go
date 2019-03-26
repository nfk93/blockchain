package lexer

import (
	"github.com/nfk93/blockchain/interpreter/token"
	"testing"
)

var code_storage = []byte("type storage = {owner : key_hash; funding_goal : tez; amount_raised : tez; soft_cap : tez; }")
var code_init_storage = []byte(
	"let%init storage = {\n" +
		"\towner = koin1YLtLqD1fWHthSVHPD116oYvsd4PTAHUoc;\n" +
		"\tfunding_goal = 100tz;\n" +
		"\tamount_raised = 0tz;\n" +
		"\tsoft_cap  = 75tz;\n" +
		"}")
var correct_tokens []token.Type

func TestLexStorage(t *testing.T) {
	lex := NewLexer(code_storage)

	strings := []string{"type", "id", "eq", "lbrace", "id", "colon", "keyhash", "semicolon", "id", "colon", "tez", "semicolon",
		"id", "colon", "tez", "semicolon", "id", "colon", "tez", "semicolon", "rbrace", "$"}
	correct_tokens = make([]token.Type, 22)
	for idx, str := range strings {
		correct_tokens[idx] = token.TokMap.Type(str)
	}

	for i := 0; i < len(correct_tokens); i++ {
		nextToken := lex.Scan()
		if nextToken.Type != correct_tokens[i] {
			t.Error("\nWrong token at index", i,
				"\n\tGot token:", nextToken.Type, nextToken.String(), token.TokMap.Id(nextToken.Type),
				"\n\tExpected:", correct_tokens[i], token.TokMap.Id(correct_tokens[i]))
		}
	}
}

func TestLexInitStorage(t *testing.T) {
	lex := NewLexer(code_init_storage)

	strings := []string{
		"let", "percentage", "id", "id", "eq", "lbrace", "id", "eq", "hash", "semicolon", "id", "eq", "tez_lit", "semicolon",
		"id", "eq", "tez_lit", "semicolon", "id", "eq", "tez_lit", "semicolon", "rbrace", "$"}
	correct_tokens = make([]token.Type, 24)
	for idx, str := range strings {
		correct_tokens[idx] = token.TokMap.Type(str)
	}
	for i := 0; i < len(correct_tokens); i++ {
		nextToken := lex.Scan()
		if nextToken.Type != correct_tokens[i] {
			t.Error("\nWrong token at index", i,
				"\n\tGot token:", nextToken.Type, nextToken.String(), token.TokMap.Id(nextToken.Type),
				"\n\tExpected:", correct_tokens[i], token.TokMap.Id(correct_tokens[i]))
		}
	}
}
