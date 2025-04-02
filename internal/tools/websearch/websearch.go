package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/tools/core"
)

// Tool provides web search capabilities
type Tool struct {
	name        string
	description string
	parameters  map[string]core.Parameter
	httpClient  *http.Client
}

// New creates a new WebSearchTool
func New() *Tool {
	parameters := map[string]core.Parameter{
		"method": {
			Type:        "string",
			Description: "Method to use: 'visit' to visit a URL and read its content.",
			Required:    true,
		},
		"query": {
			Type:        "string",
			Description: "URL to visit",
			Required:    true,
		},
	}

	return &Tool{
		name:        "websearch",
		description: "Search the web, visit URLs, or conduct multi-step research. The 'research' method automatically searches Google, visits the top sites, extracts relevant information, and summarizes the findings - use this for any information-gathering task. When you need detailed information about any topic (current events, facts, news, or data), always use the 'research' method, not just 'google'.",
		parameters:  parameters,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Name returns the name of the tool
func (t *Tool) Name() string {
	return t.name
}

// Description returns the description of the tool
func (t *Tool) Description() string {
	return t.description
}

// Parameters returns the parameters for the tool
func (t *Tool) Parameters() map[string]core.Parameter {
	return t.parameters
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	// Extract method parameter
	method, ok := args["method"].(string)
	if !ok || method == "" {
		return core.ToolExecutionResult{}, fmt.Errorf("method must be a non-empty string ('visit')")
	}

	// Extract query parameter
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return core.ToolExecutionResult{}, fmt.Errorf("query must be a non-empty string")
	}

	// Only support the 'visit' method
	if strings.ToLower(method) != "visit" {
		return core.ToolExecutionResult{}, fmt.Errorf("unknown method: %s, supported method is 'visit'", method)
	}

	// Visit the URL
	result, err := t.VisitURL(ctx, query)
	if err != nil {
		return core.ToolExecutionResult{}, err
	}

	return core.ToolExecutionResult{
		ToolMethod: method,
		Output:     result,
	}, nil
}

// VisitURL visits a URL and returns its text content
func (t *Tool) VisitURL(ctx context.Context, urlStr string) (string, error) {
	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure the URL has a scheme
	if parsedURL.Scheme == "" {
		urlStr = "https://" + urlStr
		parsedURL, err = url.Parse(urlStr)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
	}

	// Create a request
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set a user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Execute the request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if content type is text-based
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") &&
		!strings.Contains(strings.ToLower(contentType), "text/plain") {
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extract readable content from HTML
	content := t.extractReadableContent(string(body))

	return fmt.Sprintf("Content from %s:\n\n%s", urlStr, content), nil
}

// extractReadableContent extracts the main content from an HTML page
func (t *Tool) extractReadableContent(html string) string {
	// Remove script and style elements
	scriptPattern := regexp.MustCompile(`(?s)<script.*?</script>`)
	stylePattern := regexp.MustCompile(`(?s)<style.*?</style>`)

	html = scriptPattern.ReplaceAllString(html, "")
	html = stylePattern.ReplaceAllString(html, "")

	// Replace common block-level elements with newlines
	html = regexp.MustCompile(`<(?:div|p|h[1-6]|table|tr|ul|ol)[^>]*>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`</(?:div|p|h[1-6]|table|tr|ul|ol)>`).ReplaceAllString(html, "\n")

	// Replace list items with bullet points
	html = regexp.MustCompile(`<li[^>]*>`).ReplaceAllString(html, "\n• ")

	// Remove all other HTML tags
	html = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, "")

	// Replace HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")
	html = strings.ReplaceAll(html, "&nbsp;", " ")

	// Replace multiple newlines with a single one
	html = regexp.MustCompile(`\n{3,}`).ReplaceAllString(html, "\n\n")

	// Trim whitespace
	html = strings.TrimSpace(html)

	// Limit content length to avoid very long responses
	const maxLength = 8000
	if len(html) > maxLength {
		html = html[:maxLength] + "...\n[Content truncated due to length]"
	}

	return html
}
