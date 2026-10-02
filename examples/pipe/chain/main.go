package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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
		pipe.HandlerFunc(prompt("Translate the following text to French. Output only the translation.")),
		llm.Generate(model),
		pipe.HandlerFunc(prompt("Summarise the following French text in two sentences.")),
		llm.Generate(model),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}

func prompt(instruction string) func(ctx context.Context, r io.Reader, w io.Writer) error {
	return func(ctx context.Context, r io.Reader, w io.Writer) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(llm.Message{Role: "system", Content: instruction}); err != nil {
			return err
		}
		return enc.Encode(llm.Message{Role: "user", Content: strings.TrimSpace(string(input))})
	}
}
