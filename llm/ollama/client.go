package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bit8bytes/beago/llm"
)

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

// Handle implements the pipe.Handler interface.
// It reads a prompt from r, sends it to Ollama, and streams the text response to w.
func (oc *Client) Generate(ctx context.Context, r io.Reader, w io.Writer) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read prompt from pipe: %w", err)
	}

	var messages []llm.Message
	dec := json.NewDecoder(bytes.NewReader(raw))
	for {
		var msg llm.Message
		if err := dec.Decode(&msg); err != nil {
			break
		}
		messages = append(messages, msg)
	}
	if len(messages) == 0 {
		messages = []llm.Message{{Role: "user", Content: strings.TrimSpace(string(raw))}}
	}

	request := ChatRequest{
		Model:    oc.Model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", oc.Endpoint, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned error status: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chatResp ChatResponse
		if err := decoder.Decode(&chatResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding stream: %w", err)
		}

		if _, err := io.WriteString(w, chatResp.Message.Content); err != nil {
			return err
		}

		if chatResp.Done {
			break
		}
	}

	return nil
}
