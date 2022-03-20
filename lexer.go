package ksql

import (
	"errors"
	"io"
	"strconv"
)

// Result represents a token lexed result
type Result struct {
	token Token
	end   int
}

// Token represents a lexed token with value
type Token struct {
	kind  TokenKind
	value interface{}
}

// TokenKind is the type of token lexed.
type TokenKind uint8

const (
	Identifier = iota
	String
	Number
	Boolean
	Null
	Equals
	Add
	Subtract
	Multiply
	Divide
	Gt
	Gte
	Lt
	Lte
	And
	Or
	Not
	Contains
	In
	StartsWith
	EndsWith
	OpenBracket
	CloseBracket
	Comma
	OpenParen
	CloseParen
)

/// Try to lex a single token from the input stream.
func tokenizeSingleToken(data []byte) (result Result, err error) {
	b := data[0]

	switch b {
	case '=':
		result = Result{token: Token{kind: Equals}, end: 1}
	case '+':
		result = Result{token: Token{kind: Add}, end: 1}
	case '-':
		result = Result{token: Token{kind: Subtract}, end: 1}
	case '*':
		result = Result{token: Token{kind: Multiply}, end: 1}
	case '/':
		result = Result{token: Token{kind: Divide}, end: 1}
	case '>':
		if len(data) > 1 && data[1] == '=' {
			result = Result{token: Token{kind: Gte}, end: 2}

		} else if len(data) > 1 {
			err = ErrUnsupportedCharacter{b: b}
		} else {
			result = Result{token: Token{kind: Gt}, end: 1}
		}
	case '<':
		if len(data) > 1 && data[1] == '=' {
			result = Result{token: Token{kind: Lte}, end: 2}
		} else if len(data) > 1 {
			err = ErrUnsupportedCharacter{b: b}
		} else {
			result = Result{token: Token{kind: Lt}, end: 1}
		}
	case '(':
		result = Result{token: Token{kind: OpenParen}, end: 1}
	case ')':
		result = Result{token: Token{kind: CloseParen}, end: 1}
	case '[':
		result = Result{token: Token{kind: OpenBracket}, end: 1}
	case ']':
		result = Result{token: Token{kind: CloseBracket}, end: 1}
	case ',':
		result = Result{token: Token{kind: Comma}, end: 1}
	case '!':
		result = Result{token: Token{kind: Not}, end: 1}
	case '"', '\'':
		result, err = tokenizeString(data, b)
	case '.':
		result, err = tokenizeIdentifier(data)
	case 't', 'f':
		result, err = tokenizeBool(data)
	case '&':
		if len(data) > 1 && data[1] == '&' {
			result = Result{token: Token{kind: And}, end: 2}
		} else {
			err = ErrUnsupportedCharacter{b: b}
		}
	case '|':
		if len(data) > 1 && data[1] == '|' {
			result = Result{token: Token{kind: Or}, end: 2}
		} else {
			err = ErrUnsupportedCharacter{b: b}
		}
	case 'C':
		result, err = tokenizeKeyword(data, "CONTAINS", Contains)
	case 'I':
		result, err = tokenizeKeyword(data, "IN", In)
	case 'S':
		result, err = tokenizeKeyword(data, "STARTSWITH", StartsWith)
	case 'E':
		result, err = tokenizeKeyword(data, "ENDSWITH", EndsWith)
	case 'N':
		result, err = tokenizeNull(data)
	default:
		if isDigit(b) {
			result, err = tokenizeNumber(data)
		} else {
			err = ErrUnsupportedCharacter{b: b}
		}
	}

	return
}

func tokenizeNumber(data []byte) (result Result, err error) {
	var dotSeen, badNumber bool

	end := takeWhile(data, func(b byte) bool {
		switch b {
		case '.':
			if dotSeen {
				badNumber = true
				return false
			}
			dotSeen = true
			return true
		case '-', '+':
			return true
		default:
			return isAlphanumeric(b)
		}
	})

	if end > 0 && !badNumber {
		n, err := strconv.ParseFloat(string(data[:end]), 64)
		if err != nil {
			err = ErrInvalidNumber{s: string(data[:end])}
		} else {
			result = Result{
				token: Token{
					kind:  Number,
					value: n,
				},
				end: end,
			}
		}
	} else {
		err = ErrInvalidNumber{s: string(data)}
	}
	return
}

func tokenizeKeyword(data []byte, keyword string, kind TokenKind) (result Result, err error) {
	end := takeWhile(data, func(b byte) bool {
		return !isWhitespace(b)
	})
	if end > 0 && string(data[:end]) == keyword && len(data) > len(keyword) {
		result = Result{
			token: Token{
				kind: kind,
			},
			end: end,
		}
	} else {
		err = ErrInvalidKeyword{s: string(data)}
	}
	return
}

func tokenizeNull(data []byte) (result Result, err error) {
	end := takeWhile(data, func(b byte) bool {
		return isAlphabetical(b)
	})
	if end > 0 && string(data[:end]) == "NULL" {
		result = Result{
			token: Token{
				kind: Null,
			},
			end: end,
		}
	} else {
		err = ErrInvalidKeyword{s: string(data)}
	}
	return
}

func tokenizeBool(data []byte) (result Result, err error) {
	end := takeWhile(data, func(b byte) bool {
		return isAlphabetical(b)
	})
	if end > 0 {
		switch string(data[:end]) {
		case "true":
			result = Result{
				token: Token{
					kind:  Boolean,
					value: true,
				},
				end: end,
			}
		case "false":
			result = Result{
				token: Token{
					kind:  Boolean,
					value: false,
				},
				end: end,
			}
		default:
			err = ErrInvalidBool{s: string(data)}
		}
	} else {
		err = ErrInvalidBool{s: string(data)}
	}
	return
}

func tokenizeIdentifier(data []byte) (result Result, err error) {
	end := takeWhile(data[1:], func(b byte) bool {
		return !isWhitespace(b)
	})
	if end > 0 {
		if len(data) > end {
			end += 1
		}
		result = Result{token: Token{
			kind:  Identifier,
			value: string(data[1:end]),
		}, end: end}
	} else {
		err = ErrInvalidIdentifier{s: string(data)}
	}
	return
}

func tokenizeString(data []byte, quote byte) (result Result, err error) {
	var lastBackslash, endedWithoutTerminator bool

	end := takeWhile(data[1:], func(b byte) bool {
		switch b {
		case '\\':
			lastBackslash = true
			return true
		case quote:
			if lastBackslash {
				lastBackslash = false
				return true
			}
			endedWithoutTerminator = true
			return false
		default:
			return true
		}
	})

	if end > 0 {
		if !endedWithoutTerminator {
			err = ErrUnterminatedString{s: string(data)}
		} else {
			result = Result{token: Token{
				kind:  String,
				value: string(data[1 : end+1]),
			}, end: end + 2}
		}
	} else {
		if !endedWithoutTerminator || len(data) < 2 {
			err = ErrUnterminatedString{s: string(data)}
		} else {
			result = Result{token: Token{
				kind:  String,
				value: string(data[:0]),
			}, end: 1}
		}
	}
	return
}

/// Consumes bytes while a predicate evaluates to true.
func takeWhile(data []byte, pred func(byte) bool) (end int) {
	for _, b := range data {
		if !pred(b) {
			break
		}
		end++
	}
	return
}

type Tokenizer struct {
	current   int
	remaining []byte
}

func skipWhitespace(data []byte) int {
	return takeWhile(data, func(b byte) bool {
		return isWhitespace(b)
	})
}

// NewTokenizer creates a new tokenizer for use
func newTokenizer(src []byte) *Tokenizer {
	return &Tokenizer{
		current:   0,
		remaining: src,
	}
}

func (t *Tokenizer) nextToken() (token Token, err error) {
	t.skipWhitespace()

	if len(t.remaining) == 0 {
		err = io.EOF
		return
	}
	return t.next()
}

func (t *Tokenizer) skipWhitespace() {
	skipped := skipWhitespace(t.remaining)
	t.chomp(skipped)
}

func (t *Tokenizer) next() (token Token, err error) {
	var result Result
	result, err = tokenizeSingleToken(t.remaining)
	if err != nil {
		return
	}
	t.chomp(result.end)
	return result.token, nil
}

func (t *Tokenizer) chomp(num int) {
	t.remaining = t.remaining[num:]
	t.current += num
}

// Tokenize tokenizes the input and returns tokens or error lexing them.
func Tokenize(src []byte) (tokens []Token, err error) {
	tokenizer := newTokenizer(src)

	for {
		token, err := tokenizer.nextToken()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tokens, nil
			}
			return tokens, err
		}
		tokens = append(tokens, token)
	}
	return
}

func isAlphanumeric(c byte) bool {
	return isLower(c) || isUpper(c) || isDigit(c)
}

func isAlphabetical(c byte) bool {
	return isLower(c) || isUpper(c)
}

func isUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isWhitespace(b byte) bool {
	switch b {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	default:
		return false
	}
}
