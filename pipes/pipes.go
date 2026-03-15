// Package pipes provides a generic pipeline for generating structured output
// from a language model.
package pipes

import (
	"context"
	"errors"

	"github.com/bit8bytes/beago/llms"
)

type parser[T any] interface {
	Parse(text string) (T, error)
	// Instructions returns formatting instructions to append to the prompt.
	Instructions() string
}

type llm interface {
	Generate(ctx context.Context, messages []llms.Message) (string, error)
}

// Pipe is a generic pipeline that sends messages to a language model and parses
// the response into a structured type T.
type Pipe[T any] struct {
	messages []llms.Message
	model    llm
	parser   parser[T]
}

// New creates a new Pipe with the given messages, model, and parser.
func New[T any](messages []llms.Message, model llm, parser parser[T]) *Pipe[T] {
	return &Pipe[T]{
		messages: messages,
		model:    model,
		parser:   parser,
	}
}

// Invoke executes the pipeline by appending parser instructions to the first
// message, generating a response from the model, and parsing the result into
// a value of type T.
func (pipe *Pipe[T]) Invoke(ctx context.Context) (*T, error) {
	if len(pipe.messages) == 0 {
		return nil, errors.New("pipe has no messages")
	}

	instructions := pipe.parser.Instructions()
	msgs := pipe.messages
	if instructions != "" {
		// Copy first message so instructions aren't appended on repeated Invoke calls.
		first := pipe.messages[0]
		first.Content += " " + instructions
		msgs = append([]llms.Message{first}, pipe.messages[1:]...)
	}

	output, err := pipe.model.Generate(ctx, msgs)
	if err != nil {
		return nil, err
	}

	parsed, err := pipe.parser.Parse(output)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
