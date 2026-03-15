// Package agents provides a ReAct loop for LLM-powered agents.
//
// LLMs alone can't take actions or recover from mistakes mid-task. The ReAct
// pattern (Reasoning + Acting) solves this by interleaving LLM reasoning steps
// with tool executions, so the model can observe results and adapt before
// committing to a final answer. This package wires that loop together.
package agents

import (
	"context"
	"fmt"

	"github.com/bit8bytes/beago/inputs/roles"
	"github.com/bit8bytes/beago/llms"
	"github.com/bit8bytes/beago/tools"
)

// response is the wire format the LLM produces each iteration.
// It maps directly to the JSON schema described in the system prompt.
// FinalAnswer being non-empty signals the agent is done; otherwise Action and
// ActionInput describe the next tool to call.
type response struct {
	Thought     string `json:"thought"`
	Action      string `json:"action"`
	ActionInput string `json:"action_input"`
	FinalAnswer string `json:"final_answer"`
}

// Action is the domain representation of a tool call, extracted from a
// response. Name matches the tool's Name() and Input is passed verbatim to
// Tool.Execute.
type Action struct {
	Name  string
	Input string
}

type llm interface {
	Generate(ctx context.Context, messages []llms.Message) (string, error)
}

type store interface {
	Add(ctx context.Context, msgs ...llms.Message) error
	List(ctx context.Context) ([]llms.Message, error)
	Clear(ctx context.Context) error
}

type parser interface {
	Parse(text string) (response, error)
	Instructions() string
}

// Agent executes tasks using the ReAct pattern (reasoning + acting).
// Call Plan to generate the next action, then Act to execute it.
// Repeat until Plan returns Finish=true, then retrieve the result with Answer.
type Agent struct {
	model   llm
	tools   map[string]tools.Tool
	history store
	actions []Action
	parser  parser
	answer  string
}

// New creates an agent with the given model, tools, storage, and parser.
// For the ReAct pattern, prefer NewReAct.
func New(model llm, tools []tools.Tool, storage store, p parser) *Agent {
	return &Agent{
		model:   model,
		tools:   toolNames(tools),
		history: storage,
		parser:  p,
	}
}

// Task sets the user's question or task for the agent to solve.
// Call this before starting the Plan-Act loop.
func (a *Agent) Task(ctx context.Context, prompt string) error {
	return a.history.Add(ctx, llms.Message{
		Role:    roles.User,
		Content: "Question: " + prompt,
	})
}

// Plan calls the LLM to decide the next action or provide a final answer.
// Returns Response.Finish=true when the task is complete.
func (a *Agent) Plan(ctx context.Context) error {
	history, err := a.history.List(ctx)
	if err != nil {
		return err
	}

	generated, err := a.model.Generate(ctx, history)
	if err != nil {
		return err
	}

	parsed, err := a.parser.Parse(generated)
	if err != nil {
		return fmt.Errorf("failed to parse agent response: %w", err)
	}

	if err := a.addAssistantMessage(ctx, generated); err != nil {
		return fmt.Errorf("failed to store assistant message: %w", err)
	}

	if parsed.FinalAnswer != "" {
		a.answer = parsed.FinalAnswer
		return nil
	}

	action := Action{
		Name:  parsed.Action,
		Input: parsed.ActionInput,
	}
	a.actions = []Action{action}

	return nil
}

func (a *Agent) Answer(ctx context.Context) (string, bool) {
	return a.answer, a.answer != ""
}

// Act executes the tool chosen by Plan and adds the result as an observation.
// Always call this after Plan (unless Plan returned Finish=true).
func (a *Agent) Act(ctx context.Context) error {
	for _, action := range a.actions {
		if err := a.handleAction(ctx, action); err != nil {
			return err
		}
	}
	a.clearActions()
	return nil
}

func (a *Agent) handleAction(ctx context.Context, action Action) error {
	t, exists := a.tools[action.Name]
	if !exists {
		return a.addObservationMessage(ctx, "The action "+action.Name+" doesn't exist.")
	}

	observation, err := t.Execute(ctx, action.Input)
	if err != nil {
		return a.addObservationMessage(ctx, "Error: "+err.Error())
	}

	return a.addObservationMessage(ctx, observation)
}

func (a *Agent) clearActions() {
	a.actions = nil
}

func toolNames(tls []tools.Tool) map[string]tools.Tool {
	t := make(map[string]tools.Tool, len(tls))
	for _, tool := range tls {
		t[tool.Name()] = tool
	}
	return t
}
