# CrawlDown

A Go-based web crawler that downloads and converts website content to Markdown format.

## Purpose

CrawlDown was initially designed to extract website content and convert it to Markdown format for use as context in AI agents and language models. By crawling and converting web content to clean Markdown, it makes website information easily consumable by AI systems that need structured, text-based context for processing, analysis, or response generation.

The tool is particularly useful for:

- Building knowledge bases for AI assistants
- Creating training or reference datasets from web content
- Preparing documentation for AI-powered search and retrieval systems
- Extracting structured information from websites for LLM context windows

## Features

- Web crawling with configurable depth
- HTML to Markdown conversion
- Extracts main content from pages
- Saves each page as a separate Markdown file
- Respects robots.txt by default
- Automatic filename generation from URLs
- Query parameter normalization (URLs with different parameter orders are treated as the same page)
- Path exclusion support (exclude specific URL paths from crawling)
- Filters non-HTTP protocols (mailto:, tel:, sms:, etc.)
- Smart email and phone number detection (even without protocol prefix)
- Configurable request timeout and delay
- Async crawling for better performance
- Named CLI arguments with short and long forms (powered by go-arg)

## Installation

### Using Go

```bash
go build -o crawldown ./cmd/main.go
```

### Using Task

First, install [Task](https://taskfile.dev/installation/):

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

Then build the project:

```bash
task build
```

## Usage

```bash
crawldown [OPTIONS] <url> <output-directory>
```

### Arguments

- `url` - The starting URL to crawl (required)
- `output-directory` - The directory where markdown files will be saved (required)

### Options

- `-d, --depth DEPTH` - Maximum crawl depth (default: 2)
- `-e, --exclude EXCLUDE` - URL path prefixes to exclude from crawling (can be specified multiple times)
- `-t, --timeout TIMEOUT` - Request timeout in seconds (default: 60)
- `--delay DELAY` - Delay between requests in seconds (default: 1)
- `-h, --help` - Display help message
- `--version` - Display version information

### Examples

```bash
# Crawl example.com with default settings
crawldown https://example.com ./output

# Crawl with custom depth
crawldown -d 3 https://example.com ./output
crawldown --depth 3 https://example.com ./output

# Crawl excluding specific paths
crawldown -d 3 -e "https://example.com/admin/" https://example.com ./output
crawldown -d 3 -e "https://example.com/admin/" -e "https://example.com/private/" https://example.com ./output

# Crawl with custom timeout and delay
crawldown -d 3 -t 30 --delay 2 https://example.com ./output
```

## Architecture

The project is organized into the following packages:

### cmd/

Contains the main CLI application entry point.

### src/crawler/

Handles web crawling functionality using [colly](https://github.com/gocolly/colly):

- Configurable crawl depth
- Domain filtering
- Main content extraction
- Link following

### src/converter/

Handles HTML to Markdown conversion using [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown):

- GitHub Flavored Markdown support
- Tables, task lists, and strikethrough
- Filename generation from URLs
- Content cleanup

## Development

### Available Tasks

View all available tasks:

```bash
task --list
```

### Common Tasks

```bash
# Build the binary
task build

# Run all tests with coverage
task test-cover

# Run linter
task lint

# Check for vulnerabilities
task vuln

# Run all checks (test, lint, vuln)
task check

# Generate detailed coverage report
task test-coverage

# Clean build artifacts
task clean

# Run example crawl
task example
```

### Manual Commands

#### Running Tests

```bash
go test -v -cover ./src/...
```

#### Linting

```bash
golangci-lint run ./src/...
```

#### Security Checks

```bash
govulncheck ./...
```

## Dependencies

- [github.com/gocolly/colly](https://github.com/gocolly/colly) - Web crawling
- [github.com/JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown conversion

## License

See LICENSE file for details.
