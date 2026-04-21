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
- Subcommands with backward-compatible root execution (powered by Cobra)
- Agent skill scaffold generation for CrawlDown automation
- GoReleaser + UPX release pipeline for version tags

## Installation

### Using Go

```bash
go build -o crawldown ./cmd
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
crawldown [flags] <url>
crawldown get [flags] <url>
crawldown add-skill <name> [flags]
```

### Crawl Arguments

- `url` - The starting URL to crawl when `--single` is not used

### Crawl Options

- `-o, --output DIR` - The directory where Markdown files will be saved (required)
- `-d, --depth DEPTH` - Maximum crawl depth (default: 2)
- `-e, --exclude PATH` - URL path prefixes to exclude from crawling (can be specified multiple times)
- `-t, --timeout TIMEOUT` - Request timeout in seconds (default: 60)
- `--delay DELAY` - Delay between requests in seconds (default: 1)
- `-s, --single URL` - Download a single page URL instead of crawling from the positional URL
- `--ignore-robots-txt` - Ignore robots.txt while crawling
- `--follow-external-links` - Allow following external links
- `--user-agent VALUE` - Override the default HTTP user agent
- `-h, --help` - Display help message
- `--version` - Display version information

### add-skill Options

- `--base-dir DIR` - Base directory where the `.agents/skills` scaffold will be created (default: current directory)
- `--binary NAME` - Binary name to embed in the generated skill instructions (default: `crawldown`)
- `--force` - Overwrite an existing `SKILL.md`

### Examples

```bash
# Crawl example.com with default settings
crawldown -o ./output https://example.com

# Same crawl using the explicit get subcommand
crawldown get -o ./output https://example.com

# Crawl with custom depth
crawldown get -o ./output -d 3 https://example.com
crawldown get -o ./output --depth 3 https://example.com

# Crawl excluding specific paths
crawldown get -o ./output -d 3 -e /admin https://example.com
crawldown get -o ./output -d 3 -e /admin -e /private https://example.com

# Crawl with custom timeout and delay
crawldown get -o ./output -d 3 -t 30 --delay 2 https://example.com

# Download a single indicated page
crawldown get -o ./output -s "https://example.com/articles/2025/interesting.html"

# Create an agent skill scaffold in the current directory
crawldown add-skill site-fetch

# Create an agent skill scaffold with a custom binary name
crawldown add-skill site-fetch --binary ./crawldown

```

## Architecture

The project is organized into the following packages:

### cmd/

Contains the Cobra-based CLI application entry point and subcommands.

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

# Validate the GoReleaser configuration
task release-check

# Build a local release snapshot
task release-snapshot
```

### Manual Commands

#### Running Tests

```bash
go test -v -cover ./...
```

#### Linting

```bash
golangci-lint run ./...
```

#### Security Checks

```bash
govulncheck ./...
```

## Dependencies

- [github.com/gocolly/colly](https://github.com/gocolly/colly) - Web crawling
- [github.com/JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown conversion
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI command structure

## Release Process

Push a tag with the `v` prefix, for example `v1.1.0`, to trigger the GitHub Actions release workflow.

The workflow:

- runs tests
- installs UPX on the runner
- builds release archives with GoReleaser
- publishes artifacts to the GitHub release that matches the tag

To verify the release configuration locally without publishing:

```bash
task release-check
task release-snapshot
```

## License

See LICENSE file for details.
