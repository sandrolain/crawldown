package converter

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
)

// Options defines converter configuration
type Options struct {
	Domain           string
	EscapeMode       string
	BulletListMarker string
	CodeBlockStyle   string
	EmDelimiter      string
	StrongDelimiter  string
	LinkStyle        string
}

// Converter handles HTML to Markdown conversion
type Converter struct {
	converter *md.Converter
	options   Options
}

// NewConverter creates a new converter instance
func NewConverter(opts Options) (*Converter, error) {
	converter := md.NewConverter(opts.Domain, true, nil)

	// Add plugins for better conversion
	converter.Use(plugin.GitHubFlavored())
	converter.Use(plugin.Table())
	converter.Use(plugin.TaskListItems())
	converter.Use(plugin.Strikethrough("~~"))

	return &Converter{
		converter: converter,
		options:   opts,
	}, nil
}

// Convert converts HTML content to Markdown
func (c *Converter) Convert(html string) (string, error) {
	if html == "" {
		return "", fmt.Errorf("empty HTML content")
	}

	markdown, err := c.converter.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("conversion failed: %w", err)
	}

	// Clean up the markdown
	markdown = c.cleanMarkdown(markdown)

	return markdown, nil
}

// cleanMarkdown performs post-processing cleanup on the markdown
func (c *Converter) cleanMarkdown(markdown string) string {
	// Remove excessive newlines (more than 2 consecutive)
	re := regexp.MustCompile(`\n{3,}`)
	markdown = re.ReplaceAllString(markdown, "\n\n")

	// Trim leading and trailing whitespace
	markdown = strings.TrimSpace(markdown)

	return markdown
}

// ConvertLinksToLocal converts absolute URLs to local markdown file references
func ConvertLinksToLocal(markdown string, baseURL string, urlToFileMap map[string]string) string {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return markdown
	}

	// Replace markdown links [text](url) with local file references
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	markdown = re.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		linkText := parts[1]
		linkURL := parts[2]

		// Skip anchor links, external protocols, and fragments
		if strings.HasPrefix(linkURL, "#") ||
			strings.HasPrefix(linkURL, "mailto:") ||
			strings.HasPrefix(linkURL, "javascript:") {
			return match
		}

		// Parse the link URL
		parsedLink, err := url.Parse(linkURL)
		if err != nil {
			return match
		}

		// Make relative URLs absolute
		if !parsedLink.IsAbs() {
			parsedLink = parsedBase.ResolveReference(parsedLink)
		}

		// Check if we have a local file for this URL
		// Try with full URL including query parameters (normalized without trailing slash)
		fullURL := parsedLink.Scheme + "://" + parsedLink.Host + strings.TrimSuffix(parsedLink.Path, "/")
		if parsedLink.RawQuery != "" {
			fullURL += "?" + parsedLink.RawQuery
		}

		if localFile, exists := urlToFileMap[fullURL]; exists {
			// Convert to local markdown file reference
			if parsedLink.Fragment != "" {
				return fmt.Sprintf("[%s](%s#%s)", linkText, localFile, parsedLink.Fragment)
			}
			return fmt.Sprintf("[%s](%s)", linkText, localFile)
		}

		// Try without query parameters as fallback (also normalized)
		cleanURL := parsedLink.Scheme + "://" + parsedLink.Host + strings.TrimSuffix(parsedLink.Path, "/")
		if localFile, exists := urlToFileMap[cleanURL]; exists {
			// Convert to local markdown file reference
			if parsedLink.Fragment != "" {
				return fmt.Sprintf("[%s](%s#%s)", linkText, localFile, parsedLink.Fragment)
			}
			return fmt.Sprintf("[%s](%s)", linkText, localFile)
		}

		// Keep external links as-is
		return match
	})

	return markdown
}

// GenerateFilename creates a safe filename from a URL
func GenerateFilename(pageURL string) string {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return "index.md"
	}

	path := parsedURL.Path
	query := parsedURL.RawQuery

	if path == "" || path == "/" {
		// Handle query parameters for root path
		if query != "" {
			filename := "index-" + sanitizeFilename(query)
			if !strings.HasSuffix(filename, ".md") {
				filename += ".md"
			}
			return filename
		}
		return "index.md"
	}

	// Remove leading slash and clean the path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	// Replace slashes with dashes for subdirectories
	filename := strings.ReplaceAll(path, "/", "-")

	// Append query parameters to filename if present
	if query != "" {
		filename = filename + "-" + query
	}

	// Remove or replace invalid characters
	filename = sanitizeFilename(filename)

	// Add .md extension if not present
	if !strings.HasSuffix(filename, ".md") {
		if filepath.Ext(filename) != "" {
			filename = strings.TrimSuffix(filename, filepath.Ext(filename))
		}
		filename += ".md"
	}

	return filename
}

// sanitizeFilename removes or replaces invalid filename characters
func sanitizeFilename(filename string) string {
	// Replace invalid characters with dash (including = and & from query params)
	re := regexp.MustCompile(`[<>:"/\\|?*=&]`)
	filename = re.ReplaceAllString(filename, "-")

	// Remove multiple consecutive dashes
	re = regexp.MustCompile(`-+`)
	filename = re.ReplaceAllString(filename, "-")

	// Trim dashes from start and end
	filename = strings.Trim(filename, "-")

	// If empty, return default
	if filename == "" {
		filename = "page"
	}

	return filename
}
