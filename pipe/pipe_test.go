package pipe_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/bit8bytes/beago/pipe"
)

// uppercase transforms input to uppercase as a simple deterministic handler.
func uppercase() pipe.HandlerFunc {
	return func(ctx context.Context, r io.Reader, w io.Writer) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes.ToUpper(data))
		return err
	}
}

// append returns a handler that appends suffix to whatever it reads.
func appendSuffix(suffix string) pipe.HandlerFunc {
	return func(ctx context.Context, r io.Reader, w io.Writer) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = w.Write(append(data, []byte(suffix)...))
		return err
	}
}

// counter counts how many times it was invoked.
func counter(n *int) pipe.HandlerFunc {
	return func(ctx context.Context, r io.Reader, w io.Writer) error {
		*n++
		_, err := io.Copy(w, r)
		return err
	}
}

func TestDo_SingleHandler(t *testing.T) {
	var out bytes.Buffer
	err := pipe.Do(context.Background(), strings.NewReader("hello"), &out, uppercase())
	if err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "HELLO" {
		t.Errorf("got %q, want %q", got, "HELLO")
	}
}

func TestDo_ChainedHandlers(t *testing.T) {
	var out bytes.Buffer
	err := pipe.Do(context.Background(), strings.NewReader("hi"),
		&out,
		uppercase(),
		appendSuffix("!"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "HI!" {
		t.Errorf("got %q, want %q", got, "HI!")
	}
}

func TestDo_PropagatesHandlerError(t *testing.T) {
	boom := errors.New("boom")
	fail := pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		return boom
	})

	var out bytes.Buffer
	err := pipe.Do(context.Background(), strings.NewReader("x"), &out, uppercase(), fail)
	if !errors.Is(err, boom) {
		t.Errorf("expected boom error, got %v", err)
	}
}

func TestDo_NoHandlers(t *testing.T) {
	// With no handlers Do should return nil without writing anything.
	var out bytes.Buffer
	err := pipe.Do(context.Background(), strings.NewReader("x"), &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestLoop_CancelsOnContextDone(t *testing.T) {
	n := 0
	ctx, cancel := context.WithCancel(context.Background())

	h := pipe.Loop(pipe.HandlerFunc(func(c context.Context, r io.Reader, w io.Writer) error {
		n++
		if n >= 3 {
			cancel()
		}
		_, err := io.Copy(w, r)
		return err
	}))

	var out bytes.Buffer
	err := h.Handle(ctx, strings.NewReader("x"), &out)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 iterations before cancel, got %d", n)
	}
}

func TestLoop_PropagatesHandlerError(t *testing.T) {
	boom := errors.New("boom")
	fail := pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		return boom
	})

	h := pipe.Loop(fail)
	var out bytes.Buffer
	err := h.Handle(context.Background(), strings.NewReader("x"), &out)
	if !errors.Is(err, boom) {
		t.Errorf("expected boom, got %v", err)
	}
}

func TestHandlerFunc_ImplementsHandler(t *testing.T) {
	// Compile-time check that HandlerFunc satisfies Handler.
	var _ pipe.Handler = pipe.HandlerFunc(nil)
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	slow := pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := io.Copy(w, r)
			return err
		}
	})

	var out bytes.Buffer
	err := pipe.Do(ctx, strings.NewReader("x"), &out, slow)
	if err == nil {
		t.Error("expected an error due to cancelled context")
	}
}
