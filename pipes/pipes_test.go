package pipes

import (
	"context"
	"errors"
	"testing"

	"github.com/bit8bytes/beago/llms"
)

type mockLLM struct {
	result string
	err    error
}

func (m *mockLLM) Generate(_ context.Context, _ []llms.Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.result, nil
}

type mockParser struct {
	instructions string
	result       string
	err          error
}

func (m *mockParser) Parse(_ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.result, nil
}

func (m *mockParser) Instructions() string {
	return m.instructions
}

func TestInvoke(t *testing.T) {
	t.Run("returns parsed result", func(t *testing.T) {
		pipe := New(
			[]llms.Message{{Content: "hello"}},
			&mockLLM{result: "raw"},
			&mockParser{result: "parsed"},
		)
		got, err := pipe.Invoke(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *got != "parsed" {
			t.Errorf("got %q, want %q", *got, "parsed")
		}
	})

	t.Run("errors when no messages", func(t *testing.T) {
		pipe := New([]llms.Message{}, &mockLLM{}, &mockParser{})
		_, err := pipe.Invoke(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("errors when llm fails", func(t *testing.T) {
		pipe := New(
			[]llms.Message{{Content: "hello"}},
			&mockLLM{err: errors.New("llm error")},
			&mockParser{},
		)
		_, err := pipe.Invoke(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("errors when parser fails", func(t *testing.T) {
		pipe := New(
			[]llms.Message{{Content: "hello"}},
			&mockLLM{result: "raw"},
			&mockParser{err: errors.New("parse error")},
		)
		_, err := pipe.Invoke(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("instructions not duplicated on repeated Invoke", func(t *testing.T) {
		llm := &mockLLM{result: "raw"}
		var capturedContent string
		// Use a custom llm to capture the message content
		capturingLLM := &captureLLM{result: "raw", captured: &capturedContent}
		pipe := New(
			[]llms.Message{{Content: "hello"}},
			capturingLLM,
			&mockParser{instructions: "format as JSON", result: "parsed"},
		)

		pipe.Invoke(context.Background())
		first := capturedContent
		pipe.Invoke(context.Background())
		second := capturedContent

		if first != second {
			t.Errorf("instructions duplicated on second call:\n first: %q\nsecond: %q", first, second)
		}
		_ = llm
	})
}

type captureLLM struct {
	result   string
	captured *string
}

func (c *captureLLM) Generate(_ context.Context, msgs []llms.Message) (string, error) {
	if len(msgs) > 0 {
		*c.captured = msgs[0].Content
	}
	return c.result, nil
}
