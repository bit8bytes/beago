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

// Usage: echo "The quick brown fox jumps over the lazy dog." | go run .
//
// Two LLM calls are chained like Unix pipes: the first translates to French,
// the second summarises the French text into two sentences, still in French.
// Each handler reads from the previous handler's output — just like:
//
//	echo "..." | translate | summarise
func main() {
	client := ollama.New(ollama.WithModel("gemma4:e4b"))

	ctx := context.Background()

	if err := pipe.Do(ctx, os.Stdin, os.Stdout,
		llm.Prompt("Translate the following text to French. Output only the translation."),
		llm.Generate(client),
		llm.Prompt("Summarise the following French text in two sentences. Respond in French."),
		llm.Generate(client),
	); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
