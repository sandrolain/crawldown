package crawler

import (
	"testing"
)

func TestLooksLikeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid email",
			input: "test@example.com",
			want:  true,
		},
		{
			name:  "valid email with subdomain",
			input: "info@mail.example.com",
			want:  true,
		},
		{
			name:  "invalid - no @",
			input: "notanemail",
			want:  false,
		},
		{
			name:  "invalid - no domain",
			input: "test@",
			want:  false,
		},
		{
			name:  "invalid - no local part",
			input: "@example.com",
			want:  false,
		},
		{
			name:  "invalid - no dot in domain",
			input: "test@example",
			want:  false,
		},
		{
			name:  "valid - complex email",
			input: "user+tag@subdomain.example.co.uk",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikeEmail(tt.input); got != tt.want {
				t.Errorf("looksLikeEmail(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLooksLikePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid international format",
			input: "+39 0429 783026",
			want:  true,
		},
		{
			name:  "valid with parentheses",
			input: "+1 (555) 123-4567",
			want:  true,
		},
		{
			name:  "valid with dashes",
			input: "555-123-4567",
			want:  true,
		},
		{
			name:  "valid minimal",
			input: "+123456789",
			want:  true,
		},
		{
			name:  "invalid - too few digits",
			input: "+12345",
			want:  false,
		},
		{
			name:  "invalid - too many digits",
			input: "+1234567890123456",
			want:  false,
		},
		{
			name:  "invalid - no phone chars",
			input: "1234567890",
			want:  false,
		},
		{
			name:  "invalid - just text",
			input: "not a phone",
			want:  false,
		},
		{
			name:  "valid with spaces",
			input: "+39 333 1234567",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikePhone(tt.input); got != tt.want {
				t.Errorf("looksLikePhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "URL without query params",
			input: "https://example.com/page",
			want:  "https://example.com/page",
		},
		{
			name:  "URL with single query param",
			input: "https://example.com/page?foo=bar",
			want:  "https://example.com/page?foo=bar",
		},
		{
			name:  "URL with multiple query params in order",
			input: "https://example.com/page?a=1&b=2&c=3",
			want:  "https://example.com/page?a=1&b=2&c=3",
		},
		{
			name:  "URL with query params out of order",
			input: "https://example.com/page?c=3&a=1&b=2",
			want:  "https://example.com/page?a=1&b=2&c=3",
		},
		{
			name:  "URL with duplicate keys should be normalized",
			input: "https://example.com/page?foo=bar&foo=baz",
			want:  "https://example.com/page?foo=bar&foo=baz",
		},
		{
			name:  "invalid URL returns as-is",
			input: "://invalid",
			want:  "://invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeURL(tt.input); got != tt.want {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
