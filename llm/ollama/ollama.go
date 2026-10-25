package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bit8bytes/beago/llm"
)

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []llm.Message `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Model   string      `json:"model"`
	Message llm.Message `json:"message"`
	Done    bool        `json:"done"`
}

// Client represents an Ollama API client designed for streaming pipelines.
type Client struct {
	Model    string
	Endpoint string
}

// New creates a new Ollama client.
func New(model string, endpoint string) *Client {
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/chat"
	}
	return &Client{
		Model:    model,
		Endpoint: endpoint,
	}
}

// Generate implements the llm.Generator interface.
// It reads []llm.Message from r, sends them to Ollama, and streams the text response to w.
func (oc *Client) Generate(ctx context.Context, r io.Reader, w io.Writer) error {
	fail := func(err error) error {
		return fmt.Errorf("Generate: %w", err)
	}

	var messages []llm.Message
	if err := json.NewDecoder(r).Decode(&messages); err != nil {
		return fail(err)
	}

	body, err := json.Marshal(chatRequest{
		Model:    oc.Model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return fail(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", oc.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fail(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fail(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fail(fmt.Errorf("ollama returned error status: %d", resp.StatusCode))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chatResp chatResponse
		if err := decoder.Decode(&chatResp); err != nil {
			if err == io.EOF {
				break
			}
			return fail(err)
		}

		if _, err := io.WriteString(w, chatResp.Message.Content); err != nil {
			return fail(err)
		}

		if chatResp.Done {
			break
		}
	}

	return nil
}
