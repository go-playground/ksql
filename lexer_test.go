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
			tokens: []Token{{kind: String, value: "quoted"}},
		},
		{
			name:   "parse double quote string blank",
			input:  "\"\"",
			tokens: []Token{{kind: String, value: ""}},
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
			tokens: []Token{{kind: String, value: "quoted"}},
		},
		{
			name:   "parse single quote string blank",
			input:  "''",
			tokens: []Token{{kind: String, value: ""}},
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
			tokens: []Token{{kind: Boolean, value: true}},
		},
		{
			name:   "parse bool false",
			input:  "false",
			tokens: []Token{{kind: Boolean, value: false}},
		},
		{
			name:  "parse invalid bool",
			input: "fool",
			err:   ErrInvalidBool{s: "fool"},
		},
		{
			name:   "parse number float",
			input:  "123.23",
			tokens: []Token{{kind: Number, value: 123.23}},
		},
		{
			name:   "parse number exp",
			input:  "1e-10",
			tokens: []Token{{kind: Number, value: 1e-10}},
		},
		{
			name:   "parse number int",
			input:  "123",
			tokens: []Token{{kind: Number, value: float64(123)}},
		},
		{
			name:  "parse number invalid",
			input: "123.23.23",
			err:   ErrInvalidNumber{s: "123.23.23"},
		},
		{
			name:   "parse identifier",
			input:  ".properties.first_name",
			tokens: []Token{{kind: Identifier, value: "properties.first_name"}},
		},
		{
			name:  "parse identifier blank",
			input: ".",
			err:   ErrInvalidIdentifier{s: "."},
		},
		{
			name:   "parse equals",
			input:  "=",
			tokens: []Token{{kind: Equals}},
		},
		{
			name:   "parse add",
			input:  "+",
			tokens: []Token{{kind: Add}},
		},
		{
			name:   "parse Substract",
			input:  "-",
			tokens: []Token{{kind: Subtract}},
		},
		{
			name:   "parse multiply",
			input:  "*",
			tokens: []Token{{kind: Multiply}},
		},
		{
			name:   "parse divide",
			input:  "/",
			tokens: []Token{{kind: Divide}},
		},
		{
			name:   "parse gt",
			input:  ">",
			tokens: []Token{{kind: Gt}},
		},
		{
			name:   "parse gte",
			input:  ">=",
			tokens: []Token{{kind: Gte}},
		},
		{
			name:   "parse lt",
			input:  "<",
			tokens: []Token{{kind: Lt}},
		},
		{
			name:   "parse lte",
			input:  "<=",
			tokens: []Token{{kind: Lte}},
		},
		{
			name:   "parse open paren",
			input:  "(",
			tokens: []Token{{kind: OpenParen}},
		},
		{
			name:   "parse close paren",
			input:  ")",
			tokens: []Token{{kind: CloseParen}},
		},
		{
			name:   "parse open bracket",
			input:  "[",
			tokens: []Token{{kind: OpenBracket}},
		},
		{
			name:   "parse close bracket",
			input:  "]",
			tokens: []Token{{kind: CloseBracket}},
		},
		{
			name:   "parse comma",
			input:  ",",
			tokens: []Token{{kind: Comma}},
		},
		{
			name:   "parse add ident",
			input:  ".field1 + .field2",
			tokens: []Token{{kind: Identifier, value: "field1"}, {kind: Add}, {kind: Identifier, value: "field2"}},
		},
		{
			name:   "parse sub ident",
			input:  ".field1 - .field2",
			tokens: []Token{{kind: Identifier, value: "field1"}, {kind: Subtract}, {kind: Identifier, value: "field2"}},
		},
		{
			name:   "parse brackets",
			input:  ".field1 - ( .field2 + .field3 )",
			tokens: []Token{{kind: Identifier, value: "field1"}, {kind: Subtract}, {kind: OpenParen}, {kind: Identifier, value: "field2"}, {kind: Add}, {kind: Identifier, value: "field3"}, {kind: CloseParen}},
		},
		{
			name:   "parse or",
			input:  " OR ",
			tokens: []Token{{kind: Or}},
		},
		{
			name:   "parse in",
			input:  " IN ",
			tokens: []Token{{kind: In}},
		},
		{
			name:   "parse contains",
			input:  " CONTAINS ",
			tokens: []Token{{kind: Contains}},
		},
		{
			name:   "parse STARTSWITH",
			input:  " STARTSWITH ",
			tokens: []Token{{kind: StartsWith}},
		},
		{
			name:   "parse ENDSWITH",
			input:  " ENDSWITH ",
			tokens: []Token{{kind: EndsWith}},
		},
		{
			name:   "parse AND",
			input:  " AND ",
			tokens: []Token{{kind: And}},
		},
		{
			name:   "parse NULL",
			input:  "NULL",
			tokens: []Token{{kind: Null}},
		},
		{
			name:  "parse bad or",
			input: " OR",
			err:   ErrInvalidKeyword{s: "OR"},
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
			input: " AND",
			err:   ErrInvalidKeyword{s: "AND"},
		},
		{
			name:  "parse bad NULL",
			input: "NULLL",
			err:   ErrInvalidKeyword{s: "NULLL"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			Tokenize([]byte(tc.input))
			tokens, err := Tokenize([]byte(tc.input))
			if tc.err != nil {
				assert.Error(err)
			} else {
				assert.Equal(tc.tokens, tokens)
			}
		})
	}

}
