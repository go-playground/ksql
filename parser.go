package ksql

type Expression interface {

	// Calculate will execute the parsed expression and apply it against the supplied data.
	//
	// # Errors
	//
	// Will return `Err` if the expression cannot be applied to the supplied data due to invalid
	// data type comparisons.
	Calculate(src []byte) (any, error)
}
