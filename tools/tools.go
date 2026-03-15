// Package tools defines the Tool interface and ships built-in tool
// implementations that agents can use to interact with the outside world.
package tools

import "context"

// Tool represents an action the agent can perform.
// Each tool must provide a name, description, and execution logic.
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
}
