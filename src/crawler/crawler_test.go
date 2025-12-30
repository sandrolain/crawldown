package crawler

import (
	"testing"
)

func TestNewCrawler(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		opts      Options
		wantError bool
	}{
		{
			name:      "valid URL with default options",
			url:       "https://example.com",
			opts:      Options{},
			wantError: false,
		},
		{
			name: "valid URL with custom options",
			url:  "https://example.com",
			opts: Options{
				MaxDepth:  3,
				UserAgent: "TestBot/1.0",
			},
			wantError: false,
		},
		{
			name:      "invalid URL",
			url:       "://invalid-url",
			opts:      Options{},
			wantError: true,
		},
		{
			name:      "empty URL",
			url:       "",
			opts:      Options{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCrawler(tt.url, tt.opts)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewCrawler() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewCrawler() unexpected error: %v", err)
				return
			}

			if c == nil {
				t.Errorf("NewCrawler() returned nil crawler")
				return
			}

			if c.collector == nil {
				t.Errorf("NewCrawler() collector is nil")
			}

			if c.baseURL == nil {
				t.Errorf("NewCrawler() baseURL is nil")
			}

			// Check default values
			if tt.opts.MaxDepth == 0 && c.options.MaxDepth != 2 {
				t.Errorf("NewCrawler() MaxDepth = %d, want 2", c.options.MaxDepth)
			}

			if tt.opts.UserAgent == "" && c.options.UserAgent != "CrawlDown/1.0" {
				t.Errorf("NewCrawler() UserAgent = %s, want CrawlDown/1.0", c.options.UserAgent)
			}
		})
	}
}

func TestCrawlerGetPages(t *testing.T) {
	opts := Options{
		MaxDepth: 1,
	}

	c, err := NewCrawler("https://example.com", opts)
	if err != nil {
		t.Fatalf("NewCrawler() unexpected error: %v", err)
	}

	pages := c.GetPages()
	if pages == nil {
		t.Errorf("GetPages() returned nil")
	}

	if len(pages) != 0 {
		t.Errorf("GetPages() expected empty slice, got %d pages", len(pages))
	}
}
