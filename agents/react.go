package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/bit8bytes/beago/inputs/roles"
	"github.com/bit8bytes/beago/llms"
	"github.com/bit8bytes/beago/outputs/json"
	"github.com/bit8bytes/beago/tools"
)

// NewReAct creates an agent pre-configured for the ReAct pattern.
// It seeds the ReAct system prompt into storage.
func NewReAct(ctx context.Context, model llm, tls []tools.Tool, storage store) (*Agent, error) {
	p := json.NewParser[response]()
	t := toolNames(tls)

	msgs := buildReActPrompt(t, p.Instructions())
	if err := storage.Add(ctx, msgs...); err != nil {
		return nil, err
	}

	return &Agent{
		model:   model,
		tools:   t,
		history: storage,
		parser:  p,
	}, nil
}

func buildReActPrompt(tls map[string]tools.Tool, jsonInstructions string) []llms.Message {
	var toolDescriptions strings.Builder
	for _, t := range tls {
		fmt.Fprintf(&toolDescriptions, "- %s: %s\n", t.Name(), t.Description())
		for _, p := range t.Parameters() {
			req := "optional"
			if p.Required {
				req = "required"
			}
			fmt.Fprintf(&toolDescriptions, "    - %s (%s): %s\n", p.Name, req, p.Description)
		}
	}

	return []llms.Message{
		{
			Role: roles.System,
			Content: fmt.Sprintf(`
You are an helpful agent. Answer questions using the available tools.
Do not estimate or predict values. Use only values returned by tools.

Available tools:
%s
%s

Respond with a JSON object on each turn with these fields:
- "thought": your reasoning about what to do next
- "action": the exact tool name to call (empty string when giving final answer)
- "action_input": a JSON object whose keys are the tool's parameter names (empty object {} when giving final answer)
- "final_answer": your final answer to the user — MUST be non-empty when you are done; empty string ONLY when calling a tool

When you have enough information to answer, set "action" to "" and "action_input" to {} and put a detailed answer based on your observations — MUST be non-empty when done; be thorough and include all relevant findings.

Think step by step. Do not hallucinate.`, toolDescriptions.String(), jsonInstructions),
		},
	}
}
