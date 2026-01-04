package crawler

import (
	"net/http"
	"net/http/httptest"
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

func TestCrawlerSinglePageMode(t *testing.T) {
	// Create a test server with two pages: /index links to /next
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Index</title></head><body><a href="/next">Next</a><main><p>Index content</p></main></body></html>`))
	})
	mux.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Next</title></head><body><main><p>Next content</p></main></body></html>`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// SinglePage mode: should only fetch the start URL
	opts := Options{
		SinglePage: true,
	}

	c, err := NewCrawler(srv.URL, opts)
	if err != nil {
		t.Fatalf("NewCrawler() unexpected error: %v", err)
	}

	if err := c.Start(); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}

	pages := c.GetPages()
	if len(pages) != 1 {
		t.Fatalf("SinglePage mode expected 1 page, got %d", len(pages))
	}
	if pages[0].Title != "Index" {
		t.Fatalf("Unexpected page fetched: %s", pages[0].Title)
	}

	// Non-single mode: should fetch both pages when following links
	opts2 := Options{}
	c2, err := NewCrawler(srv.URL, opts2)
	if err != nil {
		t.Fatalf("NewCrawler() unexpected error: %v", err)
	}

	if err := c2.Start(); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}

	pages2 := c2.GetPages()
	if len(pages2) < 2 {
		t.Fatalf("Normal mode expected at least 2 pages, got %d", len(pages2))
	}
}
