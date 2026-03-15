// Package stores defines the Store interface for persisting conversation
// history between agent turns.
//
// LLMs are stateless — they have no memory of prior interactions unless the
// full message history is included in every request. A Store gives the agent
// that memory by accumulating messages as the conversation progresses and
// replaying them on each LLM call.
//
// The package ships with an in-memory implementation (see package
// stores/memory) suitable for testing and development. Production use cases
// that require durability across restarts should use an external
// implementation backed by a database.
package stores

import (
	"context"

	"github.com/bit8bytes/beago/llms"
)

// Store is the interface that wraps the basic message persistence methods.
//
// Implementations must be safe for concurrent use by multiple goroutines.
//
// # Tamper-evident chain
//
// Every implementation must call [Stamp] on each message inside Add, passing
// the hash of the last stored message as prevHash (empty string for the first
// message). Stamp sets Message.Timestamp and computes Message.Hash as:
//
//	SHA256(prevHash + role + timestamp + content)
//
// This chains every message to its predecessor. Any modification, insertion,
// or deletion of a message breaks every subsequent hash, making the history
// tamper-evident. Callers can verify integrity by replaying the chain with
// [Verify].
//
// # Minimal implementation checklist
//
//   - Add: call Stamp on each message before persisting; track lastHash
//   - List: return messages in insertion order; never mutate the stored slice
//   - Clear: reset lastHash to "" alongside clearing messages
//   - Close: release database connections or file handles; no-op is valid
type Store interface {
	// Add appends one or more messages to the store. The agent calls this after
	// every LLM turn so the full conversation history is available on the next
	// Plan call. Implementations must stamp each message via [Stamp] before
	// persisting to maintain the hash chain.
	Add(ctx context.Context, msgs ...llms.Message) error

	// List returns all messages in insertion order. The agent passes the full
	// history to the LLM on every turn so it has context from prior steps.
	List(ctx context.Context) ([]llms.Message, error)

	// Clear removes all messages and resets the hash chain, allowing the store
	// to be reused across independent agent runs.
	Clear(ctx context.Context) error

	// Close releases any resources held by the store (e.g. database connections).
	// Call this when the agent is done to avoid resource leaks.
	Close() error
}
