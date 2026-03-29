package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bit8bytes/beago/agents"
	"github.com/bit8bytes/beago/llms/ollama"
	"github.com/bit8bytes/beago/runner"
	"github.com/bit8bytes/beago/stores/memory"
	"github.com/bit8bytes/beago/tools"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	storage := memory.New()
	defer storage.Close()

	model := ollama.New(ollama.Model{
		Model:   "gemma3:12b",
		Options: ollama.Options{NumCtx: 4096},
		Stream:  false,
		Format:  ollama.JSON,
	})

	tools := []tools.Tool{&tools.Search{}}

	agent, err := agents.NewReAct(ctx, model, tools, storage)
	if err != nil {
		panic(err)
	}

	// The llm will use the search tool to fetch the content of the page and extract the arguments if any.
	// The search tool is currently just for testing and needs to be remodeled to be production ready.
	task := "Fetch https://httpbin.org/get?q=lorem+ipsum and tell me what arguments the response contains."
	if err := agent.Task(ctx, task); err != nil {
		panic(err)
	}

	r := runner.New(agent)

	res, err := r.Run(ctx)
	if err != nil {
		switch {
		case errors.Is(err, runner.ErrNoFinalAnswer):
			fmt.Println("No final answer found")
		default:
			panic(err)
		}
		return
	}

	fmt.Println(res)
}
