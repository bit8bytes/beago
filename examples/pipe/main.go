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

// Usage: echo "What is 2+2?" | go run .
func main() {
	client := ollama.New(ollama.WithModel("gemma4:e4b"))

	ctx := context.Background()

	err := pipe.Do(ctx, os.Stdin, os.Stdout,
		llm.Generate(client),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
