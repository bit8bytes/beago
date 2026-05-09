// Package llms provides pipe.Handler adapters for integrating LLMs into a pipeline.
package llm

import (
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
	Generate(ctx context.Context, r io.Reader, w io.Writer) error
}

// Generate wraps an LLM implementation as a pipe.Handler.
func Generate(l llm) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		return l.Generate(ctx, r, w)
	})
}

// Prompt prepends a system message with the given instruction to the pipeline,
// then forwards the remaining input as a user message.
func Prompt(instruction string) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(Message{Role: "system", Content: instruction}); err != nil {
			return err
		}
		return enc.Encode(Message{Role: "user", Content: strings.TrimSpace(string(input))})
	})
}
