package ksql

import (
	"errors"
	"io"
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
			tokens: []Token{{kind: QuotedString, start: 0, len: 8}},
		},
		{
			name:   "parse double quote string blank",
			input:  "\"\"",
			tokens: []Token{{kind: QuotedString, start: 0, len: 2}},
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
			tokens: []Token{{kind: QuotedString, start: 0, len: 8}},
		},
		{
			name:   "parse single quote string blank",
			input:  "''",
			tokens: []Token{{kind: QuotedString, start: 0, len: 2}},
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
			tokens: []Token{{kind: BooleanTrue, start: 0, len: 4}},
		},
		{
			name:   "parse bool false",
			input:  "false",
			tokens: []Token{{kind: BooleanFalse, start: 0, len: 5}},
		},
		{
			name:  "parse invalid bool",
			input: "fool",
			err:   ErrInvalidBool{s: "fool"},
		},
		{
			name:   "parse number float",
			input:  "123.23",
			tokens: []Token{{kind: Number, start: 0, len: 6}},
		},
		{
			name:   "parse number exp",
			input:  "1e-10",
			tokens: []Token{{kind: Number, len: 5}},
		},
		{
			name:   "parse number int",
			input:  "123",
			tokens: []Token{{kind: Number, len: 3}},
		},
		{
			name:  "parse number invalid",
			input: "123.23.23",
			err:   ErrInvalidNumber{s: "123.23.23"},
		},
		{
			name:   "parse identifier",
			input:  ".properties.first_name",
			tokens: []Token{{kind: SelectorPath, len: 22}},
		},
		{
			name:  "parse identifier blank",
			input: ".",
			err:   ErrInvalidIdentifier{s: "."},
		},
		{
			name:   "parse equals",
			input:  "==",
			tokens: []Token{{kind: Equals, len: 2}},
		},
		{
			name:   "parse add",
			input:  "+",
			tokens: []Token{{kind: Add, len: 1}},
		},
		{
			name:   "parse Substract",
			input:  "-",
			tokens: []Token{{kind: Subtract, len: 1}},
		},
		{
			name:   "parse multiply",
			input:  "*",
			tokens: []Token{{kind: Multiply, len: 1}},
		},
		{
			name:   "parse divide",
			input:  "/",
			tokens: []Token{{kind: Divide, len: 1}},
		},
		{
			name:   "parse gt",
			input:  ">",
			tokens: []Token{{kind: Gt, len: 1}},
		},
		{
			name:   "parse gte",
			input:  ">=",
			tokens: []Token{{kind: Gte, len: 2}},
		},
		{
			name:   "parse lt",
			input:  "<",
			tokens: []Token{{kind: Lt, len: 1}},
		},
		{
			name:   "parse lte",
			input:  "<=",
			tokens: []Token{{kind: Lte, len: 2}},
		},
		{
			name:   "parse open paren",
			input:  "(",
			tokens: []Token{{kind: OpenParen, len: 1}},
		},
		{
			name:   "parse close paren",
			input:  ")",
			tokens: []Token{{kind: CloseParen, len: 1}},
		},
		{
			name:   "parse open bracket",
			input:  "[",
			tokens: []Token{{kind: OpenBracket, len: 1}},
		},
		{
			name:   "parse close bracket",
			input:  "]",
			tokens: []Token{{kind: CloseBracket, len: 1}},
		},
		{
			name:   "parse comma",
			input:  ",",
			tokens: []Token{{kind: Comma, len: 1}},
		},
		{
			name:  "parse add selectorPath",
			input: ".field1 + .field2",
			tokens: []Token{
				{kind: SelectorPath, len: 7},
				{kind: Add, start: 8, len: 1},
				{kind: SelectorPath, start: 10, len: 7},
			},
		},
		{
			name:  "parse sub selectorPath",
			input: ".field1 - .field2",
			tokens: []Token{
				{kind: SelectorPath, len: 7},
				{kind: Subtract, start: 8, len: 1},
				{kind: SelectorPath, start: 10, len: 7},
			},
		},
		{
			name:  "parse brackets",
			input: ".field1 - ( .field2 + .field3 )",
			tokens: []Token{
				{kind: SelectorPath, len: 7},
				{kind: Subtract, start: 8, len: 1},
				{kind: OpenParen, start: 10, len: 1},
				{kind: SelectorPath, start: 12, len: 7},
				{kind: Add, start: 20, len: 1},
				{kind: SelectorPath, start: 22, len: 7},
				{kind: CloseParen, start: 30, len: 1},
			},
		},
		{
			name:   "parse or",
			input:  "||",
			tokens: []Token{{kind: Or, len: 2}},
		},
		{
			name:   "parse in",
			input:  " IN ",
			tokens: []Token{{kind: In, start: 1, len: 2}},
		},
		{
			name:   "parse contains",
			input:  " CONTAINS ",
			tokens: []Token{{kind: Contains, start: 1, len: 8}},
		},
		{
			name:   "parse STARTSWITH",
			input:  " STARTSWITH ",
			tokens: []Token{{kind: StartsWith, start: 1, len: 10}},
		},
		{
			name:   "parse ENDSWITH",
			input:  " ENDSWITH ",
			tokens: []Token{{kind: EndsWith, start: 1, len: 8}},
		},
		{
			name:   "parse AND",
			input:  "&&",
			tokens: []Token{{kind: And, start: 0, len: 2}},
		},
		{
			name:   "parse NULL",
			input:  "NULL",
			tokens: []Token{{kind: Null, start: 0, len: 4}},
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
			tokens: []Token{{kind: Not, start: 0, len: 1}},
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
		token, err := tokenizer.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tokens, nil
			}
			return tokens, err
		}
		tokens = append(tokens, token)
	}
}
