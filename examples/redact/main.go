package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
	"github.com/bit8bytes/beago/redact"
)

// Usage: echo 'Alice sent Bob $500 last Tuesday.' | go run .
func main() {
	mask := redact.Mask{}
	// Adding values to the mask allows them to be anonymized and rehydrated later.
	// Anonymize must be called before [llm.Generate] to ensure that the LLM does
	// not see the sensitive values.
	mask.Add("Alice", "PERSON")
	mask.Add("Bob", "PERSON")
	mask.Add("$500", "AMOUNT")

	client := ollama.New(ollama.WithModel("gemma4:e4b"))

	pipe.Do(context.Background(), os.Stdin, os.Stdout,
		llm.Prompt("Summarize the following in one sentence."),
		mask.Anonymize(),
		// The LLM doesn't see the values that have been masked by the [redact.Mask] handler,
		// but it can still generate a summary based on the masked text.
		llm.Generate(client),
		mask.Rehydrate(),
	)
	fmt.Fprintln(os.Stdout)
}
