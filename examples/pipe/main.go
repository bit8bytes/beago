package main

import (
	"context"
	"fmt"
	"log"
	"os"

	llm "github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// Usage: echo "What is 2+2?" | go run .
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx := context.Background()

	err := pipe.Execute(ctx, os.Stdin, os.Stdout,
		llm.Generate(model),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
