package hue

import (
	"fmt"
	"testing"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
		{
			`should add string bridgeIP param to flags`,
			`bridgeIP`,
			`*string`,
		},
		{
			`should add string username param to flags`,
			`username`,
			`*string`,
		},
		{
			`should add string config param to flags`,
			`config`,
			`*string`,
		},
		{
			`should add bool clean param to flags`,
			`clean`,
			`*bool`,
		},
	}

	for _, testCase := range cases {
		result := Flags(testCase.intention)[testCase.want]

		if result == nil {
			t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
		}

		if fmt.Sprintf(`%T`, result) != testCase.wantType {
			t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
		}
	}
}
