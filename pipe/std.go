package pipe

import (
	"context"
	"io"
)

func Tee(debug io.Writer) HandlerFunc {
	return HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		// MultiWriter creates a single writer for multiple writers.
		// Here, it is a Y-splitter that writes to w and debug.
		mw := io.MultiWriter(w, debug)

		// Here, we pipe everything into our Y-splitter.
		_, err := io.Copy(mw, r)
		return err
	})
}

// Exit returns a Handler that calls f with the full stream content.
// If f returns true, the handler passes the content through and returns ErrDone,
// signalling the enclosing Loop to stop.
func Exit(f func([]byte) bool) Handler {
	return HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
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
