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

// Usage: echo '[{"role":"user","content":"What is 2+2?"}]' | go run .
func main() {
	client := ollama.New(ollama.WithModel("gemma4:e4b"))
	if err := pipe.Do(context.Background(), os.Stdin, os.Stdout,
		llm.Generate(client),
	); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
