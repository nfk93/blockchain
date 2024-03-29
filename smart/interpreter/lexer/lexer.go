// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"github.com/nfk93/blockchain/smart/interpreter/token"
)

const (
	NoState    = -1
	NumStates  = 153
	NumSymbols = 190
)

type Lexer struct {
	src    []byte
	pos    int
	line   int
	column int
}

func NewLexer(src []byte) *Lexer {
	lexer := &Lexer{
		src:    src,
		pos:    0,
		line:   1,
		column: 1,
	}
	return lexer
}

func NewLexerFile(fpath string) (*Lexer, error) {
	src, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return NewLexer(src), nil
}

func (l *Lexer) Scan() (tok *token.Token) {
	tok = new(token.Token)
	if l.pos >= len(l.src) {
		tok.Type = token.EOF
		tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = l.pos, l.line, l.column
		return
	}
	start, startLine, startColumn, end := l.pos, l.line, l.column, 0
	tok.Type = token.INVALID
	state, rune1, size := 0, rune(-1), 0
	for state != -1 {
		if l.pos >= len(l.src) {
			rune1 = -1
		} else {
			rune1, size = utf8.DecodeRune(l.src[l.pos:])
			l.pos += size
		}

		nextState := -1
		if rune1 != -1 {
			nextState = TransTab[state](rune1)
		}
		state = nextState

		if state != -1 {

			switch rune1 {
			case '\n':
				l.line++
				l.column = 1
			case '\r':
				l.column = 1
			case '\t':
				l.column += 4
			default:
				l.column++
			}

			switch {
			case ActTab[state].Accept != -1:
				tok.Type = ActTab[state].Accept
				end = l.pos
			case ActTab[state].Ignore != "":
				start, startLine, startColumn = l.pos, l.line, l.column
				state = 0
				if start >= len(l.src) {
					tok.Type = token.EOF
				}

			}
		} else {
			if tok.Type == token.INVALID {
				end = l.pos
			}
		}
	}
	if end > start {
		l.pos = end
		tok.Lit = l.src[start:end]
	} else {
		tok.Lit = []byte{}
	}
	tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = start, startLine, startColumn

	return
}

func (l *Lexer) Reset() {
	l.pos = 0
}

/*
Lexer symbols:
0: ','
1: '>'
2: '='
3: '>'
4: '<'
5: '='
6: '<'
7: '-'
8: '<'
9: '>'
10: '<'
11: '&'
12: '&'
13: '&'
14: 'l'
15: 'a'
16: 'n'
17: 'd'
18: '|'
19: '|'
20: 'l'
21: 'o'
22: 'r'
23: 'o'
24: 'r'
25: 'n'
26: 'o'
27: 't'
28: '='
29: '+'
30: '-'
31: '~'
32: '-'
33: '/'
34: '{'
35: '}'
36: '['
37: ']'
38: '('
39: ')'
40: ':'
41: ':'
42: ':'
43: ';'
44: '*'
45: 'k'
46: 'e'
47: 'y'
48: 'o'
49: 'p'
50: 'e'
51: 'r'
52: 'a'
53: 't'
54: 'i'
55: 'o'
56: 'n'
57: 'o'
58: 'p'
59: 't'
60: 'i'
61: 'o'
62: 'n'
63: 'l'
64: 'i'
65: 's'
66: 't'
67: 'b'
68: 'o'
69: 'o'
70: 'l'
71: 'u'
72: 'n'
73: 'i'
74: 't'
75: 'n'
76: 'a'
77: 't'
78: 'i'
79: 'n'
80: 't'
81: 'a'
82: 'd'
83: 'd'
84: 'r'
85: 'e'
86: 's'
87: 's'
88: 's'
89: 't'
90: 'r'
91: 'i'
92: 'n'
93: 'g'
94: 'f'
95: 'a'
96: 'l'
97: 's'
98: 'e'
99: 't'
100: 'r'
101: 'u'
102: 'e'
103: 'l'
104: 'e'
105: 't'
106: '%'
107: 'i'
108: 'n'
109: 'i'
110: 't'
111: 'l'
112: 'e'
113: 't'
114: '%'
115: 'e'
116: 'n'
117: 't'
118: 'r'
119: 'y'
120: 'l'
121: 'e'
122: 't'
123: 'i'
124: 'n'
125: 'i'
126: 'f'
127: 't'
128: 'h'
129: 'e'
130: 'n'
131: 'e'
132: 'l'
133: 's'
134: 'e'
135: 't'
136: 'y'
137: 'p'
138: 'e'
139: 'k'
140: 'o'
141: 'i'
142: 'n'
143: 'k'
144: 'n'
145: '1'
146: 'k'
147: 'n'
148: '2'
149: '_'
150: '"'
151: '"'
152: '.'
153: 'k'
154: 'n'
155: 'k'
156: 'n'
157: 'p'
158: '-'
159: '.'
160: '_'
161: '('
162: '*'
163: '*'
164: ')'
165: '['
166: '%'
167: '%'
168: 'v'
169: 'e'
170: 'r'
171: 's'
172: 'i'
173: 'o'
174: 'n'
175: ' '
176: ']'
177: ' '
178: '\t'
179: '\n'
180: '\r'
181: 'a'-'z'
182: 'A'-'Z'
183: '0'-'9'
184: 'a'-'f'
185: 'a'-'z'
186: 'A'-'Z'
187: '0'-'9'
188: '0'-'9'
189: .
*/
