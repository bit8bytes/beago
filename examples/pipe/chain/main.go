package main

import (
	"context"
	"log"
	"os"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// Usage: echo "The quick brown fox jumps over the lazy dog." | go run .
//
// Two LLM calls are chained like Unix pipes: the first translates to French,
// the second summarises the French text into one sentence.
// Each handler reads from the previous handler's output — just like:
//
//	echo "..." | translate | summarise
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx := context.Background()

	err := pipe.Execute(ctx, os.Stdin, os.Stdout,
		llm.Prompt("Translate the following text to French. Output only the translation."),
		llm.Generate(model),
		llm.Prompt("Summarise the following French text in two sentences."),
		llm.Generate(model),
	)
	if err != nil {
		log.Fatal(err)
	}
}
