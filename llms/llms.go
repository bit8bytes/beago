// Package llms defines the core types used to communicate with language model backends.
// It provides a provider-agnostic interface so higher-level packages (agents, pipes, runner)
// can work with any LLM without being coupled to a specific API.
package llms

import "github.com/bit8bytes/beago/inputs/roles"

// Message represents a single turn in a conversation with an LLM.
// Role identifies who produced the content (e.g. system, user, assistant),
// which lets the model understand conversational context and respond appropriately.
type Message struct {
	Role    roles.Role `json:"role"`
	Content string     `json:"content"`
}

// StreamHandler is a callback invoked incrementally as the model generates a response.
// content holds the latest token(s) and done signals that the stream has ended.
// Returning a non-nil error cancels the stream.
type StreamHandler func(content string, done bool) error
