package history

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/pipe"
)

// History accumulates stream content across loop iterations.
type History struct {
	buf bytes.Buffer
}

func New() *History {
	return &History{}
}

// Replay writes the full accumulated history to the next stage, ignoring its input.
func (h *History) Replay() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := io.Copy(io.Discard, r); err != nil {
			return err
		}
		_, err := w.Write(h.buf.Bytes())
		return err
	})
}

// System stores the incoming bytes as-is (expected to be JSON-encoded llm.Message objects)
// and passes them through.
func (h *History) System() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		h.buf.Write(data)
		_, err = w.Write(data)
		return err
	})
}

// User wraps the input in a user llm.Message, stores it, and passes the raw content through.
func (h *History) User() pipe.Handler {
	return h.record("user")
}

// Assistant wraps the input in an assistant llm.Message, stores it, and passes the raw content through.
func (h *History) Assistant() pipe.Handler {
	return h.record("assistant")
}

func (h *History) record(role string) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		json.NewEncoder(&h.buf).Encode(llm.Message{Role: role, Content: string(data)})
		_, err = w.Write(data)
		return err
	})
}
