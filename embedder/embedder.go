// Package embedder provides a thin wrapper around an LLM's embedding capability.
// Keeping embedding behind its own interface isolates callers from the underlying
// model provider and makes it straightforward to swap or mock the backend in tests.
package embedder

import (
	"context"
)

// llm is the subset of an LLM client that this package requires.
// Using a narrow interface instead of a concrete type keeps the package decoupled
// from any specific provider implementation.
type llm interface {
	GenerateEmbedding(ctx context.Context, prompt string) ([]float32, error)
}

// embedder wraps an LLM to expose only the embedding operation.
type embedder struct {
	llm llm
}

// New creates an embedder backed by the given LLM.
func New(llm llm) *embedder {
	return &embedder{
		llm: llm,
	}
}

// Embed converts query into a vector representation using the underlying LLM.
// The returned slice can be used for semantic similarity comparisons, retrieval, or storage in a vector store.
func (e *embedder) Embed(ctx context.Context, query string) ([]float32, error) {
	return e.llm.GenerateEmbedding(ctx, query)
}
