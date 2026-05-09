package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	llm "github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// Usage: echo "Count down from 3 to 1, one number per line. When done write DONE." | go run .
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := pipe.Execute(ctx, os.Stdin, os.Stdout,
		pipe.Loop(
			llm.Generate(model),
			pipe.Exit(func(b []byte) bool {
				return bytes.Contains(b, []byte("DONE"))
			}),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
