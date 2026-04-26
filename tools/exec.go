package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Exec creates a Tool that runs shell commands restricted to the given allowlist.
func Exec(allowed ...string) Tool {
	set := make(map[string]bool, len(allowed))
	for _, cmd := range allowed {
		set[cmd] = true
	}
	return Tool{
		Name:        "shell",
		Description: "Run a shell command",
		Params: []Param{
			{Name: "cmd", Description: "command to run, e.g. ls -la /tmp", Required: true},
		},
		Run: func(ctx context.Context, args map[string]string) (string, error) {
			parts := strings.Fields(args["cmd"])
			if len(parts) == 0 {
				return "", nil
			}
			if len(set) > 0 && !set[parts[0]] {
				return "", fmt.Errorf("command %q is not allowed", parts[0])
			}
			out, err := exec.CommandContext(ctx, parts[0], parts[1:]...).CombinedOutput()
			return string(out), err
		},
	}
}
