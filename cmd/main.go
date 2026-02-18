package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/sandrolain/crawldown/src/converter"
	"github.com/sandrolain/crawldown/src/crawler"
)

// Args defines command-line arguments
type Args struct {
	URL            string   `arg:"-u,--url" help:"The starting URL to crawl (not needed with --single)"`
	OutputDir      string   `arg:"-o,--output,required" help:"The directory where markdown files will be saved"`
	Single         string   `arg:"-s,--single" help:"Download a single page URL instead of crawling (overrides --url)"`
	MaxDepth       int      `arg:"-d,--depth" default:"2" help:"Maximum crawl depth"`
	ExcludedPaths  []string `arg:"-e,--exclude,separate" help:"URL path prefixes to exclude from crawling"`
	RequestTimeout int      `arg:"-t,--timeout" default:"60" help:"Request timeout in seconds"`
	RequestDelay   int      `arg:"--delay" default:"1" help:"Delay between requests in seconds"`
}

// Description returns the program description
func (Args) Description() string {
	return "A web crawler that downloads and converts website content to Markdown format"
}

// Version returns the program version
func (Args) Version() string {
	return "crawldown 1.0.0"
}

func main() {
	var args Args
	arg.MustParse(&args)

	// Validate required arguments
	if args.OutputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: OUTPUTDIR is required\n")
		os.Exit(1)
	}

	if args.Single == "" && args.URL == "" {
		fmt.Fprintf(os.Stderr, "Error: either a URL (positional) or --single flag is required\n")
		os.Exit(1)
	}

	fmt.Printf("Starting crawl of: %s\n", args.URL)
	fmt.Printf("Output directory: %s\n", args.OutputDir)
	fmt.Printf("Max depth: %d\n", args.MaxDepth)
	fmt.Printf("Request timeout: %ds\n", args.RequestTimeout)
	fmt.Printf("Request delay: %ds\n", args.RequestDelay)
	if len(args.ExcludedPaths) > 0 {
		fmt.Printf("Excluded paths: %v\n", args.ExcludedPaths)
	}
	fmt.Println()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(args.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize converter
	converterOpts := converter.Options{
		Domain:           "",
		BulletListMarker: "-",
		CodeBlockStyle:   "fenced",
		EmDelimiter:      "*",
		StrongDelimiter:  "**",
		LinkStyle:        "inlined",
	}

	conv, err := converter.NewConverter(converterOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating converter: %v\n", err)
		os.Exit(1)
	}

	// Track URL to filename mapping for link conversion
	urlToFile := make(map[string]string)
	var urlToFileMutex sync.Mutex
	pageData := make(map[string]struct {
		markdown string
		filename string
		pageURL  string
	})
	var pageDataMutex sync.Mutex
	pageCount := 0
	var pageCountMutex sync.Mutex

	// Determine start URL and single-page mode
	startURL := args.URL
	isSingle := false
	if args.Single != "" {
		startURL = args.Single
		isSingle = true
		fmt.Printf("Single-page mode: fetching %s only\n", startURL)
	}

	// Initialize crawler
	crawlerOpts := crawler.Options{
		MaxDepth:            args.MaxDepth,
		UserAgent:           "CrawlDown/1.0",
		IgnoreRobotsTxt:     false,
		FollowExternalLinks: false,
		SinglePage:          isSingle,
		RequestTimeout:      args.RequestTimeout,
		RequestDelay:        args.RequestDelay,
		ExcludedPaths:       args.ExcludedPaths,
	}

	c, err := crawler.NewCrawler(startURL, crawlerOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating crawler: %v\n", err)
		os.Exit(1)
	}

	// Set callback to process pages as they are crawled
	c.OnPage(func(page crawler.Page) {
		pageCountMutex.Lock()
		pageCount++
		currentCount := pageCount
		pageCountMutex.Unlock()

		fmt.Printf("[%d] Crawling: %s\n", currentCount, page.URL)

		// Convert HTML to Markdown
		markdown, err := conv.Convert(page.Content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error converting page: %v\n", err)
			return
		}

		// Generate filename
		filename := converter.GenerateFilename(page.URL)

		// Normalize URL (remove trailing slash for consistency)
		normalizedURL := strings.TrimSuffix(page.URL, "/")

		urlToFileMutex.Lock()
		urlToFile[normalizedURL] = filename
		urlToFileMutex.Unlock()

		// Add metadata header
		header := fmt.Sprintf("# %s\n\nURL: %s\n\n---\n\n", page.Title, page.URL)
		markdown = header + markdown

		// Store page data for later processing
		pageDataMutex.Lock()
		pageData[normalizedURL] = struct {
			markdown string
			filename string
			pageURL  string
		}{
			markdown: markdown,
			filename: filename,
			pageURL:  page.URL,
		}
		pageDataMutex.Unlock()
	})

	// Start crawling
	if err := c.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during crawling: %v\n", err)
		os.Exit(1)
	}

	pageCountMutex.Lock()
	finalPageCount := pageCount
	pageCountMutex.Unlock()

	fmt.Printf("\nCrawled %d pages. Converting links and saving files...\n\n", finalPageCount)

	// Second pass: convert links and save files
	successCount := 0
	processedCount := 0

	pageDataMutex.Lock()
	pageDataCopy := make(map[string]struct {
		markdown string
		filename string
		pageURL  string
	})
	for k, v := range pageData {
		pageDataCopy[k] = v
	}
	pageDataMutex.Unlock()

	for _, data := range pageDataCopy {
		processedCount++
		fmt.Printf("[%d/%d] Processing: %s\n", processedCount, len(pageDataCopy), data.pageURL)

		// Convert links to local file references
		urlToFileMutex.Lock()
		urlToFileCopy := make(map[string]string)
		for k, v := range urlToFile {
			urlToFileCopy[k] = v
		}
		urlToFileMutex.Unlock()

		markdown := converter.ConvertLinksToLocal(data.markdown, data.pageURL, urlToFileCopy)

		outputPath := filepath.Join(args.OutputDir, data.filename)

		// Save to file
		if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  Error saving file: %v\n", err)
			continue
		}

		fmt.Printf("  Saved: %s\n", outputPath)
		successCount++
	}

	fmt.Printf("\nSuccessfully processed %d pages\n", successCount)
}
