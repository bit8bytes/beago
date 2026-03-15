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

	// These tools are specifically designed for Golang.
	tools := []tools.Tool{
		&tools.HelloWorldTool{},
	}

	agent, err := agents.NewReAct(ctx, model, tools, storage)
	if err != nil {
		panic(err)
	}

	task := `Use the tool helloWorld with the name Beago as input`
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
