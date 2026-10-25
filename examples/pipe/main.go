package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// Usage: echo "What is 2+2?" | go run .
func main() {
	client := ollama.New(ollama.WithModel("gemma4:e4b"))
	pipe.Do(context.Background(), os.Stdin, os.Stdout,
		llm.Generate(client),
	)
	fmt.Fprintln(os.Stdout)
}
