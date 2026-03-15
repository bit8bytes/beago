package tools

import "context"

type HelloWorldTool struct{}

func (t *HelloWorldTool) Name() string {
	return "helloWorld"
}

func (t *HelloWorldTool) Description() string {
	return "Returns a hello greeting for the given name. Input: a name string."
}

func (t *HelloWorldTool) Execute(_ context.Context, input string) (string, error) {
	return "Hello, " + input + "!", nil
}
