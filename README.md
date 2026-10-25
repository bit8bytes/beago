# beago: A Go framework for building LLM-powered applications.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) ![Test](https://github.com/bit8bytes/beago/actions/workflows/tests.yml/badge.svg) ![Sec Scan](https://github.com/bit8bytes/beago/actions/workflows/sec_scan.yml/badge.svg)

beago brings the [Unix philosophy](https://cscie2x.dce.harvard.edu/hw/ch01s06.html) to LLM applications: small, focused handlers connected by pipes. It is also inspired by the http.Handler and http.HandlerFunc (net/http). Each handler reads from an `io.Reader`, transforms the stream, and writes to an `io.Writer`.

## Core Concepts

**Pipe** is the core primitive: a `Handler` that reads `io.Reader` → transforms → writes `io.Writer`. Two combinators build on it:

- **Do**: chains handlers sequentially, connecting each output to the next input via `io.Pipe`
- **Loop**: runs a handler chain repeatedly, feeding each iteration's output as the next input; stops on `ErrDone` via context cancellation

**Redact** is a separate, composable layer: a `Mask` that anonymizes personal data to tokens before LLM processing and rehydrates them back after keeping sensitive values out of the model entirely. Redact is used in combination with Pipe and is safe for concurrent use.

## Examples

### Quick Start

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
)

// echo '[{"role":"user","content":"What is 2+2?"}]' | go run .
func main() {
	client := ollama.New(ollama.WithModel("gemma4:e4b"))
	pipe.Do(context.Background(), os.Stdin, os.Stdout,
		llm.Generate(client),
	)
	fmt.Fprintln(os.Stdout)
}
```

This example reads a prompt from stdin, sends it to an Ollama model, and streams the response to stdout.

### Redact

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bit8bytes/beago/llm"
	"github.com/bit8bytes/beago/llm/ollama"
	"github.com/bit8bytes/beago/pipe"
	"github.com/bit8bytes/beago/redact"
)

// Usage: echo 'Alice sent Bob $500 last Tuesday.' | go run .
func main() {
	mask := redact.Mask{}
	// Adding values to the mask allows them to be anonymized and rehydrated later.
	// Anonymize must be called before [llm.Generate] to ensure that the LLM does
	// not see the sensitive values.
	mask.Add("Alice", "PERSON")
	mask.Add("Bob", "PERSON")
	mask.Add("$500", "AMOUNT")

	client := ollama.New(ollama.WithModel("gemma4:e4b"))

	pipe.Do(context.Background(), os.Stdin, os.Stdout,
		llm.Prompt("Summarize the following in one sentence."),
		mask.Anonymize(),
		// The LLM doesn't see the values that have been masked by the [redact.Mask] handler,
		// but it can still generate a summary based on the masked text.
		llm.Generate(client),
		mask.Rehydrate(),
	)
	fmt.Fprintln(os.Stdout)
}
```

This example masks named values (people, amounts) before the LLM sees the text, then restores the originals in the final output so sensitive data never leaves the application.

More examples in [/examples](/examples).

## Contributions

Contributions of any kind are welcome!
