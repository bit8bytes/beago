// Package llms provides pipe.Handler adapters for integrating LLMs into a pipeline.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/bit8bytes/beago/pipe"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llm interface {
	Generate(ctx context.Context, messages []Message, w io.Writer) error
}

// Prompt returns a pipe.Handler that reads plain text from r, wraps it as a
// user message, and writes a JSON-encoded []Message to w. Each instruction
// is prepended as a system message.
func Prompt(instructions ...string) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		messages := make([]Message, 0, len(instructions)+1)
		for _, inst := range instructions {
			messages = append(messages, Message{Role: "system", Content: inst})
		}
		messages = append(messages, Message{Role: "user", Content: strings.TrimSpace(string(input))})
		return json.NewEncoder(w).Encode(messages)
	})
}

// Generate wraps an LLM implementation as a pipe.Handler.
// It accepts either JSON-encoded []Message or plain text as input.
// Plain text is treated as a user message.
func Generate(l llm) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		var messages []Message
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&messages); err != nil {
			messages = []Message{
				{Role: "user", Content: strings.TrimSpace(string(input))},
			}
		}

		return l.Generate(ctx, messages, w)
	})
}
