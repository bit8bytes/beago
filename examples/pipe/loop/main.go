package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	llm "github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// Usage: echo "Count down from 3 to 1, one number per line. When done write DONE." | go run .
func main() {
	model := ollama.New("gemma4:e4b", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := pipe.Do(ctx, os.Stdin, os.Stdout,
		pipe.Loop(
			llm.Generate(model),
			pipe.HandlerFunc(exitOnDone),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout)
}

func exitOnDone(ctx context.Context, r io.Reader, w io.Writer) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("exitOnDone: %v", err)
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("exitOnDone: %v", err)
	}
	if bytes.Contains(b, []byte("DONE")) {
		return pipe.ErrDone
	}
	return nil
}
