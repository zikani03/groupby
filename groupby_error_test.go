package main

import "testing"

func TestGroupbyError(t *testing.T) {
	tests := []struct {
		input string
	}{
		{input: "Error Message"},
		{input: "Different Error"},
	}

	for _, test := range tests {

		err := groupbyError(test.input)
		if msg := err.Error(); msg != test.input {
			t.Errorf("GroupbyError message does not match expected. Got '%s', Expected '%s'", msg, test.input)
		}
	}
}
