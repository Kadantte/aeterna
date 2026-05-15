package common

import (
	"fmt"
	"testing"
)

func TestGetenvTrim(t *testing.T) {
	t.Run("unset key returns empty", func(t *testing.T) {
		if got := GetenvTrim("_TEST_GETENVTRIM_NEVER_SET_XYZ"); got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	tests := []struct {
		value string
		want  string
	}{
		{"hello", "hello"},
		{"  hello", "hello"},
		{"hello  ", "hello"},
		{"  hello  ", "hello"},
		{"\thello\t", "hello"},
		{"hello\n", "hello"},
		{"  hello world  ", "hello world"},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("trim_case_%d", i), func(t *testing.T) {
			key := fmt.Sprintf("_TEST_GETENVTRIM_%d", i)
			t.Setenv(key, tc.value)
			got := GetenvTrim(key)
			if got != tc.want {
				t.Fatalf("input %q: got %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestWithDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{"empty uses fallback", "", "default", "default"},
		{"non-empty uses value", "custom", "default", "custom"},
		{"spaces not trimmed", "  x  ", "default", "  x  "},
		{"fallback not used when value set", "val", "fb", "val"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := WithDefault(tc.value, tc.fallback)
			if got != tc.want {
				t.Fatalf("WithDefault(%q, %q) = %q, want %q", tc.value, tc.fallback, got, tc.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		set      bool
		fallback int
		want     int
	}{
		{"unset uses fallback", "", false, 42, 42},
		{"valid positive", "100", true, 0, 100},
		{"valid zero", "0", true, 99, 0},
		{"valid negative", "-5", true, 0, -5},
		{"invalid string uses fallback", "abc", true, 7, 7},
		{"float string uses fallback", "3.14", true, 8, 8},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			key := fmt.Sprintf("_TEST_GETINT_%d", i)
			if tc.set {
				t.Setenv(key, tc.value)
			}
			got := GetInt(key, tc.fallback)
			if got != tc.want {
				t.Fatalf("GetInt(value=%q, fallback=%d) = %d, want %d", tc.value, tc.fallback, got, tc.want)
			}
		})
	}
}

func TestGetPositiveInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		set      bool
		fallback int
		want     int
	}{
		{"unset uses fallback", "", false, 5, 5},
		{"positive value", "10", true, 5, 10},
		{"zero uses fallback", "0", true, 5, 5},
		{"negative uses fallback", "-1", true, 5, 5},
		{"invalid string uses fallback", "bad", true, 5, 5},
		{"one is valid", "1", true, 5, 1},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			key := fmt.Sprintf("_TEST_GETPOSINT_%d", i)
			if tc.set {
				t.Setenv(key, tc.value)
			}
			got := GetPositiveInt(key, tc.fallback)
			if got != tc.want {
				t.Fatalf("GetPositiveInt(value=%q, fallback=%d) = %d, want %d", tc.value, tc.fallback, got, tc.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		set      bool
		fallback bool
		want     bool
	}{
		{"unset uses false fallback", "", false, false, false},
		{"unset uses true fallback", "", false, true, true},
		{"true", "true", true, false, true},
		{"false", "false", true, true, false},
		{"1", "1", true, false, true},
		{"0", "0", true, true, false},
		{"TRUE uppercase", "TRUE", true, false, true},
		{"FALSE uppercase", "FALSE", true, true, false},
		{"invalid uses fallback", "yes", true, false, false},
		{"empty value uses fallback", "", true, true, true},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			key := fmt.Sprintf("_TEST_GETBOOL_%d", i)
			if tc.set {
				t.Setenv(key, tc.value)
			}
			got := GetBool(key, tc.fallback)
			if got != tc.want {
				t.Fatalf("GetBool(value=%q, fallback=%v) = %v, want %v", tc.value, tc.fallback, got, tc.want)
			}
		})
	}
}
