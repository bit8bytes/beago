package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	scriptRx = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleRx  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	tagsRx   = regexp.MustCompile(`<[^>]+>`)
	spaceRx  = regexp.MustCompile(`\s+`)
)

// Search is a tool for performing simple web searches.
// Its currently not production ready and is only used for testing.
type Search struct{}

func (s *Search) Name() string {
	return "search"
}

func (s *Search) Description() string {
	return "A tool for performing web searches. Use this tool to crawl and extract information from web pages."
}

func (s *Search) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "url",
			Description: "URL to search. This should be a valid URL string.",
			Required:    true,
		},
	}
}

func (s *Search) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return "", fmt.Errorf("search invalid params: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, input.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed with error: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Pragmatically extract text from HTML responses.
	// This is not a robust solution and requires more work to be production ready.
	text := string(b)
	text = scriptRx.ReplaceAllString(text, " ")
	text = styleRx.ReplaceAllString(text, " ")
	text = tagsRx.ReplaceAllString(text, " ")
	text = html.UnescapeString(text)
	text = spaceRx.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text, nil
}
