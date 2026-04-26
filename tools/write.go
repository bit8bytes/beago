package tools

import (
	"context"
	"os"
)

// WriteFile creates a Tool that writes content to a file.
func WriteFile() Tool {
	return Tool{
		Name:        "write_file",
		Description: "Write text content to a file",
		Params: []Param{
			{Name: "path", Description: "file path to write to", Required: true},
			{Name: "content", Description: "content to write", Required: true},
		},
		Run: func(_ context.Context, args map[string]string) (string, error) {
			if err := os.WriteFile(args["path"], []byte(args["content"]), 0644); err != nil {
				return "", err
			}
			return "written to " + args["path"], nil
		},
	}
}
