package worker

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
			`should add string clientID param to flags`,
			`clientID`,
			`*string`,
		},
		{
			`should add string clientSecret param to flags`,
			`clientSecret`,
			`*string`,
		},
		{
			`should add string accessToken param to flags`,
			`accessToken`,
			`*string`,
		},
		{
			`should add string refreshToken param to flags`,
			`refreshToken`,
			`*string`,
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
