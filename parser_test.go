package ksql

import (
	"fmt"
	"testing"

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
		//{
		//	name:     "ident + ident",
		//	exp:      ".f1 + .f2",
		//	src:      `{"f1":1,"f2":1}`,
		//	expected: float64(2),
		//},
		{
			name:     "first_name + last_name",
			exp:      `.field1 + " " + .field2`,
			src:      `{"field1":"Dean","field2":"Karn"}`,
			expected: "Dean Karn",
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

			fmt.Printf("%#v\n", ex)

			got, err := ex.Calculate([]byte(tc.src))
			if tc.err != nil {
				assert.Error(err)
				return
			}
			assert.Equal(tc.expected, got)
		})
	}
}
