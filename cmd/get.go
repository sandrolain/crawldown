package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sandrolain/crawldown/src/converter"
	"github.com/sandrolain/crawldown/src/crawler"
)

type getOptions struct {
	outputDir           string
	singleURL           string
	maxDepth            int
	excludedPaths       []string
	requestTimeout      int
	requestDelay        int
	ignoreRobotsTxt     bool
	followExternalLinks bool
	userAgent           string
}

func defaultGetOptions() *getOptions {
	return &getOptions{
		maxDepth:       2,
		requestTimeout: 60,
		requestDelay:   1,
		userAgent:      "CrawlDown/1.0",
	}
}

func runGet(options *getOptions, args []string) error {
	startURL := ""
	if len(args) > 0 {
		startURL = args[0]
	}

	isSingle := false
	if options.singleURL != "" {
		startURL = options.singleURL
		isSingle = true
	}

	printStdout("Starting crawl of: %s\n", startURL)
	printStdout("Output directory: %s\n", options.outputDir)
	printStdout("Max depth: %d\n", options.maxDepth)
	printStdout("Request timeout: %ds\n", options.requestTimeout)
	printStdout("Request delay: %ds\n", options.requestDelay)
	printStdout("Ignore robots.txt: %t\n", options.ignoreRobotsTxt)
	printStdout("Follow external links: %t\n", options.followExternalLinks)
	if len(options.excludedPaths) > 0 {
		printStdout("Excluded paths: %v\n", options.excludedPaths)
	}
	if isSingle {
		printStdout("Single-page mode: fetching %s only\n", startURL)
	}
	printlnStdout()

	if err := os.MkdirAll(options.outputDir, 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

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
		return fmt.Errorf("create converter: %w", err)
	}

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

	crawlerOpts := crawler.Options{
		MaxDepth:            options.maxDepth,
		UserAgent:           options.userAgent,
		IgnoreRobotsTxt:     options.ignoreRobotsTxt,
		FollowExternalLinks: options.followExternalLinks,
		SinglePage:          isSingle,
		RequestTimeout:      options.requestTimeout,
		RequestDelay:        options.requestDelay,
		ExcludedPaths:       options.excludedPaths,
	}

	c, err := crawler.NewCrawler(startURL, crawlerOpts)
	if err != nil {
		return fmt.Errorf("create crawler: %w", err)
	}

	c.OnPage(func(page crawler.Page) {
		pageCountMutex.Lock()
		pageCount++
		currentCount := pageCount
		pageCountMutex.Unlock()

		printStdout("[%d] Crawling: %s\n", currentCount, page.URL)

		markdown, err := conv.Convert(page.Content)
		if err != nil {
			printStderr("  Error converting page: %v\n", err)
			return
		}

		filename := converter.GenerateFilename(page.URL)
		normalizedURL := strings.TrimSuffix(page.URL, "/")

		urlToFileMutex.Lock()
		urlToFile[normalizedURL] = filename
		urlToFileMutex.Unlock()

		header := fmt.Sprintf("# %s\n\nURL: %s\n\n---\n\n", page.Title, page.URL)
		markdown = header + markdown

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

	if err := c.Start(); err != nil {
		return fmt.Errorf("crawl: %w", err)
	}

	pageCountMutex.Lock()
	finalPageCount := pageCount
	pageCountMutex.Unlock()

	printStdout("\nCrawled %d pages. Converting links and saving files...\n\n", finalPageCount)

	successCount := 0
	processedCount := 0

	pageDataMutex.Lock()
	pageDataCopy := make(map[string]struct {
		markdown string
		filename string
		pageURL  string
	})
	for key, value := range pageData {
		pageDataCopy[key] = value
	}
	pageDataMutex.Unlock()

	for _, data := range pageDataCopy {
		processedCount++
		printStdout("[%d/%d] Processing: %s\n", processedCount, len(pageDataCopy), data.pageURL)

		urlToFileMutex.Lock()
		urlToFileCopy := make(map[string]string)
		for key, value := range urlToFile {
			urlToFileCopy[key] = value
		}
		urlToFileMutex.Unlock()

		markdown := converter.ConvertLinksToLocal(data.markdown, data.pageURL, urlToFileCopy)
		outputPath := filepath.Join(options.outputDir, data.filename)

		if err := os.WriteFile(outputPath, []byte(markdown), 0o600); err != nil {
			printStderr("  Error saving file: %v\n", err)
			continue
		}

		printStdout("  Saved: %s\n", outputPath)
		successCount++
	}

	printStdout("\nSuccessfully processed %d pages\n", successCount)

	return nil
}
