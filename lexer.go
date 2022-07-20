package ksql

import (
	"io"
)

// Token represents a lexed token
type Token struct {
	start uint32
	len   uint16
	kind  TokenKind
}

// LexerResult represents a token lexed result
type LexerResult struct {
	kind TokenKind
	len  uint16
}

// TokenKind is the type of token lexed.
type TokenKind uint8

const (
	SelectorPath = iota
	QuotedString
	Number
	BooleanTrue
	BooleanFalse
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
	ContainsAny
	ContainsAll
	In
	Between
	StartsWith
	EndsWith
	OpenBracket
	CloseBracket
	Comma
	OpenParen
	CloseParen
	Coerce
	Identifier
)

/// Try to lex a single token from the input stream.
func tokenizeSingleToken(data []byte) (result LexerResult, err error) {
	b := data[0]

	switch b {
	case '=':
		if len(data) > 1 && data[1] == '=' {
			result = LexerResult{kind: Equals, len: 2}
		} else {
			result = LexerResult{kind: Equals, len: 1}
		}
	case '+':
		if len(data) > 1 && isDigit(data[1]) {
			result, err = tokenizeNumber(data)
		} else {
			result = LexerResult{kind: Add, len: 1}
		}
	case '-':
		if len(data) > 1 && isDigit(data[1]) {
			result, err = tokenizeNumber(data)
		} else {
			result = LexerResult{kind: Subtract, len: 1}
		}
	case '*':
		result = LexerResult{kind: Multiply, len: 1}
	case '/':
		result = LexerResult{kind: Divide, len: 1}
	case '>':
		if len(data) > 1 && data[1] == '=' {
			result = LexerResult{kind: Gte, len: 2}

		} else {
			result = LexerResult{kind: Gt, len: 1}
		}
	case '<':
		if len(data) > 1 && data[1] == '=' {
			result = LexerResult{kind: Lte, len: 2}
		} else {
			result = LexerResult{kind: Lt, len: 1}
		}
	case '(':
		result = LexerResult{kind: OpenParen, len: 1}
	case ')':
		result = LexerResult{kind: CloseParen, len: 1}
	case '[':
		result = LexerResult{kind: OpenBracket, len: 1}
	case ']':
		result = LexerResult{kind: CloseBracket, len: 1}
	case ',':
		result = LexerResult{kind: Comma, len: 1}
	case '!':
		result = LexerResult{kind: Not, len: 1}
	case '"', '\'':
		result, err = tokenizeString(data, b)
	case '.':
		result, err = tokenizeSelectorPath(data)
	case 't', 'f':
		result, err = tokenizeBool(data)
	case '&':
		if len(data) > 1 && data[1] == '&' {
			result = LexerResult{kind: And, len: 2}
		} else {
			err = ErrUnsupportedCharacter{b: b}
		}
	case '|':
		if len(data) > 1 && data[1] == '|' {
			result = LexerResult{kind: Or, len: 2}
		} else {
			err = ErrUnsupportedCharacter{b: b}
		}
	case 'C':

		if len(data) > 2 && data[2] == 'N' {
			// can be one of CONTAINS, CONTAINS_ANY, CONTAINS_ALL
			if len(data) > 8 && data[8] == '_' {
				if len(data) > 10 && data[10] == 'N' {
					result, err = tokenizeKeyword(data, "CONTAINS_ANY", ContainsAny)
				} else {
					result, err = tokenizeKeyword(data, "CONTAINS_ALL", ContainsAll)
				}
			} else {
				result, err = tokenizeKeyword(data, "CONTAINS", Contains)
			}
		} else {
			result, err = tokenizeKeyword(data, "COERCE", Coerce)
		}

	case 'I':
		result, err = tokenizeKeyword(data, "IN", In)
	case 'S':
		result, err = tokenizeKeyword(data, "STARTSWITH", StartsWith)
	case 'E':
		result, err = tokenizeKeyword(data, "ENDSWITH", EndsWith)
	case 'B':
		result, err = tokenizeKeyword(data, "BETWEEN", Between)
	case 'N':
		result, err = tokenizeNull(data)
	case '_':
		result, err = tokenizeIdentifier(data)
	default:
		if isDigit(b) {
			result, err = tokenizeNumber(data)
		} else {
			err = ErrUnsupportedCharacter{b: b}
		}
	}

	return
}

func tokenizeIdentifier(data []byte) (result LexerResult, err error) {
	end := takeWhile(data, func(b byte) bool {
		return !isWhitespace(b) && b != ')' && b != ']'
	})
	// identifier must start and end with underscore
	if end > 0 && data[end-1] == '_' {
		result = LexerResult{
			kind: Identifier,
			len:  end,
		}
	} else {
		err = ErrInvalidIdentifier{s: string(data)}
	}
	return
}

func tokenizeNumber(data []byte) (result LexerResult, err error) {
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
		result = LexerResult{
			kind: Number,
			len:  end,
		}
	} else {
		err = ErrInvalidNumber{s: string(data)}
	}
	return
}

func tokenizeKeyword(data []byte, keyword string, kind TokenKind) (result LexerResult, err error) {
	end := takeWhile(data, func(b byte) bool {
		return !isWhitespace(b)
	})
	if end > 0 && len(data) > len(keyword) && string(data[:end]) == keyword {
		result = LexerResult{
			kind: kind,
			len:  end,
		}
	} else {
		err = ErrInvalidKeyword{s: string(data)}
	}
	return
}

func tokenizeNull(data []byte) (result LexerResult, err error) {
	end := takeWhile(data, func(b byte) bool {
		return isAlphabetical(b)
	})
	if end > 0 && string(data[:end]) == "NULL" {
		result = LexerResult{
			kind: Null,
			len:  end,
		}
	} else {
		err = ErrInvalidKeyword{s: string(data)}
	}
	return
}

func tokenizeBool(data []byte) (result LexerResult, err error) {
	end := takeWhile(data, func(b byte) bool {
		return isAlphabetical(b)
	})
	if end > 0 {
		switch string(data[:end]) {
		case "true":
			result = LexerResult{
				kind: BooleanTrue,
				len:  end,
			}
		case "false":
			result = LexerResult{
				kind: BooleanFalse,
				len:  end,
			}
		default:
			err = ErrInvalidBool{s: string(data)}
		}
	} else {
		err = ErrInvalidBool{s: string(data)}
	}
	return
}

func tokenizeSelectorPath(data []byte) (result LexerResult, err error) {
	end := takeWhile(data[1:], func(b byte) bool {
		return !isWhitespace(b) && b != ')' && b != ']'
	})
	if end > 0 {
		if len(data) > int(end) {
			end += 1
		}
		result = LexerResult{
			kind: SelectorPath,
			len:  end,
		}
	} else {
		err = ErrInvalidSelectorPath{s: string(data)}
	}
	return
}

func tokenizeString(data []byte, quote byte) (result LexerResult, err error) {
	var lastBackslash, endedWithTerminator bool

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
			endedWithTerminator = true
			return false
		default:
			return true
		}
	})

	if end > 0 {
		if endedWithTerminator {
			result = LexerResult{
				kind: QuotedString,
				len:  end + 2,
			}

		} else {
			err = ErrUnterminatedString{s: string(data)}
		}
	} else {
		if !endedWithTerminator || len(data) < 2 {
			err = ErrUnterminatedString{s: string(data)}
		} else {
			result = LexerResult{
				kind: QuotedString,
				len:  2,
			}
		}
	}
	return
}

/// Consumes bytes while a predicate evaluates to true.
func takeWhile(data []byte, pred func(byte) bool) (end uint16) {
	for _, b := range data {
		if !pred(b) {
			break
		}
		end++
	}
	return
}

// Tokenizer is a lexer for the KSQL expression syntax.
type Tokenizer struct {
	pos       uint32
	remaining []byte
}

func skipWhitespace(data []byte) uint16 {
	return takeWhile(data, func(b byte) bool {
		return isWhitespace(b)
	})
}

// NewTokenizer creates a new tokenizer for use
func NewTokenizer(src []byte) *Tokenizer {
	return &Tokenizer{
		pos:       0,
		remaining: src,
	}
}

func (t *Tokenizer) Next() (token Token, err error) {
	t.skipWhitespace()

	if len(t.remaining) == 0 {
		err = io.EOF
		return
	}
	return t.nextToken()
}

func (t *Tokenizer) skipWhitespace() {
	skipped := skipWhitespace(t.remaining)
	t.chomp(skipped)
}

func (t *Tokenizer) nextToken() (token Token, err error) {
	var result LexerResult
	result, err = tokenizeSingleToken(t.remaining)
	if err != nil {
		return
	}
	token = Token{
		start: t.pos,
		len:   result.len,
		kind:  result.kind,
	}
	t.chomp(result.len)
	return
}

func (t *Tokenizer) chomp(num uint16) {
	t.remaining = t.remaining[num:]
	t.pos += uint32(num)
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
