package main

import (
	"fmt"
	"testing"
)

func TestStripParts(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"This is a test (Part A)", "This is a test "},
		{"This is a test (Subpart B)", "This is a test "},
		{"This is a test (Part A) and (Subpart B)", "This is a test  and "},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := StripParts(tc.s)
			if got != tc.want {
				t.Errorf("StripParts() = %v, want %v for %#v", got, tc.want, tc.s)
			}
		})
	}
}
