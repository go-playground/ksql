package ksql

import (
	"testing"
)

func BenchmarkLexingNumPlusNum(b *testing.B) {
	benchLexing(b, "1 + 1")
}

func BenchmarkLexingIdentNum(b *testing.B) {
	benchLexing(b, ".field1 + 1")
}

func BenchmarkLexingIdentIdent(b *testing.B) {
	benchLexing(b, ".field1 + .field2")
}

func BenchmarkLexingFNameLName(b *testing.B) {
	benchLexing(b, `.first_name + " " + .last_name`)
}

func BenchmarkLexingParenDiv(b *testing.B) {
	benchLexing(b, `(1 + 1) / 2`)
}

func BenchmarkLexingParenDivIdents(b *testing.B) {
	benchLexing(b, `(.field1 + .field2) / .field3`)
}

func BenchmarkLexingCompanyEmployees(b *testing.B) {
	benchLexing(b, `.properties.employees > 20`)
}

func BenchmarkLexingParenNot(b *testing.B) {
	benchLexing(b, `!(.f1 != .f2)`)
}

func BenchmarkLexingCoerceDateTimeSelector(b *testing.B) {
	benchLexing(b, `COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_`)
}

func BenchmarkLexingCoerceDateTimeSelectorMixed(b *testing.B) {
	benchLexing(b, `COERCE .dt1 _datetime_ == COERCE "2022-01-02" _datetime_`)
}

func BenchmarkLexingCoerceDateTimeSelectorConstant(b *testing.B) {
	benchLexing(b, `COERCE "2022-01-02" _datetime_ == COERCE "2022-01-02" _datetime_`)
}

func BenchmarkParsingNumPlusNum(b *testing.B) {
	benchParsing(b, "1 + 1")
}

func BenchmarkParsingIdentNum(b *testing.B) {
	benchParsing(b, ".field1 + 1")
}

func BenchmarkParsingIdentIdent(b *testing.B) {
	benchParsing(b, ".field1 + .field2")
}

func BenchmarkParsingFNameLName(b *testing.B) {
	benchParsing(b, `.first_name + " " + .last_name`)
}

func BenchmarkParsingParenDiv(b *testing.B) {
	benchParsing(b, `(1 + 1) / 2`)
}

func BenchmarkParsingParenDivIdents(b *testing.B) {
	benchParsing(b, `(.field1 + .field2) / .field3`)
}

func BenchmarkParsingCompanyEmployees(b *testing.B) {
	benchParsing(b, `.properties.employees > 20`)
}

func BenchmarkParsingParenNot(b *testing.B) {
	benchParsing(b, `!(.f1 != .f2)`)
}

func BenchmarkParsingCoerceDateTimeSelector(b *testing.B) {
	benchParsing(b, `COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_`)
}

func BenchmarkParsingCoerceDateTimeSelectorMixed(b *testing.B) {
	benchParsing(b, `COERCE .dt1 _datetime_ == COERCE "2022-01-02" _datetime_`)
}

func BenchmarkParsingCoerceDateTimeSelectorConstant(b *testing.B) {
	benchParsing(b, `COERCE "2022-01-02" _datetime_ == COERCE "2022-01-02" _datetime_`)
}

func BenchmarkExecutionNumPlusNum(b *testing.B) {
	benchExecution(b, "1 + 1 + 1 + 1 + 1", ``)
}

func BenchmarkExecutionIdentNum(b *testing.B) {
	benchExecution(b, ".field1 + 1", `{"field1":1}`)
}

func BenchmarkExecutionIdentIdent(b *testing.B) {
	benchExecution(b, ".field1 + .field2", `{"field1":1,"field2":1}`)
}

func BenchmarkExecutionFNameLName(b *testing.B) {
	benchExecution(b, `.first_name + " " + .last_name`, `{"first_name":"Joey","last_name":"Bloggs"}`)
}

func BenchmarkExecutionParenDiv(b *testing.B) {
	benchExecution(b, `(1 + 1) / 2`, ``)
}

func BenchmarkExecutionParenDivIdents(b *testing.B) {
	benchExecution(b, `(.field1 + .field2) / .field3`, `{"field1":1,"field2":1,"field3":2}`)
}

func BenchmarkExecutionCompanyEmployees(b *testing.B) {
	benchExecution(b, `.properties.employees > 20`, `{"name":"Company","properties":{"employees":50}}`)
}

func BenchmarkExecutionParenNot(b *testing.B) {
	benchExecution(b, `!(.f1 != .f2)`, `{"f1":true,"f2":false}`)
}

func BenchmarkExecutionCoerceDateTimeSelector(b *testing.B) {
	benchExecution(b, `COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_`, `{"dt1":"2022-01-02","dt2":"2022-01-02"}`)
}

func BenchmarkExecutionCoerceDateTimeSelectorMixed(b *testing.B) {
	benchExecution(b, `COERCE .dt1 _datetime_ == COERCE "2022-01-02" _datetime_`, `{"dt1":"2022-01-02"}`)
}

func BenchmarkExecutionCoerceDateTimeSelectorConstant(b *testing.B) {
	benchExecution(b, `COERCE "2022-01-02" _datetime_ == COERCE "2022-01-02" _datetime_`, ``)
}

func benchExecution(b *testing.B, expression, input string) {
	ex, err := Parse([]byte(expression))
	if err != nil {
		b.Fatal(err)
	}
	in := []byte(input)
	b.SetBytes(int64(len(in)))

	for i := 0; i < b.N; i++ {
		_, err := ex.Calculate(in)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchParsing(b *testing.B, expression string) {
	b.SetBytes(int64(len(expression)))

	for i := 0; i < b.N; i++ {
		_, err := Parse([]byte(expression))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchLexing(b *testing.B, expression string) {
	exp := []byte(expression)
	b.SetBytes(int64(len(expression)))

	for i := 0; i < b.N; i++ {
		_, err := collect(exp)
		if err != nil {
			b.Fatal(err)
		}
	}
}
