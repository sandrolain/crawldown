package crawler

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

// Page represents a crawled web page
type Page struct {
	URL     string
	Title   string
	Content string
}

// Options defines crawler configuration
type Options struct {
	MaxDepth            int
	AllowedDomains      []string
	UserAgent           string
	IgnoreRobotsTxt     bool
	FollowExternalLinks bool
	SinglePage          bool     // When true, only the provided start URL is fetched (no link following)
	RequestTimeout      int      // Timeout in seconds for each request (default: 30)
	RequestDelay        int      // Delay in seconds between requests (default: 0)
	ExcludedPaths       []string // URL path prefixes to exclude from crawling
}

// PageCallback is called when a page is successfully crawled
type PageCallback func(page Page)

// Crawler handles web crawling operations
type Crawler struct {
	collector    *colly.Collector
	pages        []Page
	pagesMutex   sync.Mutex
	baseURL      *url.URL
	options      Options
	pageCallback PageCallback
}

// NewCrawler creates a new crawler instance
func NewCrawler(startURL string, opts Options) (*Crawler, error) {
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if opts.MaxDepth == 0 {
		opts.MaxDepth = 2
	}

	if opts.UserAgent == "" {
		opts.UserAgent = "CrawlDown/1.0"
	}

	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 30
	}

	allowedDomains := opts.AllowedDomains
	if len(allowedDomains) == 0 && !opts.FollowExternalLinks {
		allowedDomains = []string{parsedURL.Host}
	}

	c := colly.NewCollector(
		colly.MaxDepth(opts.MaxDepth),
		colly.AllowedDomains(allowedDomains...),
		colly.UserAgent(opts.UserAgent),
		colly.Async(true), // Enable async to handle multiple requests
	)

	// Set timeout
	c.SetRequestTimeout(time.Duration(opts.RequestTimeout) * time.Second)

	// Allow unlimited parallel requests
	//nolint:errcheck // Intentionally using default parallelism
	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
	})

	// Set delay between requests if specified
	if opts.RequestDelay > 0 {
		err := c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Delay:       time.Duration(opts.RequestDelay) * time.Second,
			RandomDelay: time.Duration(opts.RequestDelay/2) * time.Second,
			Parallelism: 2,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set rate limit: %w", err)
		}
	}

	if opts.IgnoreRobotsTxt {
		c.IgnoreRobotsTxt = true
	}

	crawler := &Crawler{
		collector: c,
		pages:     []Page{},
		baseURL:   parsedURL,
		options:   opts,
	}

	return crawler, nil
}

// OnPage sets a callback to be called when each page is crawled
func (c *Crawler) OnPage(callback PageCallback) {
	c.pageCallback = callback
}

// Start begins the crawling process
func (c *Crawler) Start() error {
	c.setupCallbacks()

	err := c.collector.Visit(c.baseURL.String())
	if err != nil {
		return fmt.Errorf("failed to start crawling: %w", err)
	}

	// Wait for all async requests to complete
	c.collector.Wait()

	return nil
}

// setupCallbacks configures the collector callbacks
func (c *Crawler) setupCallbacks() {
	// On HTML element callback
	c.collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Normalize URL to handle query parameters consistently
		normalizedURL := normalizeURL(e.Request.URL.String())

		page := Page{
			URL:     normalizedURL,
			Title:   e.ChildText("title"),
			Content: extractMainContent(e),
		}

		// Thread-safe append for async crawling
		c.pagesMutex.Lock()
		c.pages = append(c.pages, page)
		c.pagesMutex.Unlock()

		// Call callback if set
		if c.pageCallback != nil {
			c.pageCallback(page)
		}
	})

	// On link callback: only register if not in SinglePage mode
	if !c.options.SinglePage {
		c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")

			// Skip non-HTTP protocols and anchor links
			if strings.HasPrefix(link, "#") ||
				strings.HasPrefix(link, "javascript:") ||
				strings.HasPrefix(link, "mailto:") ||
				strings.HasPrefix(link, "tel:") ||
				strings.HasPrefix(link, "sms:") ||
				strings.HasPrefix(link, "fax:") ||
				strings.HasPrefix(link, "data:") ||
				strings.HasPrefix(link, "file:") {
				return
			}

			// Skip links that look like email addresses or phone numbers without protocol
			if looksLikeEmail(link) || looksLikePhone(link) {
				return
			}

			// Build absolute URL for checking
			absoluteURL := e.Request.AbsoluteURL(link)

			// Skip excluded paths
			if c.isExcludedPath(absoluteURL) {
				return
			}

			// Visit is best effort, errors are logged via OnError callback
			//nolint:errcheck // Intentionally ignoring error as it's handled by OnError callback
			_ = e.Request.Visit(link)
		})
	}

	// Error callback
	c.collector.OnError(func(r *colly.Response, err error) {
		// nolint:forbidigo // Logging output during crawling
		fmt.Printf("Error crawling %s: %v\n", r.Request.URL, err)
	})

	// Request callback
	c.collector.OnRequest(func(r *colly.Request) {
		// nolint:forbidigo // Logging output during crawling
		fmt.Printf("Visiting: %s\n", r.URL.String())
	})
}

// extractMainContent attempts to extract the main content from the page
func extractMainContent(e *colly.HTMLElement) string {
	var content string

	// Try to find main content areas in order of priority
	selectors := []string{
		"main",
		"article",
		"[role='main']",
		".content",
		"#content",
		".main-content",
		"#main-content",
		"body",
	}

	for _, selector := range selectors {
		if html, err := e.DOM.Find(selector).First().Html(); err == nil && html != "" {
			content = html
			break
		}
	}

	return content
}

// GetPages returns all crawled pages
func (c *Crawler) GetPages() []Page {
	c.pagesMutex.Lock()
	defer c.pagesMutex.Unlock()
	return c.pages
}

// normalizeURL normalizes URL by sorting query parameters alphabetically
func normalizeURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Sort query parameters alphabetically
	query := parsedURL.Query()
	parsedURL.RawQuery = query.Encode() // Encode() automatically sorts keys

	return parsedURL.String()
}

// isExcludedPath checks if a URL path should be excluded
func (c *Crawler) isExcludedPath(rawURL string) bool {
	if len(c.options.ExcludedPaths) == 0 {
		return false
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	fullPath := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

	for _, excluded := range c.options.ExcludedPaths {
		if strings.HasPrefix(fullPath, excluded) || strings.HasPrefix(rawURL, excluded) {
			return true
		}
	}

	return false
}

// looksLikeEmail checks if a string looks like an email address
func looksLikeEmail(s string) bool {
	// Simple check: contains @ and has text before and after it
	if !strings.Contains(s, "@") {
		return false
	}
	parts := strings.Split(s, "@")
	return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 && strings.Contains(parts[1], ".")
}

// looksLikePhone checks if a string looks like a phone number
func looksLikePhone(s string) bool {
	// Remove common phone number characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)

	// Check if we have a reasonable number of digits (7-15 is typical for phone numbers)
	// and the original string starts with + or contains common phone chars
	digitCount := len(cleaned)
	hasPhoneChars := strings.ContainsAny(s, "+()-")

	return digitCount >= 7 && digitCount <= 15 && hasPhoneChars
}
