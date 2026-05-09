package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ReadFile creates a Tool that reads a file with optional line range.
// Output includes line numbers to help the model target edits precisely.
func ReadFile() Tool {
	return Tool{
		Name:        "read_file",
		Description: "Read a file with line numbers. Optionally restrict to a line range.",
		Params: []Param{
			{Name: "path", Description: "file path to read", Required: true},
			{Name: "start_line", Description: "first line to read, 1-indexed (optional)", Required: false},
			{Name: "end_line", Description: "last line to read, inclusive (optional)", Required: false},
		},
		Run: func(_ context.Context, args map[string]string) (string, error) {
			data, err := os.ReadFile(args["path"])
			if err != nil {
				return "", err
			}
			lines := strings.Split(string(data), "\n")
			start, end := 1, len(lines)
			if s := args["start_line"]; s != "" {
				if n, err := strconv.Atoi(s); err == nil && n >= 1 {
					start = n
				}
			}
			if e := args["end_line"]; e != "" {
				if n, err := strconv.Atoi(e); err == nil {
					end = n
				}
			}
			if end > len(lines) {
				end = len(lines)
			}
			var sb strings.Builder
			for i := start; i <= end; i++ {
				fmt.Fprintf(&sb, "%d\t%s\n", i, lines[i-1])
			}
			return sb.String(), nil
		},
	}
}

// EditFile creates a Tool that replaces the first occurrence of old_string with new_string in a file.
func EditFile() Tool {
	return Tool{
		Name:        "edit_file",
		Description: "Replace the first occurrence of old_string with new_string in a file.",
		Params: []Param{
			{Name: "path", Description: "file path to edit", Required: true},
			{Name: "old_string", Description: "exact string to find and replace", Required: true},
			{Name: "new_string", Description: "replacement string", Required: true},
		},
		Run: func(_ context.Context, args map[string]string) (string, error) {
			data, err := os.ReadFile(args["path"])
			if err != nil {
				return "", err
			}
			old := args["old_string"]
			if !strings.Contains(string(data), old) {
				return "", fmt.Errorf("old_string not found in %s", args["path"])
			}
			updated := strings.Replace(string(data), old, args["new_string"], 1)
			if err := os.WriteFile(args["path"], []byte(updated), 0644); err != nil {
				return "", err
			}
			return "edited " + args["path"], nil
		},
	}
}

// GoBuild creates a Tool that runs go build for a package.
func GoBuild() Tool {
	return Tool{
		Name:        "go_build",
		Description: "Build a Go package and report compile errors. Use ./... for all packages.",
		Params: []Param{
			{Name: "package", Description: "package pattern to build, e.g. ./... or ./cmd/foo (default: ./...)", Required: false},
		},
		Run: func(ctx context.Context, args map[string]string) (string, error) {
			pkg := args["package"]
			if pkg == "" {
				pkg = "./..."
			}
			out, err := exec.CommandContext(ctx, "go", "build", pkg).CombinedOutput()
			return string(out), err
		},
	}
}

// GoTest creates a Tool that runs go test for a package.
func GoTest() Tool {
	return Tool{
		Name:        "go_test",
		Description: "Run Go tests. Optionally filter by test name with the run param.",
		Params: []Param{
			{Name: "package", Description: "package pattern to test, e.g. ./... (default: ./...)", Required: false},
			{Name: "run", Description: "test name filter passed to -run flag (optional)", Required: false},
		},
		Run: func(ctx context.Context, args map[string]string) (string, error) {
			pkg := args["package"]
			if pkg == "" {
				pkg = "./..."
			}
			cmdArgs := []string{"test", pkg}
			if r := args["run"]; r != "" {
				cmdArgs = append(cmdArgs, "-run", r)
			}
			out, err := exec.CommandContext(ctx, "go", cmdArgs...).CombinedOutput()
			return string(out), err
		},
	}
}