# Pipe Package README

The `pipe` package is the central orchestration layer within the beago framework, designed for building complex, stateful, and streaming agent pipelines. It allows developers to connect various components (like LLM generation, history tracking, tool execution, and reactive logic) into a cohesive, step-by-step workflow.

## Purpose

Instead of writing traditional procedural code, the `pipe` package enables defining a workflow as a sequence of pipes. This structure ensures that the output of one component (the stream) becomes the input for the next, managing complex data flow and state transitions automatically.

## Key Concepts

1. **Streaming Execution:** Pipes operate on continuous streams of data, making them efficient for real-time or iterative processes.
2. **State Management:** The package handles the accumulation and passing of history and context across multiple steps (e.g., using `history.Accumulate()`).
3. **Component Integration:** It provides mechanisms to seamlessly integrate external services (LLMs via `llm.Generate()`) and predefined logic components (`react.Instructions()`, `jsonpkg.Extract()`).
4. **Looping and Iteration:** Complex agents can be built using `pipe.Loop()`, allowing the workflow to repeat and refine results until a defined stopping condition is met.

## Usage Example (Agent Pipeline)

As demonstrated in the main usage, setting up a pipeline involves defining the context and linking the components:

```go
// The main entry point for the agent workflow
err := pipe.Execute(ctx, os.Stdin, os.Stdout, 
		 react.Instructions(shell, write), 
		 pipe.Loop(10, 
		 	 hist.Accumulate(), 
		 	 llm.Generate(model), 
		 	 jsonpkg.Extract(), 
		 	 pipe.Tee(os.Stderr), 
		 	 react.ParseAction(), 
		 	 tools.Execute(shell, write), 
		 	 react.Done(),
		 ),
)
```

**Explanation of the Pipeline:**

* **`pipe.Execute()`:** Initiates the entire pipeline, managing input (`os.Stdin`) and output (`os.Stdout`).
* **`pipe.Loop()`:** Defines the iterative structure, allowing the agent to run up to 10 times.
* **`hist.Accumulate()`:** Ensures that the full conversational history is maintained and passed to the LLM.
* **`llm.Generate(model)`:** Core intelligence component: generates the next thought/action using the specified LLM.
* **`jsonpkg.Extract()`:** Parses the raw LLM output, specifically looking for structured JSON actions.
* **`react.ParseAction()`:** Interprets the structured action and links it to available tools.
* **`tools.Execute()`:** Executes the action (e.g., running shell commands or writing files).
* **`react.Done()`:** The termination condition, signaling that the process should stop.

## Conclusion

The `pipe` package abstracts away the complexities of state machine management and streaming data processing, allowing developers to focus purely on defining the sequence of actions and intelligence required for their agent.