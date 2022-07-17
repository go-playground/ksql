package ksql

import (
	"testing"
)

func BenchmarkNumPlusNum(b *testing.B) {
	bench(b, "1 + 1", ``)
}

func BenchmarkIdentNum(b *testing.B) {
	bench(b, ".field1 + 1", `{"field1":1}`)
}

func BenchmarkIdentIdent(b *testing.B) {
	bench(b, ".field1 + .field2", `{"field1":1,"field2":1}`)
}

func BenchmarkFNameLName(b *testing.B) {
	bench(b, `.first_name + " " + .last_name`, `{"first_name":"Joey","last_name":"Bloggs"}`)
}

func BenchmarkParenDiv(b *testing.B) {
	bench(b, `(1 + 1) / 2`, ``)
}

func BenchmarkParenDivIdents(b *testing.B) {
	bench(b, `(.field1 + .field2) / .field3`, `{"field1":1,"field2":1,"field3":2}`)
}

func BenchmarkCompanyEmployees(b *testing.B) {
	bench(b, `.properties.employees > 20`, `{"name":"Company","properties":{"employees":50}}`)
}

func BenchmarkParenNot(b *testing.B) {
	bench(b, `!(.f1 != .f2)`, `{"f1":true,"f2":false}`)
}

func BenchmarkCoerceDateTimeSelector(b *testing.B) {
	bench(b, `COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_`, `{"dt1":"2022-01-02","dt2":"2022-01-02"}`)
}

func BenchmarkCoerceDateTimeSelectorMixed(b *testing.B) {
	bench(b, `COERCE .dt1 _datetime_ == COERCE "2022-01-02" _datetime_`, `{"dt1":"2022-01-02"}`)
}

func BenchmarkCoerceDateTimeSelectorConstant(b *testing.B) {
	bench(b, `COERCE "2022-01-02" _datetime_ == COERCE "2022-01-02" _datetime_`, ``)
}

func bench(b *testing.B, expression, input string) {
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
