package history

import (
	"bytes"
	"context"
	"io"

	"github.com/bit8bytes/beago/pipe"
)

// History accumulates stream content across loop iterations.
type History struct {
	buf bytes.Buffer
}

func New() *History {
	return &History{}
}

// Accumulate returns a Handler that appends the current input to the history
// and writes the full accumulated context to the next stage.
func (h *History) Accumulate() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if _, err := io.Copy(&h.buf, r); err != nil {
			return err
		}
		_, err := w.Write(h.buf.Bytes())
		return err
	})
}
