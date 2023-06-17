package ksql

import (
	"errors"
	"fmt"
	"github.com/go-playground/itertools"
	syncext "github.com/go-playground/pkg/v5/sync"
	optionext "github.com/go-playground/pkg/v5/values/option"
	resultext "github.com/go-playground/pkg/v5/values/result"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/araddon/dateparse"
	"github.com/tidwall/gjson"
)

var (
	// Coercions is a `map` of all coercions guarded by a Mutex for use allowing registration,
	// removal or even replacing of existing coercions.
	Coercions = syncext.NewRWMutex2(map[string]func(p *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error){
		"_datetime_": func(_ *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			expression = coerceDateTime{value: expression}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
		"_lowercase_": func(_ *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			expression = coerceLowercase{value: expression}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
		"_string_": func(_ *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			expression = coerceString{value: expression}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
		"_number_": func(_ *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			expression = coerceNumber{value: expression}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
		"_uppercase_": func(_ *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			expression = coerceUppercase{value: expression}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
		"_title_": func(_ *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			expression = coerceTitle{value: expression}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
		"_substr_": func(p *Parser, constEligible bool, expression Expression) (stillConstEligible bool, e Expression, err error) {
			// get substring info, expect the format to be _substr_[Start:end]

			leftBracket := p.Tokenizer.Next()
			if leftBracket.IsNone() {
				return false, nil, ErrCustom{S: "Expected [ after _substr_"}
			} else if leftBracket.Unwrap().IsErr() {
				return false, nil, ErrInvalidCoerce{Err: leftBracket.Unwrap().Err()}
			} else if leftBracket.Unwrap().Unwrap().Kind != OpenBracket {
				return false, nil, ErrCustom{S: "Expected [ after _substr_"}
			}

			// number or colon
			var startIndex optionext.Option[int]
			token := p.Tokenizer.Next()
			if token.IsNone() {
				return false, nil, ErrCustom{S: "Expected number or colon after _substr_["}
			} else if token.Unwrap().IsErr() {
				return false, nil, ErrInvalidCoerce{Err: token.Unwrap().Err()}
			} else {
				token := token.Unwrap().Unwrap()
				start := int(token.Start)
				switch token.Kind {
				case Colon:
				case Number:
					i64, err := strconv.ParseInt(string(p.Exp[start:start+int(token.Len)]), 10, 64)
					if err != nil {
						return false, nil, err
					}
					startIndex = optionext.Some(int(i64))
				default:
					return false, nil, ErrCustom{S: fmt.Sprintf("Expected number after _substr_[ but got %s", string(p.Exp[start:start+int(token.Len)]))}
				}
			}

			// parse colon if not already
			if startIndex.IsSome() {
				colon := p.Tokenizer.Next()
				if colon.IsNone() {
					return false, nil, ErrCustom{S: "Expected : after _substr_[n"}
				} else if colon.Unwrap().IsErr() {
					return false, nil, ErrInvalidCoerce{Err: colon.Unwrap().Err()}
				} else if colon.Unwrap().Unwrap().Kind != Colon {
					return false, nil, ErrCustom{S: "Expected : after _substr_[n"}
				}
			}

			// number or end bracket
			var endIndex optionext.Option[int]
			token = p.Tokenizer.Next()
			if token.IsNone() {
				return false, nil, ErrCustom{S: "Expected number or ] after _substr_["}
			} else if token.Unwrap().IsErr() {
				return false, nil, ErrInvalidCoerce{Err: token.Unwrap().Err()}
			} else {
				token := token.Unwrap().Unwrap()
				start := int(token.Start)
				switch token.Kind {
				case CloseBracket:
				case Number:
					i64, err := strconv.ParseInt(string(p.Exp[start:start+int(token.Len)]), 10, 64)
					if err != nil {
						return false, nil, err
					}
					endIndex = optionext.Some(int(i64))
				default:
					return false, nil, ErrCustom{S: fmt.Sprintf("Expected number after _substr_[n: but got %s", string(p.Exp[start:start+int(token.Len)]))}
				}
			}

			// parse close bracket if not already
			if endIndex.IsSome() {
				rightBracket := p.Tokenizer.Next()
				if rightBracket.IsNone() {
					return false, nil, ErrCustom{S: "Expected ] after _substr_[n:n"}
				} else if rightBracket.Unwrap().IsErr() {
					return false, nil, ErrInvalidCoerce{Err: rightBracket.Unwrap().Err()}
				} else if rightBracket.Unwrap().Unwrap().Kind != CloseBracket {
					return false, nil, ErrCustom{S: "Expected ] after _substr_[n:n"}
				}
			}

			switch {
			case startIndex.IsSome() && endIndex.IsSome() && startIndex.Unwrap() > endIndex.Unwrap():
				return false, nil, ErrCustom{S: fmt.Sprintf("Start index %d cannot be greater than end index %d", startIndex.Unwrap(), endIndex.Unwrap())}
			case startIndex.IsNone() && endIndex.IsNone():
				return false, nil, ErrCustom{S: "Start and end index for substr cannot both be None"}
			}

			expression = coerceSubstr{
				value: expression,
				start: startIndex,
				end:   endIndex,
			}
			if constEligible {
				value, err := expression.Calculate([]byte{})
				if err != nil {
					return false, nil, err
				}
				return constEligible, coercedConstant{value: value}, nil
			} else {
				return false, expression, nil
			}
		},
	})
)

// Expression Represents a stateless parsed expression that can be applied to JSON data.
type Expression interface {

	// Calculate will execute the parsed expression and apply it against the supplied data.
	//
	// # Errors
	//
	// Will return `Err` if the expression cannot be applied to the supplied data due to invalid
	// data type comparisons.
	Calculate(src []byte) (any, error)
}

// Parse lex's' the provided expression and returns an Expression to be used/applied to data.
func Parse(expression []byte) (Expression, error) {
	p := Parser{
		Exp:       expression,
		Tokenizer: itertools.Iter[resultext.Result[Token, error]](NewTokenizer(expression)).Peekable(),
	}

	result, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("no expression results found")
	}
	return result, nil
}

// Parser parses and returns a supplied expression
type Parser struct {
	Exp       []byte
	Tokenizer itertools.PeekableIterator[resultext.Result[Token, error]]
}

func (p *Parser) parseExpression() (current Expression, err error) {

	for {
		next := p.Tokenizer.Next()
		if next.IsNone() {
			return current, nil
		}
		result := next.Unwrap()
		if result.IsErr() {
			return nil, result.Err()
		}
		token := result.Unwrap()

		if current == nil {
			// look for nextToken value
			current, err = p.parseValue(token)
			if err != nil {
				return nil, err
			}

		} else {
			if token.Kind == CloseParen {
				return current, nil
			}
			// look for nextToken operation
			current, err = p.parseOperation(token, current)
			if err != nil {
				return nil, err
			}
		}
	}
}

func (p *Parser) parseValue(token Token) (Expression, error) {
	switch token.Kind {
	case OpenBracket:
		arr := make([]Expression, 0, 2)

	FOR:
		for {
			next := p.Tokenizer.Next()
			if next.IsNone() {
				return nil, errors.New("unclosed Array '['")
			}
			result := next.Unwrap()
			if result.IsErr() {
				return nil, result.Err()
			}
			token := result.Unwrap()

			switch token.Kind {
			case CloseBracket:
				break FOR
			case Comma:
				continue
			default:
				value, err := p.parseValue(token)
				if err != nil {
					return nil, err
				}
				arr = append(arr, value)
			}
			if token.Kind == CloseBracket {
				break
			}
		}

		return array{vec: arr}, nil

	case OpenParen:
		expression, err := p.parseExpression()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.New("expression after open parenthesis '(' ends unexpectedly")
			}
			return nil, err
		}
		return expression, nil

	case SelectorPath:
		start := int(token.Start)
		return selectorPath{
			s: string(p.Exp[start+1 : start+int(token.Len)]),
		}, nil

	case QuotedString:
		start := int(token.Start)
		return str{
			s: string(p.Exp[start+1 : start+int(token.Len)-1]),
		}, nil

	case Number:
		start := int(token.Start)
		f64, err := strconv.ParseFloat(string(p.Exp[start:start+int(token.Len)]), 64)
		if err != nil {
			return nil, err
		}
		return num{
			n: f64,
		}, nil

	case BooleanTrue:
		return boolean{b: true}, nil

	case BooleanFalse:
		return boolean{b: false}, nil

	case Null:
		return null{}, nil

	case Coerce:
		// COERCE <expression> _<datatype>_
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		var constEligible bool
		switch nextToken.Kind {
		case QuotedString, Number, BooleanTrue, BooleanFalse, Null:
			constEligible = true
		}

		expression, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}

		for {
			next := p.Tokenizer.Next()
			if next.IsNone() {
				return nil, errors.New("no identifier after value for: COERCE")
			}
			result := next.Unwrap()
			if result.IsErr() {
				return nil, result.Err()
			}
			identifierToken := result.Unwrap()
			start := int(identifierToken.Start)
			identifier := string(p.Exp[start : start+int(identifierToken.Len)])

			if identifierToken.Kind != Identifier {
				return nil, fmt.Errorf("COERCE missing data type identifier, found instead: %s", identifier)
			}

			guard := Coercions.RLock()
			fn, found := guard.T[identifier]
			guard.RUnlock()

			if found {
				constEligible, expression, err = fn(p, constEligible, expression)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("invalid COERCE data type '%s'", identifier)
			}

			nextPeeked := p.Tokenizer.Peek()
			if nextPeeked.IsSome() && nextPeeked.Unwrap().IsOk() && nextPeeked.Unwrap().Unwrap().Kind == Comma {
				_ = p.Tokenizer.Next() // consume peeked comma
				continue
			}
			break
		}
		return expression, nil

	case Not:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		value, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return not{value: value}, nil

	default:
		return nil, fmt.Errorf("token is not a valid value: %v", token)
	}
}

func (p *Parser) parseOperation(token Token, current Expression) (Expression, error) {
	switch token.Kind {
	case Add:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return add{
			left:  current,
			right: right,
		}, nil

	case Subtract:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return sub{
			left:  current,
			right: right,
		}, nil

	case Multiply:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return multi{
			left:  current,
			right: right,
		}, nil

	case Divide:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return div{
			left:  current,
			right: right,
		}, nil

	case Equals:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return eq{
			left:  current,
			right: right,
		}, nil

	case Gt:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return gt{
			left:  current,
			right: right,
		}, nil

	case Gte:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return gte{
			left:  current,
			right: right,
		}, nil

	case Lt:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return lt{
			left:  current,
			right: right,
		}, nil

	case Lte:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return lte{
			left:  current,
			right: right,
		}, nil

	case Or:
		right, err := p.parseExpression()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.New("expression after or '||' ends unexpectedly")
			}
			return nil, err
		}
		return or{
			left:  current,
			right: right,
		}, nil

	case And:
		right, err := p.parseExpression()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.New("expression after or '&&' ends unexpectedly")
			}
			return nil, err
		}
		return and{
			left:  current,
			right: right,
		}, nil

	case StartsWith:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return startsWith{
			left:  current,
			right: right,
		}, nil

	case EndsWith:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return endsWith{
			left:  current,
			right: right,
		}, nil

	case In:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return in{
			left:  current,
			right: right,
		}, nil

	case Contains:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return contains{
			left:  current,
			right: right,
		}, nil

	case ContainsAny:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return containsAny{
			left:  current,
			right: right,
		}, nil

	case ContainsAll:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(nextToken)
		if err != nil {
			return nil, err
		}
		return containsAll{
			left:  current,
			right: right,
		}, nil

	case Between:
		lhsToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		left, err := p.parseValue(lhsToken)
		if err != nil {
			return nil, err
		}

		rhsToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		right, err := p.parseValue(rhsToken)
		if err != nil {
			return nil, err
		}

		return between{
			left:  left,
			right: right,
			value: current,
		}, nil

	case Not:
		nextToken, err := p.nextOperatorToken(token)
		if err != nil {
			return nil, err
		}
		value, err := p.parseOperation(nextToken, current)
		if err != nil {
			return nil, err
		}
		return not{
			value: value,
		}, nil

	case CloseBracket:
		return current, nil

	default:
		return nil, fmt.Errorf("invalid operation: %v", token)
	}
}

func (p *Parser) nextOperatorToken(operationToken Token) (token Token, err error) {
	next := p.Tokenizer.Next()
	if next.IsNone() {
		start := int(operationToken.Start)
		err = fmt.Errorf("no value found after operation: %s", string(p.Exp[start:start+int(operationToken.Len)]))
		return
	}
	result := next.Unwrap()
	if result.IsErr() {
		err = result.Err()
		return
	}
	return result.Unwrap(), nil
}

var _ Expression = (*between)(nil)

type between struct {
	left  Expression
	right Expression
	value Expression
}

func (b between) Calculate(src []byte) (any, error) {
	left, err := b.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := b.right.Calculate(src)
	if err != nil {
		return nil, err
	}
	value, err := b.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	// fast path, if any are nil/null there's no way to actually do the BETWEEN comparison
	if left == nil || right == nil || value == nil {
		return false, nil
	}

	leftType := reflect.TypeOf(left)
	if !(leftType == reflect.TypeOf(right) && reflect.TypeOf(value) == leftType) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s < %s", left, right)}
	}

	switch v := value.(type) {
	case string:
		return v > left.(string) && v < right.(string), nil
	case float64:
		return v > left.(float64) && v < right.(float64), nil
	case time.Time:
		return v.After(left.(time.Time)) && v.Before(right.(time.Time)), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s < %s", left, right)}
	}
}

var _ Expression = (*coerceNumber)(nil)

type coerceNumber struct {
	value Expression
}

func (c coerceNumber) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a number", value)}
		}
		return f, nil

	case float64:
		return v, nil

	case bool:
		if v {
			return 1.0, nil
		} else {
			return 0.0, nil
		}

	case time.Time:
		return float64(v.UnixNano()), nil

	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a number", value)}
	}
}

var _ Expression = (*coerceString)(nil)

type coerceString struct {
	value Expression
}

func (c coerceString) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case nil:
		return "null", nil
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case time.Time:
		return v.Format(time.RFC3339Nano), nil
	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a string", value)}
	}
}

var _ Expression = (*coerceLowercase)(nil)

type coerceLowercase struct {
	value Expression
}

func (c coerceLowercase) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		return strings.ToLower(v), nil
	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a lowescase", value)}
	}
}

var _ Expression = (*coerceUppercase)(nil)

type coerceUppercase struct {
	value Expression
}

func (c coerceUppercase) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		return strings.ToUpper(v), nil
	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a uppercase", value)}
	}
}

var _ Expression = (*coerceTitle)(nil)

type coerceTitle struct {
	value Expression
}

func (c coerceTitle) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		r, size := utf8.DecodeRuneInString(v)
		if size == 0 {
			return v, nil
		}
		return string(unicode.ToUpper(r)) + strings.ToLower(v[1:]), nil
	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a uppercase", value)}
	}
}

var _ Expression = (*coerceDateTime)(nil)

type coerceDateTime struct {
	value Expression
}

func (c coerceDateTime) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		t, err := dateparse.ParseAny(v)
		if err != nil {
			// don't return error at runtime but null same as not found
			// which will fail equality checks and alike which is the desired behaviour.
			return nil, nil
		}
		return t, nil
	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v to a DateTime", value)}
	}
}

var _ Expression = (*coercedConstant)(nil)

type coercedConstant struct {
	value any
}

func (c coercedConstant) Calculate(_ []byte) (any, error) {
	return c.value, nil
}

var _ Expression = (*null)(nil)

type null struct {
}

func (bn null) Calculate(_ []byte) (any, error) {
	return nil, nil
}

var _ Expression = (*boolean)(nil)

type boolean struct {
	b bool
}

func (b boolean) Calculate(_ []byte) (any, error) {
	return b.b, nil
}

var _ Expression = (*num)(nil)

type num struct {
	n float64
}

func (n num) Calculate(_ []byte) (any, error) {
	return n.n, nil
}

var _ Expression = (*str)(nil)

type str struct {
	s string
}

func (s str) Calculate(_ []byte) (any, error) {
	return s.s, nil
}

var _ Expression = (*selectorPath)(nil)

type selectorPath struct {
	s string
}

func (i selectorPath) Calculate(src []byte) (any, error) {
	return gjson.GetBytes(src, i.s).Value(), nil
}

var _ Expression = (*add)(nil)

type add struct {
	left  Expression
	right Expression
}

func (a add) Calculate(src []byte) (any, error) {
	left, err := a.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := a.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		if left != nil && right == nil {
			switch left.(type) {
			case string, float64:
				return left, nil
			}
		} else if right != nil && left == nil {
			switch right.(type) {
			case string, float64:
				return right, nil
			}
		}
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s + %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return l + right.(string), nil
	case float64:
		return l + right.(float64), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s + %s", left, right)}
	}
}

var _ Expression = (*sub)(nil)

type sub struct {
	left  Expression
	right Expression
}

func (s sub) Calculate(src []byte) (any, error) {
	left, err := s.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := s.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s - %s", left, right)}
	}

	switch l := left.(type) {
	case float64:
		return l - right.(float64), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s - %s", left, right)}
	}
}

var _ Expression = (*multi)(nil)

type multi struct {
	left  Expression
	right Expression
}

func (m multi) Calculate(src []byte) (any, error) {
	left, err := m.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := m.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s * %s", left, right)}
	}

	switch l := left.(type) {
	case float64:
		return l * right.(float64), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s * %s", left, right)}
	}
}

var _ Expression = (*div)(nil)

type div struct {
	left  Expression
	right Expression
}

func (d div) Calculate(src []byte) (any, error) {
	left, err := d.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := d.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s / %s", left, right)}
	}

	switch l := left.(type) {
	case float64:
		return l / right.(float64), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s / %s", left, right)}
	}
}

var _ Expression = (*eq)(nil)

type eq struct {
	left  Expression
	right Expression
}

func (e eq) Calculate(src []byte) (any, error) {
	left, err := e.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := e.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	return reflect.DeepEqual(left, right), nil
}

var _ Expression = (*gt)(nil)

type gt struct {
	left  Expression
	right Expression
}

func (g gt) Calculate(src []byte) (any, error) {
	left, err := g.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := g.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s > %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return l > right.(string), nil
	case float64:
		return l > right.(float64), nil
	case time.Time:
		return l.After(right.(time.Time)), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s > %s", left, right)}
	}
}

var _ Expression = (*gte)(nil)

type gte struct {
	left  Expression
	right Expression
}

func (g gte) Calculate(src []byte) (any, error) {
	left, err := g.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := g.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s >= %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return l >= right.(string), nil
	case float64:
		return l >= right.(float64), nil
	case time.Time:
		r := right.(time.Time)
		return l.After(r) || l.Equal(r), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s >= %s", left, right)}
	}
}

var _ Expression = (*lt)(nil)

type lt struct {
	left  Expression
	right Expression
}

func (l lt) Calculate(src []byte) (any, error) {
	left, err := l.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := l.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s < %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return l < right.(string), nil
	case float64:
		return l < right.(float64), nil
	case time.Time:
		return l.Before(right.(time.Time)), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s < %s", left, right)}
	}
}

var _ Expression = (*lte)(nil)

type lte struct {
	left  Expression
	right Expression
}

func (l lte) Calculate(src []byte) (any, error) {
	left, err := l.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := l.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s <= %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return l <= right.(string), nil
	case float64:
		return l <= right.(float64), nil
	case time.Time:
		r := right.(time.Time)
		return l.Before(r) || l.Equal(r), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s <= %s", left, right)}
	}
}

var _ Expression = (*not)(nil)

type not struct {
	value Expression
}

func (n not) Calculate(src []byte) (any, error) {
	value, err := n.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch t := value.(type) {
	case bool:
		return !t, nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s for !", value)}
	}
}

var _ Expression = (*or)(nil)

type or struct {
	left  Expression
	right Expression
}

func (o or) Calculate(src []byte) (any, error) {
	left, err := o.left.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch t := left.(type) {
	case bool:
		if t {
			return true, nil
		}
	}

	right, err := o.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s || %s", left, right)}
	}

	switch t := left.(type) {
	case bool:
		return t || right.(bool), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s || %s !", left, right)}
	}
}

var _ Expression = (*and)(nil)

type and struct {
	left  Expression
	right Expression
}

func (a and) Calculate(src []byte) (any, error) {
	left, err := a.left.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch t := left.(type) {
	case bool:
		if !t {
			return false, nil
		}
	default:
		return false, nil
	}

	right, err := a.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s && %s", left, right)}
	}

	switch t := left.(type) {
	case bool:
		return t && right.(bool), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s && %s !", left, right)}
	}
}

var _ Expression = (*contains)(nil)

type contains struct {
	left  Expression
	right Expression
}

func (c contains) Calculate(src []byte) (any, error) {
	left, err := c.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := c.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	leftTypeOf := reflect.TypeOf(left)

	if leftTypeOf != reflect.TypeOf(right) && leftTypeOf.Kind() != reflect.Slice {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return strings.Contains(l, right.(string)), nil
	case []any:
		for _, v := range l {
			if reflect.DeepEqual(v, right) {
				return true, nil
			}
		}
		return false, nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS %s !", left, right)}
	}
}

var _ Expression = (*containsAny)(nil)

type containsAny struct {
	left  Expression
	right Expression
}

func (c containsAny) Calculate(src []byte) (any, error) {
	left, err := c.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := c.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case string:

		switch r := right.(type) {
		case string:

			// betting that lists are short and so less expensive than iterating one to create a hash set
			for _, c := range r {
				for _, c2 := range l {
					if c == c2 {
						return true, nil
					}
				}
			}

		case []any:
			for _, v := range r {
				s, ok := v.(string)
				if !ok {
					continue
				}
				if strings.Contains(l, s) {
					return true, nil
				}
			}
			return false, nil

		default:
			return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS_ANY %s", left, right)}
		}

	case []any:
		switch r := right.(type) {
		case []any:
			// betting that lists are short and so less expensive than iterating one to create a hash set
			for _, rv := range r {
				for _, lv := range l {
					if reflect.DeepEqual(rv, lv) {
						return true, nil
					}
				}
			}

		case string:
			// betting that lists are short and so less expensive than iterating one to create a hash set
			for _, c := range r {
				for _, v := range l {
					if reflect.DeepEqual(string(c), v) {
						return true, nil
					}
				}
			}

		default:
			return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS_ANY %s", left, right)}
		}

	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS_ANY %s !", left, right)}
	}
	return false, nil
}

var _ Expression = (*containsAll)(nil)

type containsAll struct {
	left  Expression
	right Expression
}

func (c containsAll) Calculate(src []byte) (any, error) {
	left, err := c.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := c.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case string:
		switch r := right.(type) {
		case string:
			// betting that lists are short and so less expensive than iterating one to create a hash set
		OUTER1:
			for _, c := range r {
				for _, c2 := range l {
					if c == c2 {
						continue OUTER1
					}
				}
				return false, nil
			}

		case []any:
			for _, v := range r {
				s, ok := v.(string)
				if !ok || !strings.Contains(l, s) {
					return false, nil
				}
			}
			return true, nil

		default:
			return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS_ALL %s", left, right)}
		}

	case []any:
		switch r := right.(type) {
		case []any:
			// betting that lists are short and so less expensive than iterating one to create a hash set
		OUTER3:
			for _, rv := range r {
				for _, lv := range l {
					if reflect.DeepEqual(rv, lv) {
						continue OUTER3
					}
				}
				return false, nil
			}

		case string:
			// betting that lists are short and so less expensive than iterating one to create a hash set
		OUTER4:
			for _, c := range r {
				for _, v := range l {
					if reflect.DeepEqual(string(c), v) {
						continue OUTER4
					}
				}
				return false, nil
			}
		default:
			return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS_ALL %s", left, right)}
		}

	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS_ALL %s !", left, right)}
	}
	return true, nil
}

var _ Expression = (*startsWith)(nil)

type startsWith struct {
	left  Expression
	right Expression
}

func (s startsWith) Calculate(src []byte) (any, error) {
	left, err := s.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := s.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s STARTSWITH %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return strings.HasPrefix(l, right.(string)), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s STARTSWITH %s !", left, right)}
	}
}

var _ Expression = (*endsWith)(nil)

type endsWith struct {
	left  Expression
	right Expression
}

func (e endsWith) Calculate(src []byte) (any, error) {
	left, err := e.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := e.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s ENDSWITH %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return strings.HasSuffix(l, right.(string)), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s ENDSWITH %s !", left, right)}
	}
}

var _ Expression = (*in)(nil)

type in struct {
	left  Expression
	right Expression
}

func (i in) Calculate(src []byte) (any, error) {
	left, err := i.left.Calculate(src)
	if err != nil {
		return nil, err
	}
	right, err := i.right.Calculate(src)
	if err != nil {
		return nil, err
	}

	arr, ok := right.([]any)
	if !ok {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s IN %s !", left, right)}
	}
	for _, v := range arr {
		if left == v {
			return true, nil
		}
	}
	return false, nil
}

var _ Expression = (*array)(nil)

type array struct {
	vec []Expression
}

func (a array) Calculate(src []byte) (any, error) {
	arr := make([]any, 0, len(a.vec))
	for _, v := range a.vec {
		res, err := v.Calculate(src)
		if err != nil {
			return nil, err
		}
		arr = append(arr, res)
	}
	return arr, nil
}

type coerceSubstr struct {
	value Expression
	start optionext.Option[int]
	end   optionext.Option[int]
}

func (c coerceSubstr) Calculate(src []byte) (any, error) {
	value, err := c.value.Calculate(src)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		switch {
		case c.start.IsSome() && c.end.IsSome():
			start, end := c.start.Unwrap(), c.end.Unwrap()
			if start < 0 || start > len(v) || end < 0 || end > len(v) {
				return nil, nil
			}
			return v[start:end], nil

		case c.start.IsSome() && c.end.IsNone():
			start := c.start.Unwrap()
			if start < 0 || start > len(v) {
				return nil, nil
			}
			return v[start:], nil

		case c.start.IsNone() && c.end.IsSome():
			end := c.end.Unwrap()
			if end < 0 || end > len(v) {
				return nil, nil
			}
			return v[:end], nil

		default:
			return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v for substr, [%v:%v]", value, c.start, c.end)}
		}

	default:
		return nil, ErrUnsupportedCoerce{s: fmt.Sprintf("unsupported type COERCE for value: %v for substr", value)}
	}
}
