package main

import (
	"context"
	"fmt"

	"github.com/bit8bytes/beago/inputs/chats"
	"github.com/bit8bytes/beago/inputs/roles"
	"github.com/bit8bytes/beago/llms"
	"github.com/bit8bytes/beago/llms/ollama"
)

func main() {
	messages := []llms.Message{
		{
			Role:    roles.System,
			Content: "Translate {{.InputLanguage}} to {{.OutputLanguage}}.",
		},
		{
			Role:    roles.System,
			Content: "Return only the concrete translation.",
		},
		{
			Role:    roles.User,
			Content: "{{.Text}}",
		},
	}

	type chatData struct {
		InputLanguage  string
		OutputLanguage string
		Text           string
	}

	data := chatData{
		InputLanguage:  "English",
		OutputLanguage: "French",
		Text:           "I love programming.",
	}

	formattedMessages, err := chats.New(messages).Execute(data)
	if err != nil {
		panic(err)
	}

	client := ollama.New(ollama.Model{
		Model:   "gemma3n:e2b",
		Options: ollama.Options{NumCtx: 4096},
		Stream:  false,
	})
	out, _ := client.Generate(context.Background(), formattedMessages)
	fmt.Println(out)
}
