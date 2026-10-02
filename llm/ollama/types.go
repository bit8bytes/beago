package ollama

import "github.com/bit8bytes/beago/llm"

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []llm.Message `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Model   string      `json:"model"`
	Message llm.Message `json:"message"`
	Done    bool        `json:"done"`
}
