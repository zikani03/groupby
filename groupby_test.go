package main

import (
	"testing"
)

func TestFileNameByDepth(t *testing.T) {
	tests := []struct {
		filename string
		depth    int
		expected string
	}{
		{"03", 1, "03"},
		{"03", 2, "March"},
		{"Non-Number", 2, "Non-Number"},
	}

	for _, test := range tests {
		result := FileNameByDepth(test.filename, test.depth)
		if result != test.expected {
			t.Errorf("FileNameByDepth(\"%s\", %d) expects \"%s\", got \"%s\"", test.filename, test.depth, test.expected, result)
		}
	}
}

func TestMonthByName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", ""},
		{"1", "January"},
		{"01", "January"},
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
