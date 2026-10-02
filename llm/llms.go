// Package llms provides pipe.Handler adapters for integrating LLMs into a pipeline.
package llm

import (
	"context"
	"io"

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
