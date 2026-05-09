// Package json provides a pipe.Handler for extracting JSON from LLM output.
package json

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bit8bytes/beago/pipe"
)

// Extract strips prose from the stream and emits the first JSON blob.
func Extract() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		blob := extractJSON(data)
		if blob == nil {
			fmt.Fprintf(w, "Generated JSON is malformed, try again.\n")
			return nil
		}
		_, err = w.Write(blob)
		return err
	})
}

// Pretty pretty-prints the JSON in the stream to w and passes the original bytes through.
func Pretty(w io.Writer) pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w2 io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if json.Indent(&buf, data, "", "  ") == nil {
			fmt.Fprintln(w, buf.String())
		}
		_, err = w2.Write(data)
		return err
	})
}

func extractJSON(data []byte) []byte {
	start := -1
	depth := 0
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
				return data[start : i+1]
			}
		}
	}
	return nil
}
