package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type HelloWorldTool struct{}

func (t *HelloWorldTool) Name() string {
	return "helloWorld"
}

func (t *HelloWorldTool) Description() string {
	return "Returns a hello greeting for the given name."
}

func (t *HelloWorldTool) Parameters() []Parameter {
	return []Parameter{
		{Name: "name", Description: "The name to greet", Required: true},
	}
}

func (t *HelloWorldTool) Execute(_ context.Context, params json.RawMessage) (string, error) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return "", fmt.Errorf("helloWorld: invalid params: %w", err)
	}
	return "Hello, " + input.Name + "!", nil
}
