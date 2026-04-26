package tools

import (
	"context"
	"os/exec"
	"strings"
)

// Shell creates a Tool that runs a system binary.
// Args are built from params in definition order, splitting each value on spaces.
// e.g. params [{flags}, {path}] with args {"flags":"-la","path":"/tmp"} runs: name -la /tmp
func Shell(name, description string, params []Param) Tool {
	return Tool{
		Name:        name,
		Description: description,
		Params:      params,
		Run: func(ctx context.Context, args map[string]string) (string, error) {
			var cmdArgs []string
			for _, p := range params {
				if v := args[p.Name]; v != "" {
					cmdArgs = append(cmdArgs, strings.Fields(v)...)
				}
			}
			out, err := exec.CommandContext(ctx, name, cmdArgs...).CombinedOutput()
			return string(out), err
		},
	}
}
