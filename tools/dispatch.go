package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bit8bytes/beago/pipe"
)

type action struct {
	Action      string                     `json:"action"`
	ActionInput map[string]json.RawMessage `json:"action_input"`
	FinalAnswer string                     `json:"final_answer"`
}

func coerceArgs(input map[string]json.RawMessage) map[string]string {
	args := make(map[string]string, len(input))
	for k, v := range input {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			args[k] = s
			continue
		}
		var ss []string
		if err := json.Unmarshal(v, &ss); err == nil {
			args[k] = strings.Join(ss, " ")
			continue
		}
		args[k] = strings.Trim(string(v), `"[] `)
	}
	return args
}

// Execute reads an action JSON blob, runs the matched tool, and writes the observation.
// A final_answer passes through unchanged for pipe.Exit to catch.
func Execute(tools ...Tool) pipe.Handler {
	registry := make(map[string]Tool, len(tools))
	for _, t := range tools {
		registry[t.Name] = t
	}
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var a action
		if err := json.Unmarshal(data, &a); err != nil {
			fmt.Fprintf(w, "\nObservation: invalid action JSON: %v\n", err)
			return nil
		}
		if a.FinalAnswer != "" {
			_, err = w.Write(data)
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
		t, ok := registry[a.Action]
		if !ok {
			fmt.Fprintf(w, "\nObservation: unknown tool %q\n", a.Action)
			return nil
		}
		result, err := t.Run(ctx, coerceArgs(a.ActionInput))
		if err != nil {
			if result != "" {
				fmt.Fprintf(w, "\nObservation: %s\nerror: %v\n", result, err)
			} else {
				fmt.Fprintf(w, "\nObservation: %v\n", err)
			}
			return nil
		}
		fmt.Fprintf(w, "\nObservation: %s\n", result)
		return nil
	})
}
