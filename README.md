# beago: A Go framework for building LLM-powered applications.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) ![Test](https://github.com/bit8bytes/beago/actions/workflows/tests.yml/badge.svg) ![Sec Scan](https://github.com/bit8bytes/beago/actions/workflows/sec_scan.yml/badge.svg)

beago brings the Unix philosophy to LLM applications: small, focused handlers connected by pipes. Each handler reads from an `io.Reader`, transforms the stream, and writes to an `io.Writer` — exactly like Unix programs connected with `|`. The core library has no external dependencies.

## The Unix Pipe Model

Unix pipes let you compose small programs into powerful workflows:

```
echo "text" | translate | summarise | fmt
```

beago works the same way, but for LLM pipelines:

```go
pipe.Execute(ctx, os.Stdin, os.Stdout,
    llm.Prompt("Translate to French."),
    llm.Generate(model),              // stdin | translate
    llm.Prompt("Summarise in one sentence."),
    llm.Generate(model),              // | summarise
)
```

Each `pipe.Handler` is a composable unit. Handlers are chained with `pipe.Execute`, looped with `pipe.Loop`, and debugged with `pipe.Tee` — mirroring Unix's `tee(1)`.

## Core Concepts

- **Pipe** — the core primitive: a `Handler` that reads `io.Reader` → transforms → writes `io.Writer`
- **Execute** — chains handlers sequentially, connecting each output to the next input via `io.Pipe`
- **Loop** — runs a handler chain repeatedly, feeding each iteration's output as the next input; stops on `ErrDone` or a max iteration count
- **Tee** — splits the stream like Unix `tee(1)`: passes data through while copying to a second writer for debugging
- **Agents** — ReAct (Reasoning + Acting) loops that interleave LLM reasoning with tool execution
- **Tools** — implement the `Tool` interface to give agents new capabilities

## Quick Start

```go
// Single handler: pipe stdin through an LLM to stdout
// echo "What is 2+2?" | go run .
pipe.Execute(ctx, os.Stdin, os.Stdout,
    llm.Generate(model),
)
```

```go
// Chain handlers: translate then summarise
pipe.Execute(ctx, os.Stdin, os.Stdout,
    llm.Prompt("Translate to French."),
    llm.Generate(model),
    llm.Prompt("Summarise in one sentence."),
    llm.Generate(model),
)
```

```go
// Loop until the LLM outputs "DONE"
pipe.Execute(ctx, os.Stdin, os.Stdout,
    pipe.Loop(10,
        llm.Generate(model),
        pipe.Exit(func(b []byte) bool {
            return bytes.Contains(b, []byte("DONE"))
        }),
    ),
)
```

## Examples

| Example | Description |
|---|---|
| [pipe](/examples/pipe/main.go) | Single LLM call — the simplest pipe |
| [pipe/tee](/examples/pipe/tee/main.go) | Split the stream with `Tee` to inspect output |
| [pipe/chain](/examples/pipe/chain/main.go) | Chain two LLM calls: translate → summarise |
| [pipe/loop](/examples/pipe/loop/main.go) | Loop until a stop condition is met |
| [agents](/examples/agents/main.go) | ReAct agent with tools |

## Contributions

Contributions of any kind are welcome! See [Get Involved](/docs/GET-INVOLVED.md) to get started.
