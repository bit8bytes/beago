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

const systemPrompt = `
You are a helpful assistant. Respond with a single JSON object. No text outside the JSON.

Available tools:
- reverse(message): Reverse the characters of a message.

To call a tool:
{"action": "reverse", "action_input": {"message": "<text>"}, "final_answer": ""}

When you have the final answer:
{"action": "", "action_input": {}, "final_answer": "<your answer>"}`

type action struct {
	Action      string          `json:"action"`
	ActionInput json.RawMessage `json:"action_input"`
	FinalAnswer string          `json:"final_answer"`
}

// Usage: echo "Reverse the word: hello" | go run .
func main() {
	model := ollama.New("gemma3:4b", "")

	if err := pipe.Do(context.Background(), os.Stdin, os.Stdout,
		pipe.HandlerFunc(setupSystemPromptAndUserQuestion),
		pipe.Loop(
			llm.Generate(model),
			pipe.HandlerFunc(extractJSON),
			pipe.HandlerFunc(executeTool),
			pipe.HandlerFunc(exitOnFinalAnswer),
		),
	); err != nil {
		log.Fatal(err)
	}
}

func setupSystemPromptAndUserQuestion(ctx context.Context, r io.Reader, w io.Writer) error {
	input, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(llm.Message{Role: "system", Content: systemPrompt}); err != nil {
		return err
	}
	return enc.Encode(llm.Message{Role: "user", Content: strings.TrimSpace(string(input))})
}

func extractJSON(ctx context.Context, r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	start, depth := -1, 0
	for i, b := range data {
		switch b {
		case '{':
			if start == -1 {
				start = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && start != -1 {
				_, err = w.Write(data[start : i+1])
				return err
			}
		}
	}
	return fmt.Errorf("no JSON object found in response")
}

func executeTool(ctx context.Context, r io.Reader, w io.Writer) error {
	input, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	var a action
	if err := json.Unmarshal(input, &a); err != nil {
		return err
	}
	if a.FinalAnswer != "" {
		_, err = w.Write(input)
		return err
	}
	switch a.Action {
	case "reverse":
		var p struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(a.ActionInput, &p); err != nil {
			return err
		}
		runes := []rune(p.Message)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		out, err := json.Marshal(action{FinalAnswer: string(runes)})
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	default:
		return fmt.Errorf("unknown tool: %s", a.Action)
	}
}

func exitOnFinalAnswer(ctx context.Context, r io.Reader, w io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	var v struct {
		FinalAnswer string `json:"final_answer"`
	}
	_ = json.Unmarshal(data, &v)
	if v.FinalAnswer != "" {
		return pipe.ErrDone
	}
	return nil
}
