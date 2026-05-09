package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// Usage: echo "What is the capital of France?" | go run .
//
// Tee splits the stream like a Unix tee(1): the LLM response flows to stdout
// while a copy is written to stderr for inspection.
// This is especially useful for debugging streaming handlers,
// as it allows you to see the output without interrupting the flow of the pipeline.
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx := context.Background()

	err := pipe.Execute(ctx, os.Stdin, os.Stdout,
		llm.Generate(model),
		pipe.Tee(os.Stderr),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
