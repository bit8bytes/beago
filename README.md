# beago: A Go framework for building LLM-powered applications.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) ![Test](https://github.com/bit8bytes/beago/actions/workflows/tests.yml/badge.svg) ![Sec Scan](https://github.com/bit8bytes/beago/actions/workflows/sec_scan.yml/badge.svg)

beago provides composable building blocks for LLM-powered Go applications — pipes for structured output, agents for tool-using reasoning loops, and stores for conversation history. The core library has no external dependencies.

## Core Concepts

- **Pipes** — simple `Input → LLM → Output` pipelines with typed, structured responses
- **Agents** — ReAct (Reasoning + Acting) loops that interleave LLM reasoning with tool execution
- **Stores** — tamper-evident message history with a SHA-256 hash chain, keeping LLMs stateful across turns
- **Tools** — implement the `Tool` interface to give agents new capabilities

## Quick Start

```go
// Pipe: send messages and get structured output
pipe := pipes.New(messages, model, parser)
result, _ := pipe.Invoke(ctx)

// Agent: reason and act with tools
agent, _ := agents.NewReAct(ctx, model, tools, storage)
agent.Task(ctx, "Use the helloWorld tool with name Beago")
res, _ := runner.New(agent).Run(ctx)
```

See [Example](/examples/agents/hello/main.go) for full working examples.

## Contributions

Contributions of any kind are welcome! See [Get Involved](/docs/GET-INVOLVED.md) to get started.
