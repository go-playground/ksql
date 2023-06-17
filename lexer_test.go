package ksql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		name   string
		input  string
		tokens []Token
		err    error
	}{
		{
			name:   "parse double quote string",
			input:  "\"quoted\"",
			tokens: []Token{{Kind: QuotedString, Start: 0, Len: 8}},
		},
		{
			name:   "parse double quote string blank",
			input:  "\"\"",
			tokens: []Token{{Kind: QuotedString, Start: 0, Len: 2}},
		},
		{
			name:  "parse double quote string unterminated",
			input: "\"dfg",
			err:   ErrUnterminatedString{s: "\"dfg"},
		},
		{
			name:  "parse double quote string unterminated bare",
			input: "\"",
			err:   ErrUnterminatedString{s: "\""},
		},
		{
			name:   "parse single quote string",
			input:  "'quoted'",
			tokens: []Token{{Kind: QuotedString, Start: 0, Len: 8}},
		},
		{
			name:   "parse single quote string blank",
			input:  "''",
			tokens: []Token{{Kind: QuotedString, Start: 0, Len: 2}},
		},
		{
			name:  "parse single quote string unterminated",
			input: "'dfg",
			err:   ErrUnterminatedString{s: "'dfg"},
		},
		{
			name:  "parse single quote string unterminated bare",
			input: "'",
			err:   ErrUnterminatedString{s: "'"},
		},
		{
			name:   "parse bool true",
			input:  "true",
			tokens: []Token{{Kind: BooleanTrue, Start: 0, Len: 4}},
		},
		{
			name:   "parse bool false",
			input:  "false",
			tokens: []Token{{Kind: BooleanFalse, Start: 0, Len: 5}},
		},
		{
			name:  "parse invalid bool",
			input: "fool",
			err:   ErrInvalidBool{s: "fool"},
		},
		{
			name:   "parse number float",
			input:  "123.23",
			tokens: []Token{{Kind: Number, Start: 0, Len: 6}},
		},
		{
			name:   "parse number Exp",
			input:  "1e-10",
			tokens: []Token{{Kind: Number, Len: 5}},
		},
		{
			name:   "parse number int",
			input:  "123",
			tokens: []Token{{Kind: Number, Len: 3}},
		},
		{
			name:  "parse number invalid",
			input: "123.23.23",
			err:   ErrInvalidNumber{s: "123.23.23"},
		},
		{
			name:   "parse identifier",
			input:  ".properties.first_name",
			tokens: []Token{{Kind: SelectorPath, Len: 22}},
		},
		{
			name:  "parse identifier blank",
			input: ".",
			err:   ErrInvalidSelectorPath{s: "."},
		},
		{
			name:   "parse equals",
			input:  "==",
			tokens: []Token{{Kind: Equals, Len: 2}},
		},
		{
			name:   "parse add",
			input:  "+",
			tokens: []Token{{Kind: Add, Len: 1}},
		},
		{
			name:   "parse Substract",
			input:  "-",
			tokens: []Token{{Kind: Subtract, Len: 1}},
		},
		{
			name:   "parse multiply",
			input:  "*",
			tokens: []Token{{Kind: Multiply, Len: 1}},
		},
		{
			name:   "parse divide",
			input:  "/",
			tokens: []Token{{Kind: Divide, Len: 1}},
		},
		{
			name:   "parse gt",
			input:  ">",
			tokens: []Token{{Kind: Gt, Len: 1}},
		},
		{
			name:   "parse gte",
			input:  ">=",
			tokens: []Token{{Kind: Gte, Len: 2}},
		},
		{
			name:   "parse lt",
			input:  "<",
			tokens: []Token{{Kind: Lt, Len: 1}},
		},
		{
			name:   "parse lte",
			input:  "<=",
			tokens: []Token{{Kind: Lte, Len: 2}},
		},
		{
			name:   "parse open paren",
			input:  "(",
			tokens: []Token{{Kind: OpenParen, Len: 1}},
		},
		{
			name:   "parse close paren",
			input:  ")",
			tokens: []Token{{Kind: CloseParen, Len: 1}},
		},
		{
			name:   "parse open bracket",
			input:  "[",
			tokens: []Token{{Kind: OpenBracket, Len: 1}},
		},
		{
			name:   "parse close bracket",
			input:  "]",
			tokens: []Token{{Kind: CloseBracket, Len: 1}},
		},
		{
			name:   "parse comma",
			input:  ",",
			tokens: []Token{{Kind: Comma, Len: 1}},
		},
		{
			name:  "parse add selectorPath",
			input: ".field1 + .field2",
			tokens: []Token{
				{Kind: SelectorPath, Len: 7},
				{Kind: Add, Start: 8, Len: 1},
				{Kind: SelectorPath, Start: 10, Len: 7},
			},
		},
		{
			name:  "parse sub selectorPath",
			input: ".field1 - .field2",
			tokens: []Token{
				{Kind: SelectorPath, Len: 7},
				{Kind: Subtract, Start: 8, Len: 1},
				{Kind: SelectorPath, Start: 10, Len: 7},
			},
		},
		{
			name:  "parse brackets",
			input: ".field1 - ( .field2 + .field3 )",
			tokens: []Token{
				{Kind: SelectorPath, Len: 7},
				{Kind: Subtract, Start: 8, Len: 1},
				{Kind: OpenParen, Start: 10, Len: 1},
				{Kind: SelectorPath, Start: 12, Len: 7},
				{Kind: Add, Start: 20, Len: 1},
				{Kind: SelectorPath, Start: 22, Len: 7},
				{Kind: CloseParen, Start: 30, Len: 1},
			},
		},
		{
			name:   "parse or",
			input:  "||",
			tokens: []Token{{Kind: Or, Len: 2}},
		},
		{
			name:   "parse in",
			input:  " IN ",
			tokens: []Token{{Kind: In, Start: 1, Len: 2}},
		},
		{
			name:   "parse contains",
			input:  " CONTAINS ",
			tokens: []Token{{Kind: Contains, Start: 1, Len: 8}},
		},
		{
			name:   "parse contains any",
			input:  " CONTAINS_ANY ",
			tokens: []Token{{Kind: ContainsAny, Start: 1, Len: 12}},
		},
		{
			name:   "parse contains all",
			input:  " CONTAINS_ALL ",
			tokens: []Token{{Kind: ContainsAll, Start: 1, Len: 12}},
		},
		{
			name:   "parse STARTSWITH",
			input:  " STARTSWITH ",
			tokens: []Token{{Kind: StartsWith, Start: 1, Len: 10}},
		},
		{
			name:   "parse ENDSWITH",
			input:  " ENDSWITH ",
			tokens: []Token{{Kind: EndsWith, Start: 1, Len: 8}},
		},
		{
			name:   "parse AND",
			input:  "&&",
			tokens: []Token{{Kind: And, Start: 0, Len: 2}},
		},
		{
			name:   "parse NULL",
			input:  "NULL",
			tokens: []Token{{Kind: Null, Start: 0, Len: 4}},
		},
		{
			name:  "parse bad or",
			input: "|",
			err:   ErrInvalidKeyword{s: "|"},
		},
		{
			name:  "parse bad in",
			input: " IN",
			err:   ErrInvalidKeyword{s: "IN"},
		},
		{
			name:  "parse bad contains",
			input: " CONTAINS",
			err:   ErrInvalidKeyword{s: "CONTAINS"},
		},
		{
			name:  "parse bad STARTSWITH",
			input: " STARTSWITH",
			err:   ErrInvalidKeyword{s: "STARTSWITH"},
		},
		{
			name:  "parse bad ENDSWITH",
			input: " ENDSWITH",
			err:   ErrInvalidKeyword{s: "ENDSWITH"},
		},
		{
			name:  "parse bad AND",
			input: "&",
			err:   ErrInvalidKeyword{s: "&"},
		},
		{
			name:  "parse bad NULL",
			input: "NULLL",
			err:   ErrInvalidKeyword{s: "NULLL"},
		},
		{
			name:   "parse not",
			input:  "!",
			tokens: []Token{{Kind: Not, Start: 0, Len: 1}},
		},
		{
			name:  "parse bad identifier",
			input: "_datetime",
			err:   ErrInvalidIdentifier{s: "_datetime"},
		},
		{
			name:   "parse identifier",
			input:  "_datetime_",
			tokens: []Token{{Kind: Identifier, Start: 0, Len: 10}},
		},
		{
			name:   "parse COERCE",
			input:  "COERCE ",
			tokens: []Token{{Kind: Coerce, Start: 0, Len: 6}},
		},
		{
			name:  "parse COERCE",
			input: `COERCE "2022-01-02" _datetime_`,
			tokens: []Token{
				{Kind: Coerce, Start: 0, Len: 6},
				{Kind: QuotedString, Start: 7, Len: 12},
				{Kind: Identifier, Start: 20, Len: 10},
			},
		},
		{
			name:  "parse bad contains",
			input: "CONTAINS",
			err:   ErrInvalidKeyword{s: "CONTAINS"},
		},
		{
			name:  "parse bad contains any",
			input: "CONTAINS_ANY",
			err:   ErrInvalidKeyword{s: "CONTAINS_ANY"},
		},
		{
			name:  "parse bad contains all",
			input: "CONTAINS_ALL",
			err:   ErrInvalidKeyword{s: "CONTAINS_ALL"},
		},
		{
			name:  "parse bad between",
			input: "BETWEEN",
			err:   ErrInvalidKeyword{s: "BETWEEEN"},
		},
		{
			name:   "parse BETWEEN",
			input:  "BETWEEN ",
			tokens: []Token{{Kind: Between, Start: 0, Len: 7}},
		},
		{
			name:   "parse negative number",
			input:  " -1.23 ",
			tokens: []Token{{Kind: Number, Start: 1, Len: 5}},
		},
		{
			name:   "parse positive number",
			input:  " +1.23 ",
			tokens: []Token{{Kind: Number, Start: 1, Len: 5}},
		},
		{
			name:   "parse positive number",
			input:  " +1.23 ",
			tokens: []Token{{Kind: Number, Start: 1, Len: 5}},
		},
		{
			name:   "parse negative exponential number",
			input:  " -1e10 ",
			tokens: []Token{{Kind: Number, Start: 1, Len: 5}},
		},
		{
			name:   "parse positive exponential number",
			input:  " +1e10 ",
			tokens: []Token{{Kind: Number, Start: 1, Len: 5}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tokens, err := collect([]byte(tc.input))
			if tc.err != nil {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.tokens, tokens)
		})
	}
}

// Collect tokenizes the input and returns tokens or error lexing them.
func collect(src []byte) (tokens []Token, err error) {
	tokenizer := NewTokenizer(src)

	for {
		next := tokenizer.Next()
		if next.IsNone() {
			break
		}
		result := next.Unwrap()
		if result.IsErr() {
			return nil, result.Err()
		}
		tokens = append(tokens, result.Unwrap())
	}
	return
}
