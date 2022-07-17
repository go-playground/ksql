package ksql

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/require"
//)
//
//func TestParser(t *testing.T) {
//	assert := require.New(t)
//
//	tests := []struct {
//		name     string
//		exp      string
//		src      string
//		expected any
//		err      error
//		parseErr error
//	}{
//		{
//			name:     "ident + ident",
//			exp:      ".f1 + .f2",
//			src:      `{"f1":1,"f2":1}`,
//			expected: float64(2),
//		},
//		{
//			name:     "first_name + last_name",
//			exp:      `.field1 + " " + .field2`,
//			src:      `{"field1":"Dean","field2":"Karn"}`,
//			expected: "Dean Karn",
//		},
//		{
//			name:     "ident + ident num",
//			exp:      `.field1 + .field2`,
//			src:      `{"field1":10.1,"field2":23.23}`,
//			expected: 33.33,
//		},
//		{
//			name:     "ident - ident num",
//			exp:      `.field2 - .field1`,
//			src:      `{"field1":10.1,"field2":23.23}`,
//			expected: 13.13,
//		},
//		{
//			name:     "ident * ident num",
//			exp:      `.field2 * .field1`,
//			src:      `{"field1":11.1,"field2":3}`,
//			expected: 33.3,
//		},
//		{
//			name:     "ident / ident num",
//			exp:      `.field2 / .field1`,
//			src:      `{"field1":3,"field2":33.3}`,
//			expected: 11.1,
//		},
//		{
//			name:     "num + num",
//			exp:      `11.1 + 22.2`,
//			expected: 33.3,
//		},
//		{
//			name:     "ident + num",
//			exp:      `11.1 + .field1`,
//			src:      `{"field1":3,"field2":33.3}`,
//			expected: 14.1,
//		},
//		{
//			name:     "ident == num false",
//			exp:      `11.1 == .field1`,
//			src:      `{"field1":3,"field2":33.3}`,
//			expected: false,
//		},
//		{
//			name:     "ident == num true",
//			exp:      `11.1 == .field1`,
//			src:      `{"field1":11.1,"field2":33.3}`,
//			expected: true,
//		},
//		{
//			name:     "ident > num false",
//			exp:      `11.1 > .field1`,
//			src:      `{"field1":11.1,"field2":33.3}`,
//			expected: false,
//		},
//		{
//			name:     "ident > num true",
//			exp:      `11.1 > .field1`,
//			src:      `{"field1":11.0,"field2":33.3}`,
//			expected: true,
//		},
//		{
//			name:     "ident >= num false",
//			exp:      `11.1 >= .field1`,
//			src:      `{"field1":12.0,"field2":33.3}`,
//			expected: false,
//		},
//		{
//			name:     "ident >= num true",
//			exp:      `11.1 >= .field1`,
//			src:      `{"field1":11.0,"field2":33.3}`,
//			expected: true,
//		},
//		{
//			name:     "bool true",
//			exp:      `true == true`,
//			expected: true,
//		},
//		{
//			name:     "bool false",
//			exp:      `false == true`,
//			expected: false,
//		},
//		{
//			name:     "null eq",
//			exp:      `NULL == NULL`,
//			expected: true,
//		},
//		{
//			name:     "or true 1",
//			exp:      `false || true`,
//			expected: true,
//		},
//		{
//			name:     "or true 2",
//			exp:      `true || false`,
//			expected: true,
//		},
//		{
//			name:     "or false",
//			exp:      `false || false`,
//			expected: false,
//		},
//		{
//			name:     "and true true",
//			exp:      `true && true`,
//			expected: true,
//		},
//		{
//			name:     "and false false",
//			exp:      `false && false`,
//			expected: false,
//		},
//		{
//			name:     "and true false",
//			exp:      `true && false`,
//			expected: false,
//		},
//		{
//			name:     "and false true",
//			exp:      `false && true`,
//			expected: false,
//		},
//		{
//			name:     "contains false",
//			exp:      `"team" CONTAINS "i"`,
//			expected: false,
//		},
//		{
//			name:     "contains true",
//			exp:      `"team" CONTAINS "ea"`,
//			expected: true,
//		},
//		{
//			name:     "startswith false",
//			exp:      `"team" STARTSWITH "i"`,
//			expected: false,
//		},
//		{
//			name:     "startswith true",
//			exp:      `"team" STARTSWITH "te"`,
//			expected: true,
//		},
//		{
//			name:     "endswith false",
//			exp:      `"team" ENDSWITH "i"`,
//			expected: false,
//		},
//		{
//			name:     "endswith true",
//			exp:      `"team" ENDSWITH "am"`,
//			expected: true,
//		},
//		{
//			name:     "IN true",
//			exp:      `"test" IN .field1`,
//			src:      `{"field1":["test"]}`,
//			expected: true,
//		},
//		{
//			name:     "IN false",
//			exp:      `"me" IN .field1`,
//			src:      `{"field1":["test"]}`,
//			expected: false,
//		},
//		{
//			name:     "IN false empty",
//			exp:      `"me" IN .field1`,
//			src:      `{"field1":[]}`,
//			expected: false,
//		},
//		{
//			name: 		"ident IN value true",
//			exp:		`.field1 IN ["test"]`,
//			src: 		`{"field1":"test"}`,
//			expected:	true,
//		},
//		{
//			name: 		"ident IN value true multiple",
//			exp:		`.field1 IN ["test","foo","bar",]`,
//			src: 		`{"field1":"test"}`,
//			expected:	true,
//		},
//		{
//			name:     "array eq false",
//			exp:      `[] == ["test"]`,
//			expected: false,
//		},
//		{
//			name:     "ampersand calc",
//			exp:      `(1 + 1) / 2`,
//			expected: float64(1),
//		},
//		{
//			name:     "not ampersand same calc",
//			exp:      `1 + 1 / 2`,
//			expected: 1.5,
//		},
//		{
//			name:     "company employees true",
//			exp:      `.properties.employees > 20`,
//			src:      `{"name":"Company","properties":{"employees":50}}`,
//			expected: true,
//		},
//		{
//			name:     "company employees false",
//			exp:      `.properties.employees > 50`,
//			src:      `{"name":"Company","properties":{"employees":50}}`,
//			expected: false,
//		},
//		{
//			name:     "company not employees true",
//			exp:      `.properties.employees !> 50`,
//			src:      `{"name":"Company","properties":{"employees":50}}`,
//			expected: true,
//		},
//		{
//			name:     "company not employees false",
//			exp:      `.properties.employees !> 20`,
//			src:      `{"name":"Company","properties":{"employees":50}}`,
//			expected: false,
//		},
//		{
//			name:     "company not employees !=",
//			exp:      `.properties.employees != 50`,
//			src:      `{"name":"Company","properties":{"employees":50}}`,
//			expected: false,
//		},
//		{
//			name:     "company not ident",
//			exp:      `!.f1`,
//			src:      `{"f1":true,"f2":false}`,
//			expected: false,
//		},
//		{
//			name:     "company not ident 2",
//			exp:      `!.f2`,
//			src:      `{"f1":true,"f2":false}`,
//			expected: true,
//		},
//		{
//			name:     "company != ident 2",
//			exp:      `!(.f1 != .f2) && !.f2`,
//			src:      `{"f1":true,"f2":false}`,
//			expected: false,
//		},
//		{
//			name:     "company != ident complex",
//			exp:      `!(.f1 != .f2) && !.f2`,
//			src:      `{"f1":true,"f2":false}`,
//			expected: false,
//		},
//		{
//			name:     "company ! paran ident &&",
//			exp:      `!(.f1 && .f2)`,
//			src:      `{"f1":true,"f2":false}`,
//			expected: true,
//		},
//	}
//
//	for _, tc := range tests {
//		tc := tc
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//
//			ex, err := Parse([]byte(tc.exp))
//			if tc.parseErr != nil {
//				assert.Error(err)
//				return
//			}
//			assert.NoError(err)
//
//			got, err := ex.Calculate([]byte(tc.src))
//			if tc.err != nil {
//				assert.Error(err)
//				return
//			}
//			assert.NoError(err)
//			assert.Equal(tc.expected, got)
//		})
//	}
//}
