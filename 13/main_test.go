package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseFieldsSimple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"1", "1", false},
		{"1,3", "1,3", false},
		{"2-4", "2,3,4", false},
		{"1,3-5", "1,3,4,5", false},
		{"abc", "", true},
		{"1-", "", true},
	}

	for _, test := range tests {
		result, err := parseFields(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("parseFields(%s) expected error, got none", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("parseFields(%s) unexpected error: %v", test.input, err)
			continue
		}

		// Проверяем что ожидаемые поля присутствуют
		expectedFields := strings.Split(test.expected, ",")
		for _, fieldStr := range expectedFields {
			if fieldStr == "" {
				continue
			}
			field, _ := strconv.Atoi(fieldStr)
			if !result[field] {
				t.Errorf("parseFields(%s) missing field %d", test.input, field)
			}
		}
	}
}

func TestProcessLineBasic(t *testing.T) {
	tests := []struct {
		line      string
		delimiter string
		fields    string
		expected  string
	}{
		{"a\tb\tc", "\t", "1", "a"},
		{"a,b,c", ",", "2", "b"},
		{"1-2-3", "-", "1,3", "1-3"},
		{"a\tb\tc\td", "\t", "1,3", "a\tc"},
	}

	for _, test := range tests {
		fieldSet, err := parseFields(test.fields)
		if err != nil {
			t.Fatalf("Failed to parse fields %s: %v", test.fields, err)
		}

		got := processLine(test.line, test.delimiter, fieldSet, false)
		if got != test.expected {
			t.Errorf("Line %q, delim %q, fields %q: got %q, want %q",
				test.line, test.delimiter, test.fields, got, test.expected)
		}
	}
}
