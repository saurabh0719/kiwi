package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	// Using free Serper API by default, but can be configured with other providers
	searchAPIURL string
	apiKey       string
}

// New creates a new WebSearchTool
func New(apiKey string) *Tool {
	parameters := map[string]core.Parameter{
		"query": {
			Type:        "string",
			Description: "Search query string to look up on the web",
			Required:    true,
		},
		"num_results": {
			Type:        "integer",
			Description: "Number of search results to return (default: 5)",
			Required:    false,
			Default:     5,
		},
	}

	return &Tool{
		name:        "websearch",
		description: "Search the web for information about a topic",
		parameters:  parameters,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		searchAPIURL: "https://google.serper.dev/search",
		apiKey:       apiKey,
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
	// Extract query parameter
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return core.ToolExecutionResult{}, fmt.Errorf("query must be a non-empty string")
	}

	// Extract numResults parameter with default fallback
	numResults := 5
	if numResultsArg, ok := args["num_results"]; ok {
		// Try to convert to int
		if numInt, ok := numResultsArg.(float64); ok {
			numResults = int(numInt)
		} else if numInt, ok := numResultsArg.(int); ok {
			numResults = numInt
		}
	}

	// Limit numResults to reasonable range
	if numResults < 1 {
		numResults = 1
	} else if numResults > 10 {
		numResults = 10
	}

	// If no API key is configured, return an error with instructions
	if t.apiKey == "" {
		return core.ToolExecutionResult{
			ToolMethod: "search",
			Output:     "Error: Web search API key not configured. Please set the SERPER_API_KEY environment variable or configure it in your config file.",
		}, nil
	}

	// Perform the search
	results, err := t.search(ctx, query, numResults)
	if err != nil {
		return core.ToolExecutionResult{}, fmt.Errorf("search failed: %w", err)
	}

	return core.ToolExecutionResult{
		ToolMethod: "search",
		Output:     results,
	}, nil
}

// search performs the actual web search
func (t *Tool) search(ctx context.Context, query string, numResults int) (string, error) {
	// Prepare request body
	reqBody, err := json.Marshal(map[string]interface{}{
		"q": query,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", t.searchAPIURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", t.apiKey)

	// Execute the request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Format the search results
	return t.formatResults(result, numResults), nil
}

// formatResults formats the search results into a readable string
func (t *Tool) formatResults(result map[string]interface{}, numResults int) string {
	var sb strings.Builder

	// Add organic search results
	if organic, ok := result["organic"].([]interface{}); ok {
		sb.WriteString("Search Results:\n\n")

		// Limit to specified number of results
		count := 0
		for _, item := range organic {
			result, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract fields
			title, _ := result["title"].(string)
			link, _ := result["link"].(string)
			snippet, _ := result["snippet"].(string)

			// Format result
			sb.WriteString(fmt.Sprintf("%d. %s\n", count+1, title))
			sb.WriteString(fmt.Sprintf("   URL: %s\n", link))
			sb.WriteString(fmt.Sprintf("   %s\n\n", snippet))

			count++
			if count >= numResults {
				break
			}
		}
	}

	return sb.String()
}

// SearchWithDuckDuckGo is an alternative search method using DuckDuckGo
// This is implemented as a fallback but not currently used
func (t *Tool) SearchWithDuckDuckGo(ctx context.Context, query string) (string, error) {
	// Encode the query for URL
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", encodedQuery)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", err
	}

	// Set a user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Send request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// In a real implementation, we would parse the HTML response here
	// For simplicity, we just return a message
	return fmt.Sprintf("DuckDuckGo search for '%s' returned %d bytes of HTML. Parsing HTML responses requires a full HTML parser, which is outside the scope of this example.", query, len(body)), nil
}
