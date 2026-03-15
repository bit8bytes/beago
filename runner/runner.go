// Package runner separates the loop control logic from agent logic so that
// different agent implementations can be driven by the same execution strategy
// without duplicating iteration, cancellation, or error-handling concerns.
package runner

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrNoFinalAnswer is returned by Run when the context is cancelled before
	// the agent produces a final answer. Callers can use errors.Is to
	// distinguish this from real planning or action failures, similar to how
	// http.ErrServerClosed signals a clean shutdown versus an unexpected error.
	ErrNoFinalAnswer = errors.New("no final answer")
)

// Agent is the contract the runner relies on to stay decoupled from any specific
// LLM backend or tool set. Keeping Plan, Act, and Answer as separate steps lets
// the runner interleave cancellation checks between them and makes each stage
// independently testable.
type Agent interface {
	Act(ctx context.Context) error
	Plan(ctx context.Context) error
	Task(ctx context.Context, prompt string) error
	Answer(ctx context.Context) (string, bool)
}

// Runner owns the loop so that agents don't need to implement their own
// iteration or cancellation logic.
type Runner struct {
	agent Agent
}

// New wires an agent to a runner without immediately starting work, allowing
// callers to configure context (deadlines, cancellation) before committing to a run.
func New(agent Agent) *Runner {
	return &Runner{agent: agent}
}

// Run drives the Plan→Answer?→Act cycle, checking context cancellation on every
// iteration so long-running tool calls don't silently outlive a deadline.
func (r *Runner) Run(ctx context.Context) (string, error) {
RUN:
	for {
		select {
		case <-ctx.Done():
			break RUN
		default:
			if err := r.agent.Plan(ctx); err != nil {
				return "", fmt.Errorf("planning failed: %w", err)
			}

			if answer, ok := r.agent.Answer(ctx); ok {
				return answer, nil
			}

			if err := r.agent.Act(ctx); err != nil {
				return "", fmt.Errorf("action failed: %w", err)
			}
		}
	}
	return "", fmt.Errorf("run cancelled: %w", ErrNoFinalAnswer)
}
