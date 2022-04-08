package ksql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

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
	tokens, err := Tokenize(expression)
	if err != nil {
		return nil, err
	}

	pos := new(int)
	result, err := parseValue(tokens, pos)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("no expression results found")
	}
	return result, nil
}

func parseValue(tokens []Token, pos *int) (Expression, error) {
	if *pos > len(tokens)-1 {
		return nil, nil
	}
	tok := tokens[*pos]
	*pos += 1

	switch tok.kind {
	case Identifier:
		return parseOp(ident{s: tok.value.(string)}, tokens, pos)
	case String:
		return parseOp(str{s: tok.value.(string)}, tokens, pos)
	case Number:
		return parseOp(num{n: tok.value.(float64)}, tokens, pos)
	case Boolean:
		return parseOp(boolean{b: tok.value.(bool)}, tokens, pos)
	case Null:
		return parseOp(null{}, tokens, pos)
	case Not:
		v, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, errors.New("no identifier after !")
		}
		return parseOp(not{value: v}, tokens, pos)
	case OpenBracket:
		var arr []Expression
		for {
			v, err := parseValue(tokens, pos)
			if err != nil {
				return nil, err
			}
			if v == nil {
				break
			}
			arr = append(arr, v)
		}
		return parseOp(array{
			vec: arr,
		}, tokens, pos)
	case Comma:
		v, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, fmt.Errorf("value required after comma: %s", tok.value)
		}
		return v, nil
	case OpenParen:
		v, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, errors.New("no value between ()")
		}
		return parseOp(v, tokens, pos)
	case CloseParen:
		return nil, errors.New("no value between ()")
	case CloseBracket:
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid value %d %s", tok, tok.value)
	}
}

func parseOp(value Expression, tokens []Token, pos *int) (Expression, error) {
	if *pos > len(tokens)-1 {
		return value, nil
	}
	tok := tokens[*pos]
	*pos += 1

	switch tok.kind {
	case In:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after IN")
		}
		return in{
			left:  value,
			right: right,
		}, nil
	case Contains:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after CONTAINS")
		}
		return contains{
			left:  value,
			right: right,
		}, nil
	case StartsWith:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after STARTSWITH")
		}
		return startsWith{
			left:  value,
			right: right,
		}, nil
	case EndsWith:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after ENDSWITH")
		}
		return endsWith{
			left:  value,
			right: right,
		}, nil
	case And:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after AND")
		}
		return and{
			left:  value,
			right: right,
		}, nil
	case Or:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after OR")
		}
		return or{
			left:  value,
			right: right,
		}, nil
	case Gt:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after >")
		}
		return gt{
			left:  value,
			right: right,
		}, nil
	case Gte:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after >=")
		}
		return gte{
			left:  value,
			right: right,
		}, nil
	case Lt:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after <")
		}
		return lt{
			left:  value,
			right: right,
		}, nil
	case Lte:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after <=")
		}
		return lte{
			left:  value,
			right: right,
		}, nil
	case Equals:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after ==")
		}
		return eq{
			left:  value,
			right: right,
		}, nil
	case Add:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after +")
		}
		return add{
			left:  value,
			right: right,
		}, nil
	case Subtract:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after -")
		}
		return sub{
			left:  value,
			right: right,
		}, nil
	case Multiply:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after *")
		}
		return multi{
			left:  value,
			right: right,
		}, nil
	case Divide:
		right, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, errors.New("no value after /")
		}
		return div{
			left:  value,
			right: right,
		}, nil
	case Not:
		op, err := parseOp(value, tokens, pos)
		if err != nil {
			return nil, err
		}
		if op == nil {
			return nil, errors.New("no operator after !")
		}
		return not{
			value: op,
		}, nil
	case OpenParen:
		op, err := parseValue(tokens, pos)
		if err != nil {
			return nil, err
		}
		if op == nil {
			return nil, errors.New("no value between ()")
		}
		return parseOp(op, tokens, pos)
	case CloseParen, CloseBracket:
		return value, nil
	default:
		return nil, fmt.Errorf("invalid token after ident %d %s", tok.kind, tok.value)
	}
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

var _ Expression = (*ident)(nil)

type ident struct {
	s string
}

func (i ident) Calculate(src []byte) (any, error) {
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

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS %s", left, right)}
	}

	switch l := left.(type) {
	case string:
		return strings.Contains(l, right.(string)), nil
	default:
		return nil, ErrUnsupportedTypeComparison{s: fmt.Sprintf("%s CONTAINS %s !", left, right)}
	}
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
