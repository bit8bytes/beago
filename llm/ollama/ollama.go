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
	model    string
	endpoint string
	stream   bool
}

// Option configures a Client.
type Option func(*Client)

// WithModel sets the model to use for generation.
func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

// WithEndpoint sets the Ollama API endpoint URL.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithStream enables or disables streaming mode for the client.
func WithStream(stream bool) Option {
	return func(c *Client) {
		c.stream = stream
	}
}

// New creates a new Ollama client.
func New(opts ...Option) *Client {
	c := &Client{
		endpoint: "http://localhost:11434/api/chat",
		stream:   true,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.model == "" {
		panic("model must be specified")
	}
	return c
}

// Generate implements the llm.Generator interface.
func (oc *Client) Generate(ctx context.Context, messages []llm.Message, w io.Writer) error {
	fail := func(err error) error {
		return fmt.Errorf("Generate: %w", err)
	}

	body, err := json.Marshal(chatRequest{
		Model:    oc.model,
		Messages: messages,
		Stream:   oc.stream,
	})
	if err != nil {
		return fail(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oc.endpoint, bytes.NewReader(body))
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
