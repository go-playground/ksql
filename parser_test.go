package ksql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		name     string
		exp      string
		src      string
		expected any
		err      error
		parseErr error
	}{
		{
			name:     "selectorPath + selectorPath",
			exp:      ".f1 + .f2",
			src:      `{"f1":1,"f2":1}`,
			expected: float64(2),
		},
		{
			name:     "first_name + last_name",
			exp:      `.field1 + " " + .field2`,
			src:      `{"field1":"Dean","field2":"Karn"}`,
			expected: "Dean Karn",
		},
		{
			name:     "selectorPath + selectorPath num",
			exp:      `.field1 + .field2`,
			src:      `{"field1":10.1,"field2":23.23}`,
			expected: 33.33,
		},
		{
			name:     "selectorPath - selectorPath num",
			exp:      `.field2 - .field1`,
			src:      `{"field1":10.1,"field2":23.23}`,
			expected: 13.13,
		},
		{
			name:     "selectorPath * selectorPath num",
			exp:      `.field2 * .field1`,
			src:      `{"field1":11.1,"field2":3}`,
			expected: 33.3,
		},
		{
			name:     "selectorPath / selectorPath num",
			exp:      `.field2 / .field1`,
			src:      `{"field1":3,"field2":33.3}`,
			expected: 11.1,
		},
		{
			name:     "num + num",
			exp:      `11.1 + 22.2`,
			expected: 33.3,
		},
		{
			name:     "selectorPath + num",
			exp:      `11.1 + .field1`,
			src:      `{"field1":3,"field2":33.3}`,
			expected: 14.1,
		},
		{
			name:     "selectorPath == num false",
			exp:      `11.1 == .field1`,
			src:      `{"field1":3,"field2":33.3}`,
			expected: false,
		},
		{
			name:     "selectorPath == num true",
			exp:      `11.1 == .field1`,
			src:      `{"field1":11.1,"field2":33.3}`,
			expected: true,
		},
		{
			name:     "selectorPath > num false",
			exp:      `11.1 > .field1`,
			src:      `{"field1":11.1,"field2":33.3}`,
			expected: false,
		},
		{
			name:     "selectorPath > num true",
			exp:      `11.1 > .field1`,
			src:      `{"field1":11.0,"field2":33.3}`,
			expected: true,
		},
		{
			name:     "selectorPath >= num false",
			exp:      `11.1 >= .field1`,
			src:      `{"field1":12.0,"field2":33.3}`,
			expected: false,
		},
		{
			name:     "selectorPath >= num true",
			exp:      `11.1 >= .field1`,
			src:      `{"field1":11.0,"field2":33.3}`,
			expected: true,
		},
		{
			name:     "bool true",
			exp:      `true == true`,
			expected: true,
		},
		{
			name:     "bool false",
			exp:      `false == true`,
			expected: false,
		},
		{
			name:     "null eq",
			exp:      `NULL == NULL`,
			expected: true,
		},
		{
			name:     "or true 1",
			exp:      `false || true`,
			expected: true,
		},
		{
			name:     "or true 2",
			exp:      `true || false`,
			expected: true,
		},
		{
			name:     "or false",
			exp:      `false || false`,
			expected: false,
		},
		{
			name:     "and true true",
			exp:      `true && true`,
			expected: true,
		},
		{
			name:     "and false false",
			exp:      `false && false`,
			expected: false,
		},
		{
			name:     "and true false",
			exp:      `true && false`,
			expected: false,
		},
		{
			name:     "and false true",
			exp:      `false && true`,
			expected: false,
		},
		{
			name:     "contains substr",
			exp:      `"team" CONTAINS "i"`,
			expected: false,
		},
		{
			name:     "contains substr2",
			exp:      `"team" CONTAINS "ea"`,
			expected: true,
		},
		{
			name:     "array contains string",
			exp:      `["ea"] CONTAINS "ea"`,
			expected: true,
		},
		{
			name:     "array contains array",
			exp:      `["a",["b","a"]] CONTAINS ["b","a"]`,
			expected: true,
		},
		{
			name:     "startswith false",
			exp:      `"team" STARTSWITH "i"`,
			expected: false,
		},
		{
			name:     "startswith true",
			exp:      `"team" STARTSWITH "te"`,
			expected: true,
		},
		{
			name:     "endswith false",
			exp:      `"team" ENDSWITH "i"`,
			expected: false,
		},
		{
			name:     "endswith true",
			exp:      `"team" ENDSWITH "am"`,
			expected: true,
		},
		{
			name:     "IN true",
			exp:      `"test" IN .field1`,
			src:      `{"field1":["test"]}`,
			expected: true,
		},
		{
			name:     "IN false",
			exp:      `"me" IN .field1`,
			src:      `{"field1":["test"]}`,
			expected: false,
		},
		{
			name:     "IN false empty",
			exp:      `"me" IN .field1`,
			src:      `{"field1":[]}`,
			expected: false,
		},
		{
			name:     "selectorPath IN value true",
			exp:      `.field1 IN ["test"]`,
			src:      `{"field1":"test"}`,
			expected: true,
		},
		{
			name:     "selectorPath IN value true multiple",
			exp:      `.field1 IN ["test","foo","bar",]`,
			src:      `{"field1":"test"}`,
			expected: true,
		},
		{
			name:     "array eq false",
			exp:      `[] == ["test"]`,
			expected: false,
		},
		{
			name:     "ampersand calc",
			exp:      `(1 + 1) / 2`,
			expected: float64(1),
		},
		{
			name:     "not ampersand same calc",
			exp:      `1 + (1 / 2)`,
			expected: 1.5,
		},
		{
			name:     "company employees true",
			exp:      `.properties.employees > 20`,
			src:      `{"name":"Company","properties":{"employees":50}}`,
			expected: true,
		},
		{
			name:     "company employees false",
			exp:      `.properties.employees > 50`,
			src:      `{"name":"Company","properties":{"employees":50}}`,
			expected: false,
		},
		{
			name:     "company not employees true",
			exp:      `.properties.employees !> 50`,
			src:      `{"name":"Company","properties":{"employees":50}}`,
			expected: true,
		},
		{
			name:     "company not employees false",
			exp:      `.properties.employees !> 20`,
			src:      `{"name":"Company","properties":{"employees":50}}`,
			expected: false,
		},
		{
			name:     "company not employees !=",
			exp:      `.properties.employees != 50`,
			src:      `{"name":"Company","properties":{"employees":50}}`,
			expected: false,
		},
		{
			name:     "company not selectorPath",
			exp:      `!.f1`,
			src:      `{"f1":true,"f2":false}`,
			expected: false,
		},
		{
			name:     "company not selectorPath 2",
			exp:      `!.f2`,
			src:      `{"f1":true,"f2":false}`,
			expected: true,
		},
		{
			name:     "company != selectorPath 2",
			exp:      `!(.f1 != .f2) && !.f2`,
			src:      `{"f1":true,"f2":false}`,
			expected: false,
		},
		{
			name:     "company != selectorPath complex",
			exp:      `!(.f1 != .f2) && !.f2`,
			src:      `{"f1":true,"f2":false}`,
			expected: false,
		},
		{
			name:     "company ! paren selectorPath &&",
			exp:      `!(.f1 && .f2)`,
			src:      `{"f1":true,"f2":false}`,
			expected: true,
		},
		{
			name:     "COERCE DateTime",
			exp:      `COERCE .name _datetime_`,
			src:      `{"name":"2022-01-02"}`,
			expected: time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "COERCE DateTime equality",
			exp:      `COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_`,
			src:      `{"dt1":"2022-01-02","dt2":"2022-01-02"}`,
			expected: true,
		},
		{
			name:     "COERCE DateTime equality mixed",
			exp:      `COERCE .dt1 _datetime_ == COERCE "2022-07-14T17:50:08.318426001Z" _datetime_`,
			src:      `{"dt1":"2022-07-14T17:50:08.318426000Z"}`,
			expected: false,
		},
		{
			name:     "COERCE DateTime equality and eq",
			exp:      `COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_ && true == true`,
			src:      `{"dt1":"2022-07-14T17:50:08.318426000Z","dt2":"2022-07-14T17:50:08.318426001Z"}`,
			expected: false,
		},
		{
			name:     "COERCE DateTime equality and eq with parenthesis'",
			exp:      `(COERCE .dt1 _datetime_ == COERCE .dt2 _datetime_) && true == true`,
			src:      `{"dt1":"2022-07-14T17:50:08.318426000Z","dt2":"2022-07-14T17:50:08.318426001Z"}`,
			expected: false,
		},
		{
			name:     "COERCE DateTime equality and eq with parenthesis' 2",
			exp:      `(COERCE .dt1 _datetime_) == (COERCE .dt2 _datetime_) && true == true`,
			src:      `{"dt1":"2022-07-14T17:50:08.318426000Z","dt2":"2022-07-14T17:50:08.318426001Z"}`,
			expected: false,
		},
		{
			name:     "COERCE DateTime equality constants",
			exp:      `COERCE "2022-07-14T17:50:08.318426000Z" _datetime_ == COERCE "2022-07-14T17:50:08.318426000Z" _datetime_`,
			src:      ``,
			expected: true,
		},
		{
			name:     "string CONTAINS_ANY characters",
			exp:      `"team" CONTAINS_ANY "im"`,
			src:      ``,
			expected: true,
		},
		{
			name:     "array CONTAINS_ANY characters",
			exp:      `["a","b","c"] CONTAINS_ANY "eac"`,
			src:      ``,
			expected: true,
		},
		{
			name:     "array CONTAINS_ANY characters false",
			exp:      `["a","b","c"] CONTAINS_ANY "xyz"`,
			src:      ``,
			expected: false,
		},
		{
			name:     "array CONTAINS_ANY array elements true",
			exp:      `["a","b","c"] CONTAINS_ANY ["c","d","e"]`,
			src:      ``,
			expected: true,
		},
		{
			name:     "array CONTAINS_ANY array elements false",
			exp:      `["a","b","c"] CONTAINS_ANY ["d","e","f"]`,
			src:      ``,
			expected: false,
		},
		{
			name:     "array !CONTAINS_ANY array elements",
			exp:      `["a","b","c"] !CONTAINS_ANY ["d","e","f"]`,
			src:      ``,
			expected: true,
		},
		{
			name:     "string CONTAINS_ALL characters",
			exp:      `"team" CONTAINS_ALL "meat"`,
			src:      ``,
			expected: true,
		},
		{
			name:     "array CONTAINS_ALL string characters",
			exp:      `["a","b","c"] CONTAINS_ALL "cab"`,
			src:      ``,
			expected: true,
		},
		{
			name:     "array CONTAINS_ALL string characters false",
			exp:      `["a","b","c"] CONTAINS_ALL "xyz"`,
			src:      ``,
			expected: false,
		},
		{
			name:     "array CONTAINS_ALL array elements",
			exp:      `["a","b","c"] CONTAINS_ALL ["c","a","b"]`,
			src:      ``,
			expected: true,
		},
		{
			name:     "array !CONTAINS_ALL array elements",
			exp:      `["a","b","c"] !CONTAINS_ALL ["a","b"]`,
			src:      ``,
			expected: false,
		},
		{
			name:     "COERCE _datetime_ gt",
			exp:      `COERCE "2022-07-14T17:50:08.318426001Z" _datetime_ > COERCE "2022-07-14T17:50:08.318426000Z" _datetime_`,
			src:      ``,
			expected: true,
		},
		{
			name:     "COERCE _datetime_ lt",
			exp:      `COERCE "2022-07-14T17:50:08.318426000Z" _datetime_ < COERCE "2022-07-14T17:50:08.318426001Z" _datetime_`,
			src:      ``,
			expected: true,
		},
		{
			name:     "COERCE _datetime_ gte gt",
			exp:      `COERCE "2022-07-14T17:50:08.318426001Z" _datetime_ >= COERCE "2022-07-14T17:50:08.318426000Z" _datetime_`,
			src:      ``,
			expected: true,
		},
		{
			name:     "COERCE _datetime_ lte lt",
			exp:      `COERCE "2022-07-14T17:50:08.318426000Z" _datetime_ <= COERCE "2022-07-14T17:50:08.318426001Z" _datetime_`,
			src:      ``,
			expected: true,
		},
		{
			name:     "COERCE _datetime_ gte equal",
			exp:      `COERCE "2022-07-14T17:50:08.318426000Z" _datetime_ >= COERCE "2022-07-14T17:50:08.318426000Z" _datetime_`,
			src:      ``,
			expected: true,
		},
		{
			name:     "COERCE _datetime_ lte equal",
			exp:      `COERCE "2022-07-14T17:50:08.318426000Z" _datetime_ <= COERCE "2022-07-14T17:50:08.318426000Z" _datetime_`,
			src:      ``,
			expected: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ex, err := Parse([]byte(tc.exp))
			if tc.parseErr != nil {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			//fmt.Printf("%#v\n", ex)
			got, err := ex.Calculate([]byte(tc.src))
			if tc.err != nil {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.expected, got)
		})
	}
}
