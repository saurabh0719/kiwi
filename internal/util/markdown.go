package util

import (
	"regexp"
)

// ANSI color codes for terminal output
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Cyan      = "\033[36m"
	Yellow    = "\033[33m"
	Green     = "\033[32m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	DarkGray  = "\033[90m"
	LightGray = "\033[37m"

	// Background colors
	BgDarkGray  = "\033[100m"
	BgLightGray = "\033[47m"
	BgBlack     = "\033[40m"
)

// RenderMarkdown formats text with styling for headers, bold, and italic
// Perfect for streaming content - simple and handles partial chunks
func RenderMarkdown(text string, shouldRender ...bool) string {
	// Check if rendering is disabled by caller
	if len(shouldRender) > 0 && !shouldRender[0] {
		return text
	}

	// Don't try to render very small chunks (likely partial words in a stream)
	if len(text) < 5 {
		return text
	}

	// Format headers with proportional styling (h1-h6)
	text = regexp.MustCompile(`(?m)^#\s+(.+)$`).ReplaceAllString(text, Bold+Cyan+"$1"+Reset)
	text = regexp.MustCompile(`(?m)^##\s+(.+)$`).ReplaceAllString(text, Bold+Yellow+"$1"+Reset)
	text = regexp.MustCompile(`(?m)^###\s+(.+)$`).ReplaceAllString(text, Bold+Green+"$1"+Reset)
	text = regexp.MustCompile(`(?m)^####\s+(.+)$`).ReplaceAllString(text, Bold+Blue+"$1"+Reset)
	text = regexp.MustCompile(`(?m)^#####\s+(.+)$`).ReplaceAllString(text, Bold+Magenta+"$1"+Reset)
	text = regexp.MustCompile(`(?m)^######\s+(.+)$`).ReplaceAllString(text, DarkGray+"$1"+Reset)

	// Format inline code with a background highlight
	text = regexp.MustCompile("`([^`\n]+?)`").ReplaceAllString(text, BgDarkGray+" $1 "+Reset)

	// Format bold and italic text
	text = regexp.MustCompile(`\*\*([^*\n]+?)\*\*`).ReplaceAllString(text, Bold+"$1"+Reset)
	text = regexp.MustCompile(`\*([^*\n]+?)\*`).ReplaceAllString(text, Italic+"$1"+Reset)
	text = regexp.MustCompile(`_([^_\n]+?)_`).ReplaceAllString(text, Italic+"$1"+Reset)

	return text
}

// ShouldRenderMarkdown returns true if the feature is enabled
func ShouldRenderMarkdown(renderMarkdownFlag bool) bool {
	return renderMarkdownFlag
}
