package main

import (
	"fmt"
	"strings"

	"github.com/go-playground/ksql"
)

type Star struct {
	expression ksql.Expression
}

func (s *Star) Calculate(json []byte) (interface{}, error) {
	inner, err := s.expression.Calculate(json)
	if err != nil {
		return nil, err
	}

	switch t := inner.(type) {
	case string:
		return strings.Repeat("*", len(t)), nil
	default:
		return nil, fmt.Errorf("cannot star value %v", inner)
	}
}

func main() {
	// Add custom coercion to the parser.
	// REMEMBER: coercions start and end with an _(underscore).
	guard := ksql.Coercions.Lock()
	guard.T["_star_"] = func(constEligible bool, expression ksql.Expression) (stillConstEligible bool, e ksql.Expression, err error) {
		return constEligible, &Star{expression}, nil
	}
	guard.Unlock()

	expression := []byte(`COERCE "My Name" _star_`)
	input := []byte(`{}`)
	ex, err := ksql.Parse(expression)
	if err != nil {
		panic(err)
	}

	result, err := ex.Calculate(input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", result)
}
