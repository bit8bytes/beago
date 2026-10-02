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
