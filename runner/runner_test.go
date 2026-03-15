package runner

import (
	"context"
	"errors"
	"testing"
)

type mockAgent struct {
	planErr     error
	actErr      error
	answer      string
	planCalls   int
	actCalls    int
	answerAfter int // how many Plan calls before Answer returns true
}

func (m *mockAgent) Plan(_ context.Context) error {
	m.planCalls++
	return m.planErr
}

func (m *mockAgent) Act(_ context.Context) error {
	m.actCalls++
	return m.actErr
}

func (m *mockAgent) Task(_ context.Context, _ string) error { return nil }

func (m *mockAgent) Answer(_ context.Context) (string, bool) {
	if m.answer != "" && m.planCalls >= m.answerAfter {
		return m.answer, true
	}
	return "", false
}

func TestRun_ReturnsAnswerImmediately(t *testing.T) {
	agent := &mockAgent{answer: "done", answerAfter: 1}
	got, err := New(agent).Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "done" {
		t.Errorf("got %q, want %q", got, "done")
	}
	if agent.actCalls != 0 {
		t.Errorf("Act should not be called when Answer is ready, got %d calls", agent.actCalls)
	}
}

func TestRun_ActsUntilAnswerReady(t *testing.T) {
	agent := &mockAgent{answer: "ready", answerAfter: 3}
	got, err := New(agent).Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ready" {
		t.Errorf("got %q, want %q", got, "ready")
	}
	if agent.planCalls != 3 {
		t.Errorf("expected 3 Plan calls, got %d", agent.planCalls)
	}
	if agent.actCalls != 2 {
		t.Errorf("expected 2 Act calls, got %d", agent.actCalls)
	}
}

func TestRun_PlanError(t *testing.T) {
	planErr := errors.New("plan boom")
	_, err := New(&mockAgent{planErr: planErr}).Run(context.Background())
	if !errors.Is(err, planErr) {
		t.Errorf("expected planErr in chain, got: %v", err)
	}
}

func TestRun_ActError(t *testing.T) {
	actErr := errors.New("act boom")
	_, err := New(&mockAgent{actErr: actErr, answerAfter: 999}).Run(context.Background())
	if !errors.Is(err, actErr) {
		t.Errorf("expected actErr in chain, got: %v", err)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(&mockAgent{answerAfter: 999}).Run(ctx)
	if !errors.Is(err, ErrNoFinalAnswer) {
		t.Errorf("expected ErrNoFinalAnswer, got: %v", err)
	}
}
