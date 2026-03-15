// Package memory provides an in-memory Store for use in tests and local
// development. It is not intended for production — all messages are lost when
// the process exits and there is no size bound on the message history.
package memory

import (
	"context"
	"sync"

	"github.com/bit8bytes/beago/llms"
	"github.com/bit8bytes/beago/stores"
)

type store struct {
	mu       sync.Mutex
	messages []llms.Message
}

func New() stores.Store {
	return &store{}
}

func (s *store) Add(_ context.Context, msgs ...llms.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msgs...)
	return nil
}

func (s *store) Clear(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Retain the underlying array to avoid an allocation on the next Add.
	s.messages = s.messages[:0]
	return nil
}

func (s *store) List(_ context.Context) ([]llms.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Copy so the caller cannot mutate the store's internal slice after the
	// lock is released.
	out := make([]llms.Message, len(s.messages))
	copy(out, s.messages)

	return out, nil
}

func (s *store) Close() error { return nil }
