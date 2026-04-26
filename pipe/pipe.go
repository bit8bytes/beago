package pipe

import (
	"bytes"
	"context"
	"errors"
	"io"
)

// Loop runs handlers repeatedly, feeding each iteration's output as the next
// iteration's input. It stops when a handler returns ErrDone or maxIter is
// reached. Only the final output is written to w.
func Loop(maxIter int, handlers ...Handler) Handler {
	return HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		for range maxIter {
			var buf bytes.Buffer
			err := Execute(ctx, r, &buf, handlers...)
			if errors.Is(err, ErrDone) {
				_, copyErr := io.Copy(w, &buf)
				return copyErr
			}
			if err != nil {
				return err
			}
			r = &buf
		}
		return nil
	})
}

func Execute(ctx context.Context, r io.Reader, w io.Writer, handlers ...Handler) error {
	for i, h := range handlers {
		// The last handler will write diretcly to our final destination w.
		// All other handlers will write to the [io.Pipe]
		if i == len(handlers)-1 {
			return h.Handle(ctx, r, w)
		}

		pr, pw := io.Pipe() // Create a new pipe for each N-1 handler.

		go func(handler Handler, in io.Reader, pw *io.PipeWriter) error {
			err := handler.Handle(ctx, in, pw)
			return pw.CloseWithError(err)
		}(h, r, pw)

		r = pr // The next reader will read from the previous pipe's reader
	}
	return nil
}
