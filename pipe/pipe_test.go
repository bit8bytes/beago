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

func TestExecute_SingleHandler(t *testing.T) {
	var out bytes.Buffer
	err := pipe.Execute(context.Background(), strings.NewReader("hello"), &out, uppercase())
	if err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "HELLO" {
		t.Errorf("got %q, want %q", got, "HELLO")
	}
}

func TestExecute_ChainedHandlers(t *testing.T) {
	var out bytes.Buffer
	err := pipe.Execute(context.Background(), strings.NewReader("hi"),
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

func TestExecute_PropagatesHandlerError(t *testing.T) {
	boom := errors.New("boom")
	fail := pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		return boom
	})

	var out bytes.Buffer
	err := pipe.Execute(context.Background(), strings.NewReader("x"), &out, uppercase(), fail)
	if !errors.Is(err, boom) {
		t.Errorf("expected boom error, got %v", err)
	}
}

func TestExecute_NoHandlers(t *testing.T) {
	// With no handlers Execute should return nil without writing anything.
	var out bytes.Buffer
	err := pipe.Execute(context.Background(), strings.NewReader("x"), &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestLoop_RunsUpToMaxIter(t *testing.T) {
	n := 0
	h := pipe.Loop(3, counter(&n))

	var out bytes.Buffer
	if err := h.Handle(context.Background(), strings.NewReader("x"), &out); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("expected 3 iterations, got %d", n)
	}
}

func TestLoop_StopsOnErrDone(t *testing.T) {
	n := 0
	stopAfter := 2

	h := pipe.Loop(10, counter(&n), pipe.Exit(func(data []byte) bool {
		return n >= stopAfter
	}))

	var out bytes.Buffer
	if err := h.Handle(context.Background(), strings.NewReader("x"), &out); err != nil {
		t.Fatal(err)
	}
	if n != stopAfter {
		t.Errorf("expected %d iterations, got %d", stopAfter, n)
	}
}

func TestLoop_OutputIsLastIterationResult(t *testing.T) {
	n := 0
	h := pipe.Loop(3, appendSuffix("."), pipe.Exit(func(data []byte) bool {
		n++
		return n >= 3
	}))

	var out bytes.Buffer
	if err := h.Handle(context.Background(), strings.NewReader("x"), &out); err != nil {
		t.Fatal(err)
	}
	// After 3 iterations of appending ".", output should be "x..."
	if got := out.String(); got != "x..." {
		t.Errorf("got %q, want %q", got, "x...")
	}
}

func TestLoop_PropagatesHandlerError(t *testing.T) {
	boom := errors.New("boom")
	fail := pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		return boom
	})

	h := pipe.Loop(5, fail)
	var out bytes.Buffer
	err := h.Handle(context.Background(), strings.NewReader("x"), &out)
	if !errors.Is(err, boom) {
		t.Errorf("expected boom, got %v", err)
	}
}

func TestTee_WritesToBothSinks(t *testing.T) {
	var primary, debug bytes.Buffer
	h := pipe.Tee(&debug)

	err := h.Handle(context.Background(), strings.NewReader("hello"), &primary)
	if err != nil {
		t.Fatal(err)
	}
	if primary.String() != "hello" {
		t.Errorf("primary got %q, want %q", primary.String(), "hello")
	}
	if debug.String() != "hello" {
		t.Errorf("debug got %q, want %q", debug.String(), "hello")
	}
}

func TestExit_ReturnsDoneWhenPredicateTrue(t *testing.T) {
	h := pipe.Exit(func(data []byte) bool { return true })

	var out bytes.Buffer
	err := h.Handle(context.Background(), strings.NewReader("done"), &out)
	if !errors.Is(err, pipe.ErrDone) {
		t.Errorf("expected ErrDone, got %v", err)
	}
	// Content must still be passed through.
	if out.String() != "done" {
		t.Errorf("got %q, want %q", out.String(), "done")
	}
}

func TestExit_ContinuesWhenPredicateFalse(t *testing.T) {
	h := pipe.Exit(func(data []byte) bool { return false })

	var out bytes.Buffer
	err := h.Handle(context.Background(), strings.NewReader("keep going"), &out)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestHandlerFunc_ImplementsHandler(t *testing.T) {
	// Compile-time check that HandlerFunc satisfies Handler.
	var _ pipe.Handler = pipe.HandlerFunc(nil)
}

func TestExecute_ContextCancellation(t *testing.T) {
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
	err := pipe.Execute(ctx, strings.NewReader("x"), &out, slow)
	if err == nil {
		t.Error("expected an error due to cancelled context")
	}
}
