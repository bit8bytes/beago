package react

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/pipe"
	"github.com/bit8bytes/beago/tools"
)

type response struct {
	Thought     string          `json:"thought"`
	Action      string          `json:"action"`
	ActionInput json.RawMessage `json:"action_input"`
	FinalAnswer string          `json:"final_answer"`
}

// Instructions injects the ReAct system prompt and tool descriptions into the stream
// as JSON-encoded llm.Message objects (system + user).
func Instructions(ts ...tools.Tool) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var sb strings.Builder
		fmt.Fprintln(&sb, "You are a ReAct agent. Solve tasks step by step using the available tools.")
		fmt.Fprintln(&sb, "Do not estimate or predict values. Use only values returned by tools.")
		fmt.Fprintln(&sb, "\nAvailable tools:")
		for _, t := range ts {
			fmt.Fprintf(&sb, "\n- %s: %s\n", t.Name, t.Description)
			for _, p := range t.Params {
				req := "optional"
				if p.Required {
					req = "required"
				}
				fmt.Fprintf(&sb, "    - %s (%s): %s\n", p.Name, req, p.Description)
			}
		}
		fmt.Fprintln(&sb, `
STRICT OUTPUT RULES — you MUST follow these on every single turn:
- Output ONLY a raw JSON object. No markdown, no code fences, no prose before or after.
- Every response must be valid JSON with exactly these fields:
  {"thought": "...", "action": "...", "action_input": {...}, "final_answer": "..."}
- To call a tool: set "action" to the tool name, "action_input" to its params, "final_answer" to "".
- To give the final answer: set "action" to "", "action_input" to {}, "final_answer" to your answer.
- Never output plain text. Never wrap JSON in backticks or markdown code blocks.
- Always read the file first, then write to it.
- NEVER use write_file to fix errors. write_file is ONLY for creating new files.
- NEVER guess or read the file to find errors. ALWAYS run go build first to get the exact line number.
- To fix a compiler error: (1) run go build → get line N, (2) run cat to read the file and see what line N contains, (3) determine the correct content for that line, (4) run: sed -i '' 'Ns/.*/REPLACEMENT/' file.go
  Example: go build says main.go:6 error, cat shows line 6 is '"strconv' with missing closing quote → sed -i '' '6s/.*/"strconv"/' main.go
  Use .* to replace the whole line — never try to match the broken content.

Think step by step. Do not hallucinate.`)

		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(llm.Message{Role: "system", Content: sb.String()}); err != nil {
			return err
		}
		return enc.Encode(llm.Message{Role: "user", Content: strings.TrimSpace(string(input))})
	})
}

// Done returns a Handler that signals loop termination when the stream contains a non-empty final_answer.
func Done() pipe.Handler {
	return pipe.Exit(func(b []byte) bool {
		var resp response
		json.Unmarshal(b, &resp)
		return resp.FinalAnswer != ""
	})
}

// ParseAction validates that the stream contains a well-formed ReAct response and passes it through.
func ParseAction() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var resp response
		if err := json.NewDecoder(r).Decode(&resp); err != nil {
			fmt.Fprintf(w, "\nObservation: invalid ReAct response: %v\n", err)
			return nil
		}
		if resp.Action == "" && resp.FinalAnswer == "" {
			fmt.Fprintf(w, "\nObservation: response must set either action or final_answer\n")
			return nil
		}
		return json.NewEncoder(w).Encode(resp)
	})
}
