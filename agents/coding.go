package react

import (
	"context"
	"fmt"
	"io"

	"github.com/bit8bytes/beago/pipe"
	"github.com/bit8bytes/beago/tools"
)

// CodingInstructions injects a coding-agent system prompt and tool descriptions into the stream.
// The workflow it teaches: go_build → read_file → edit_file → go_build → go_test.
func CodingInstructions(ts ...tools.Tool) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		fmt.Fprintln(w, "You are a coding agent. Fix bugs, implement features, and write tests in Go.")
		fmt.Fprintln(w, "Work in the current directory. Make small, targeted edits. Never guess at file contents — always read first.")
		fmt.Fprintln(w, "\nAvailable tools:")
		for _, t := range ts {
			fmt.Fprintf(w, "\n- %s: %s\n", t.Name, t.Description)
			for _, p := range t.Params {
				req := "optional"
				if p.Required {
					req = "required"
				}
				fmt.Fprintf(w, "    - %s (%s): %s\n", p.Name, req, p.Description)
			}
		}
		fmt.Fprintln(w, `
Workflow:
1. go_build → see compile errors with file:line numbers
2. read_file on the failing file with a narrow line range around the error
3. edit_file with the exact old_string and corrected new_string
4. go_build again → confirm it compiles
5. go_test → confirm tests pass
6. final_answer with a summary of all changes made

STRICT OUTPUT RULES — follow on every single turn:
- Output ONLY a raw JSON object. No markdown, no code fences, no prose before or after.
- Every response must be valid JSON with exactly these fields:
  {"thought": "...", "action": "...", "action_input": {...}, "final_answer": ""}
- To call a tool: set "action" to the tool name, "action_input" to its params, "final_answer" to "".
- To finish: set "action" to "", "action_input" to {}, "final_answer" to your summary of changes.
- Never wrap JSON in backticks or markdown code blocks.
- Never guess at file content — always read_file before edit_file.
- Use edit_file for changes, write_file only for creating new files.
- When go_build succeeds (empty output), proceed to go_test.

Think step by step. Do not hallucinate.`)
		_, err := io.Copy(w, r)
		return err
	})
}
