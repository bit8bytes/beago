package pipe

import (
	"context"
	"io"
)

func Tee(debug io.Writer) HandlerFunc {
	return HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := io.Copy(io.MultiWriter(w, debug), r)
		return err
	})
}

// Exit returns a Handler that calls f with the full stream content.
// If f returns true, the handler passes the content through and returns ErrDone,
// signalling the enclosing Loop to stop.
func Exit(f func([]byte) bool) Handler {
	return HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
		if f(data) {
			return ErrDone
		}
		return nil
	})
}
