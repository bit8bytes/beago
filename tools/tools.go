package tools

import "context"

// Param describes a single input a tool accepts.
type Param struct {
	Name        string
	Description string
	Required    bool
}

// Tool defines an action the agent can perform.
type Tool struct {
	Name        string
	Description string
	Params      []Param
	Run         func(ctx context.Context, args map[string]string) (string, error)
}
