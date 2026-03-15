package agents

import (
	"context"

	"github.com/bit8bytes/beago/inputs/roles"
	"github.com/bit8bytes/beago/llms"
)

func (a *Agent) addAssistantMessage(ctx context.Context, content string) error {
	return a.history.Add(ctx, llms.Message{
		Role:    roles.Assistant,
		Content: content,
	})
}

func (a *Agent) addObservationMessage(ctx context.Context, observation string) error {
	return a.history.Add(ctx, llms.Message{
		Role:    roles.System,
		Content: "Observation: " + observation,
	})
}
