package main

import (
	"context"
	"fmt"
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

// Usage: echo "Write a function that adds two integers" | go run .
// Run from the root of the Go module you want the agent to work on.
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	readFile := tools.ReadFile()
	editFile := tools.EditFile()
	writeFile := tools.WriteFile()
	goBuild := tools.GoBuild()
	goTest := tools.GoTest()
	shell := tools.Exec("ls", "find", "grep")

	hist := history.New()

	err := pipe.Execute(ctx, os.Stdin, os.Stdout,
		react.CodingInstructions(readFile, editFile, writeFile, goBuild, goTest, shell),
		hist.System(),
		pipe.Loop(
			hist.Replay(),
			llm.Generate(model),
			jsonpkg.Extract(),
			hist.Assistant(),
			jsonpkg.Pretty(os.Stderr),
			react.ParseAction(),
			tools.Execute(readFile, editFile, writeFile, goBuild, goTest, shell),
			hist.User(),
			react.Done(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}
