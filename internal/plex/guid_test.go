package plex

import (
	"testing"

	"github.com/bnjns/plex-artwork-uploader/internal/types"
)

func TestCanonicalGuid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "tvdb://350665", expected: "tvdb://350665"},
		{input: " TMDB://1111 ", expected: "tmdb://1111"},
		{input: "com.plexapp.agents.tvdb://350665", expected: "tvdb://350665"},
	}

	for _, tt := range tests {
		if got := canonicalGuid(tt.input); got != tt.expected {
			t.Fatalf("canonicalGuid(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestInferMediaTypeFromGuid(t *testing.T) {
	tests := []struct {
		input       string
		expected    types.MediaType
		expectError bool
	}{
		{input: "tvdb://350665", expected: types.MediaTypeShow},
		{input: "tmdb://12345", expected: types.MediaTypeMovie},
		{input: "com.plexapp.agents.tvdb://350665", expected: types.MediaTypeShow},
		{input: "imdb://tt0111161", expectError: true},
	}

	for _, tt := range tests {
		got, err := inferMediaTypeFromGuid(tt.input)
		if tt.expectError {
			if err == nil {
				t.Fatalf("inferMediaTypeFromGuid(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Fatalf("inferMediaTypeFromGuid(%q) returned error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Fatalf("inferMediaTypeFromGuid(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
