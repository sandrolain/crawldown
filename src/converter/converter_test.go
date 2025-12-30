package converter

import (
	"strings"
	"testing"
)

func TestNewConverter(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{
			name: "default options",
			opts: Options{},
		},
		{
			name: "custom options",
			opts: Options{
				Domain:           "example.com",
				BulletListMarker: "*",
				CodeBlockStyle:   "fenced",
				EmDelimiter:      "_",
				StrongDelimiter:  "__",
				LinkStyle:        "referenced",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewConverter(tt.opts)

			if err != nil {
				t.Errorf("NewConverter() unexpected error: %v", err)
				return
			}

			if c == nil {
				t.Errorf("NewConverter() returned nil")
				return
			}

			if c.converter == nil {
				t.Errorf("NewConverter() converter is nil")
			}
		})
	}
}

func TestConvert(t *testing.T) {
	conv, err := NewConverter(Options{})
	if err != nil {
		t.Fatalf("NewConverter() failed: %v", err)
	}

	tests := []struct {
		name      string
		html      string
		wantError bool
		contains  string
	}{
		{
			name:      "empty HTML",
			html:      "",
			wantError: true,
		},
		{
			name:      "simple paragraph",
			html:      "<p>Hello World</p>",
			wantError: false,
			contains:  "Hello World",
		},
		{
			name:      "heading and paragraph",
			html:      "<h1>Title</h1><p>Content</p>",
			wantError: false,
			contains:  "# Title",
		},
		{
			name:      "link",
			html:      "<a href='https://example.com'>Link</a>",
			wantError: false,
			contains:  "[Link]",
		},
		{
			name:      "bold text",
			html:      "<strong>Bold</strong>",
			wantError: false,
			contains:  "**Bold**",
		},
		{
			name:      "italic text",
			html:      "<em>Italic</em>",
			wantError: false,
			contains:  "Italic",
		},
		{
			name:      "code block",
			html:      "<pre><code>code example</code></pre>",
			wantError: false,
			contains:  "code example",
		},
		{
			name:      "unordered list",
			html:      "<ul><li>Item 1</li><li>Item 2</li></ul>",
			wantError: false,
			contains:  "Item 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.html)

			if tt.wantError {
				if err == nil {
					t.Errorf("Convert() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Convert() unexpected error: %v", err)
				return
			}

			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("Convert() result does not contain '%s'\nGot: %s", tt.contains, result)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "root path",
			url:      "https://example.com/",
			expected: "index.md",
		},
		{
			name:     "simple path",
			url:      "https://example.com/about",
			expected: "about.md",
		},
		{
			name:     "nested path",
			url:      "https://example.com/docs/guide",
			expected: "docs-guide.md",
		},
		{
			name:     "path with extension",
			url:      "https://example.com/page.html",
			expected: "page.md",
		},
		{
			name:     "path with query",
			url:      "https://example.com/search?q=test",
			expected: "search-q-test.md",
		},
		{
			name:     "path with multiple query params",
			url:      "https://example.com/page?id=1&lang=en",
			expected: "page-id-1-lang-en.md",
		},
		{
			name:     "root with query",
			url:      "https://example.com/?ref=home",
			expected: "index-ref-home.md",
		},
		{
			name:     "invalid URL",
			url:      "://invalid",
			expected: "index.md",
		},
		{
			name:     "empty path",
			url:      "https://example.com",
			expected: "index.md",
		},
		{
			name:     "path with special chars",
			url:      "https://example.com/hello:world",
			expected: "hello-world.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilename(tt.url)
			if result != tt.expected {
				t.Errorf("GenerateFilename() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid filename",
			input:    "hello-world",
			expected: "hello-world",
		},
		{
			name:     "with invalid chars",
			input:    "hello<>world",
			expected: "hello-world",
		},
		{
			name:     "multiple dashes",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "leading/trailing dashes",
			input:    "-hello-world-",
			expected: "hello-world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "page",
		},
		{
			name:     "all invalid chars",
			input:    "<>:\"/\\|?*",
			expected: "page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestCleanMarkdown(t *testing.T) {
	conv, err := NewConverter(Options{})
	if err != nil {
		t.Fatalf("NewConverter() failed: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple newlines",
			input:    "line1\n\n\n\nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "leading/trailing whitespace",
			input:    "  \n  content  \n  ",
			expected: "content",
		},
		{
			name:     "normal formatting",
			input:    "line1\n\nline2",
			expected: "line1\n\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.cleanMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("cleanMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}
