package embedder

import (
	"context"
	"errors"
	"testing"
)

type mockLLM struct {
	embedding []float32
	err       error
}

func (m *mockLLM) GenerateEmbedding(_ context.Context, _ string) ([]float32, error) {
	return m.embedding, m.err
}

func TestEmbed(t *testing.T) {
	want := []float32{0.1, 0.2, 0.3}
	e := New(&mockLLM{embedding: want})

	got, err := e.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d values, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %f, want %f", i, got[i], want[i])
		}
	}
}

func TestEmbed_Error(t *testing.T) {
	e := New(&mockLLM{err: errors.New("llm failure")})

	_, err := e.Embed(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
