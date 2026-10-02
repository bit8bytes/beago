# beago: A Go framework for building LLM-powered applications.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) ![Test](https://github.com/bit8bytes/beago/actions/workflows/tests.yml/badge.svg) ![Sec Scan](https://github.com/bit8bytes/beago/actions/workflows/sec_scan.yml/badge.svg)

beago brings the Unix philosophy to LLM applications: small, focused handlers connected by pipes. It is also inspired by the http.Handler and http.Handler funcs (net/http). Each handler reads from an `io.Reader`, transforms the stream, and writes to an `io.Writer`.

## Core Concepts

- **Pipe** — the core primitive: a `Handler` that reads `io.Reader` → transforms → writes `io.Writer`
- **Execute** — chains handlers sequentially, connecting each output to the next input via `io.Pipe`
- **Loop** — runs a handler chain repeatedly, feeding each iteration's output as the next input; stops on `ErrDone` or a max iteration count

## Quick Start

```go
// echo "What is 2+2?" | go run .
model := ollama.New("gemma4:e4b", "")
pipe.Execute(context.Background(), os.Stdin, os.Stdout,
    llm.Generate(model),
)
```

More examples in [/examples](/examples).

## Contributions

Contributions of any kind are welcome! See [Get Involved](/docs/GET-INVOLVED.md) to get started.
