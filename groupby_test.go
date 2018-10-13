package main

import (
	"testing"
)

func TestMonthByName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", ""},
		{"1", "January"},
		{"2", "February"},
		{"3", "March"},
		{"4", "April"},
		{"5", "May"},
		{"6", "June"},
		{"7", "July"},
		{"8", "August"},
		{"9", "September"},
		{"10", "October"},
		{"11", "November"},
		{"12", "December"},
		{"13", ""},
		{"Non-Number", "Non-Number"},
	}

	for _, test := range tests {
		result := MonthAsName(test.input)
		if result != test.expected {
			t.Errorf("MonthAsName(\"%s\") expects \"%s\", got \"%s\"", test.input, test.expected, result)
		}
	}
}
