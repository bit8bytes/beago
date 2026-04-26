package main

import (
	"context"
	"log"
	"os"
	"time"

	react "github.com/bit8bytes/beago/agents"
	"github.com/bit8bytes/beago/history"
	jsonpkg "github.com/bit8bytes/beago/json"
	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
	"github.com/bit8bytes/beago/tools"
)

// Usage: echo "Write a README for the pipe package based on its source files" | go run .
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	shell := tools.Exec("man", "ls", "grep", "cat", "find", "sed", "go")
	write := tools.WriteFile()

	hist := history.New()

	err := pipe.Execute(ctx, os.Stdin, os.Stdout,
		react.Instructions(shell, write),
		pipe.Loop(20,
			hist.Accumulate(),
			llm.Generate(model),
			jsonpkg.Extract(),
			pipe.Tee(os.Stderr),
			react.ParseAction(),
			tools.Execute(shell, write),
			react.Done(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
}
