// Package pipe provides primitives for building streaming data pipelines.
package pipe

import (
	"context"
	"errors"
	"io"
)

// ErrDone is returned by a Handler to signal the enclosing Loop to stop.
// It is not propagated as an error to the caller.
var ErrDone = errors.New("pipe: done")

// A Handler responds to a text stream.
//
// Handle should read data from r, transform it, and write the result to w.
// It is the core primitive for building "Unix-style" AI agent pipelines.
type Handler interface {
	Handle(ctx context.Context, r io.Reader, w io.Writer) error
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions
// as stream handlers. If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler that calls f.
type HandlerFunc func(ctx context.Context, r io.Reader, w io.Writer) error

// Handle calls f(ctx, r, w).
func (f HandlerFunc) Handle(ctx context.Context, r io.Reader, w io.Writer) error {
	return f(ctx, r, w)
}
