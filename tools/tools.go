// Package tools defines the Tool interface and ships built-in tool
// implementations that agents can use to interact with the outside world.
package tools

import (
	"context"
	"encoding/json"
)

// Parameter describes a single input field a tool accepts.
type Parameter struct {
	Name        string
	Description string
	Required    bool
}

// Tool represents an action the agent can perform.
// Each tool must provide a name, description, parameter schema, and execution logic.
type Tool interface {
	Name() string
	Description() string
	Parameters() []Parameter
	Execute(ctx context.Context, params json.RawMessage) (string, error)
}
