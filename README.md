# Atomic Agents Golang version

[![Go Reference](https://pkg.go.dev/badge/github.com/bububa/atomic-agents.svg)](https://pkg.go.dev/github.com/bububa/atomic-agents)
[![Go](https://github.com/bububa/atomic-agents/actions/workflows/go.yml/badge.svg)](https://github.com/bububa/atomic-agents/actions/workflows/go.yml)
[![goreleaser](https://github.com/bububa/atomic-agents/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/bububa/atomic-agents/actions/workflows/goreleaser.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/bububa/atomic-agents.svg)](https://github.com/bububa/atomic-agents)
[![GoReportCard](https://goreportcard.com/badge/github.com/bububa/atomic-agents)](https://goreportcard.com/report/github.com/bububa/atomic-agents)
[![GitHub license](https://img.shields.io/github/license/bububa/atomic-agents.svg)](https://github.com/bububa/atomic-agents/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/bububa/atomic-agents.svg)](https://GitHub.com/bububa/atomic-agents/releases/)

Building AI agents, atomically. This is an implementation for [Atomic Agents](https://github.com/BrainBlend-AI/atomic-agents) in Golang.

> The Atomic Agents framework is designed around the concept of atomicity to be an extremely lightweight and modular framework for building Agentic AI pipelines and applications without sacrificing developer experience and maintainability. The framework provides a set of tools and agents that can be combined to create powerful applications. It is built on top of [instructor](https://go.useinstructor.com) and leverages the power of [jsonschema](https://github.com/invopop/jsonschema) for data and schema validation and serialization. All logic and control flows are written in Golang, enabling developers to apply familiar best practices and workflows from traditional software development without compromising flexibility or clarity.

## Anatomy of an Agent

In Atomic Agents, an agent is composed of several key components:

- **System Prompt:** Defines the agent's behavior and purpose.
- **Input Schema:** Specifies the structure and validation rules for the agent's input.
- **Output Schema:** Specifies the structure and validation rules for the agent's output.
- **Memory:** Stores conversation history or other relevant data.
- **Context Providers:** Inject dynamic context into the agent's system prompt at runtime.

Here's a high-level architecture diagram:

<!-- ![alt text](./.assets/architecture_highlevel_overview.png) -->
<img src="https://github.com/BrainBlend-AI/atomic-agents/blob/main/.assets/architecture_highlevel_overview.png" alt="High-level architecture overview of Atomic Agents" width="600"/>
<img src="https://github.com/BrainBlend-AI/atomic-agents/raw/main/.assets/what_is_sent_in_prompt.png" alt="Diagram showing what is sent to the LLM in the prompt" width="600"/>

For more details please read from the original website.

## Why reinvent wheel

- don't like python
- easy integrate into exists Golang projects
- better performance

## Installation

```bash
go get -u github.com/bububa/atomic-agents
```

## Project Structure

Atomic Agents with the following main components:

1. `agents/`: The core Atomic Agents library
2. `components/`: The Atomic Agents components

- `message`: Defines the Message structure for input/output
- `memory`: Defines a in memory Memory Store
- `systemprompt`: Contains SystemPrompt `Generator` and `ContextProvider`
- `embedder`: Defines the embedder interface, contains several `Provider` including `OpenAI`, `Gemini`, `VoyageAI`, `HuggingFace`, `Cohere` implementations
- `vectordb`: Defines a vectordb interface, contains several `Provider`s including `Memory`, `Chromem`, `Milvus`
- `document` Defines a `Document` interface use for RAG, implemented `File`, `Http` document types. Provide a `Parser` interface which transform document content into specific string with `PDFParser` and `HTML to markdown` parsers implementations.

3. `schema/`: Defines the Input/Output schema structures and interfaces
4. `examples/`: Example projects showcasing Atomic Agents usage
5. `tools/`: A collection of tools that can be used with Atomic Agents

## Quickstart & Examples

A complete list of examples can be found in the [examples](./examples/) directory.

Here's a quick snippet demonstrating how easy it is to create a powerful agent with Atomic Agents:

```golang
package main

import (
	"context"
	"fmt"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/schema"
)

func main() {
	ctx := context.Background()
	mem := components.NewMemory(10)
	initMsg := mem.NewMessage(components.AssistantRole, schema.CreateOutput("Hello! How can I assist you today?"))
	agent := agents.NewAgent[schema.Input, schema.Output](
		agents.WithClient(newInstructor()),
		agents.WithMemory(mem),
		agents.WithModel("gpt-4o-mini"),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000))
	output := schema.NewOutput("")
	input := schema.NewInput("Today is 2024-01-01, only response with the date without any other words")
	if err := agent.Run(ctx, input, output); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(agent.SystemPrompt())
	fmt.Println("")
	fmt.Printf("Agent: %s\n", initMsg.Content().(schema.Output).ChatMessage)
	fmt.Printf("User: %s\n", input.ChatMessage)
	fmt.Printf("Agent: %s\n", output.ChatMessage)
	// Output:
	// # IDENTITY and PURPOSE
	// - This is a conversation with a helpful and friendly AI assistant.
	//
	// # OUTPUT INSTRUCTIONS
	// - Always respond using the proper JSON schema.
	// - Always use the available additional information and context to enhance the response.
	//
	// Agent: Hello! How can I assist you today?
	// User: Today is 2024-01-01, only response with the date without any other words
	// Agent: 2024-01-01
}
```

## License

This project is licensed under the MIT License—see the [LICENSE](LICENSE) file for details.
