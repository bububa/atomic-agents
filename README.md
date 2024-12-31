# Atomic Agents golang version

[![Go Reference](https://pkg.go.dev/badge/github.com/bububa/atomic-agents.svg)](https://pkg.go.dev/github.com/bububa/atomic-agents)
[![Go](https://github.com/bububa/atomic-agents/actions/workflows/go.yml/badge.svg)](https://github.com/bububa/atomic-agents/actions/workflows/go.yml)
[![goreleaser](https://github.com/bububa/atomic-agents/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/bububa/atomic-agents/actions/workflows/goreleaser.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/bububa/atomic-agents.svg)](https://github.com/bububa/atomic-agents)
[![GoReportCard](https://goreportcard.com/badge/github.com/bububa/atomic-agents)](https://goreportcard.com/report/github.com/bububa/atomic-agents)
[![GitHub license](https://img.shields.io/github/license/bububa/atomic-agents.svg)](https://github.com/bububa/atomic-agents/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/bububa/atomic-agents.svg)](https://GitHub.com/bububa/atomic-agents/releases/)

This is a re-implementation for [Atomic Agents](https://github.com/AtomicAgents/atomic-agents) in golang.

> The Atomic Agents framework is designed around the concept of atomicity to be an extremely lightweight and modular framework for building Agentic AI pipelines and applications without sacrificing developer experience and maintainability. The framework provides a set of tools and agents that can be combined to create powerful applications. It is built on top of Instructor and leverages the power of Pydantic for data and schema validation and serialization. All logic and control flows are written in Python, enabling developers to apply familiar best practices and workflows from traditional software development without compromising flexibility or clarity.

## Anatomy of an Agent

In Atomic Agents, an agent is composed of several key components:

- **System Prompt:** Defines the agent's behavior and purpose.
- **Input Schema:** Specifies the structure and validation rules for the agent's input.
- **Output Schema:** Specifies the structure and validation rules for the agent's output.
- **Memory:** Stores conversation history or other relevant data.
- **Context Providers:** Inject dynamic context into the agent's system prompt at runtime.

Here's a high-level architecture diagram:

<!-- ![alt text](./.assets/architecture_highlevel_overview.png) -->
<img src="https://github.com/AtomicAgents/atomic-agents/blob/main/.assets/architecture_highlevel_overview.png" alt="High-level architecture overview of Atomic Agents" width="600"/>
<img src="https://github.com/AtomicAgents/atomic-agents/raw/main/.assets/what_is_sent_in_prompt.png" alt="Diagram showing what is sent to the LLM in the prompt" width="600"/>

For more details please read from the original website.

## Why reinvent wheel

- don't like python
- easy integrate into exists golang projects
- better performance

## Installation

```bash
go get -u github.com/bububa/atomic-agents
```

## Project Structure

Atomic Agents with the following main components:

1. `agents/`: The core Atomic Agents library
2. `components/`: The Atomic Agents components contains `Message`, `Memory`, `SystemPromptGenerator`, `SystemPromptContextProvider` utilities
3. `schema/`: Defines the Input/Output schema structures and interfaces
4. `atomic-examples/`: Example projects showcasing Atomic Agents usage
5. `tools/`: A collection of tools that can be used with Atomic Agents

## Quickstart & Examples

A complete list of examples can be found in the [examples](./examples/) directory.

Here's a quick snippet demonstrating how easy it is to create a powerful agent with Atomic Agents:

```golang
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/schema"
)

func ExampleBasicCustomChatBot() {
	mem := components.NewMemory(10)
	initMsg := mem.NewMessage(components.AssistantRole, schema.NewOutput("How do you do? What can I do for you? Tell me, pray, what is your need today?"))
	systemPromptGenerator := systemprompt.NewGenerator(
		systemprompt.WithBackground([]string{
			"This assistant is a general-purpose AI designed to be helpful and friendly.",
		}),
		systemprompt.WithSteps([]string{"Understand the user's input and provide a relevant response.", "Respond to the user."}),
		systemprompt.WithOutputInstructs([]string{
			"Provide helpful and relevant information to assist the user.",
			"Be friendly and respectful in all interactions.",
			"Always answer in rhyming verse.",
		}),
	)
	agent := agents.NewAgent[schema.Input, schema.Output](
		agents.WithClient(newInstructor()),
		agents.WithMemory(mem),
		agents.WithModel("gpt-4o-mini"),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000))
	fmt.Println(agent.SystemPrompt())
	fmt.Println("Agent:", initMsg.Content().(*schema.Output).ChatMessage)
	// Output:
	// # IDENTITY and PURPOSE
	// - This is a conversation with a helpful and friendly AI assistant.
	//
	// # OUTPUT INSTRUCTIONS
	// - Always respond using the proper JSON schema.
	// - Always use the available additional information and context to enhance the response.
	// Agent: Hello! How can I assist you today?
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()
	output := schema.NewOutput("")
	for {
		fmt.Print("User: ")
		txt, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		txt = strings.TrimSpace(txt)
		if txt == "/exit" || txt == "/quit" {
			break
		}
		fmt.Println(txt)
		input := schema.NewInput(txt)
		if err := agent.Run(ctx, input, output); err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println("Agent: ", output.ChatMessage)
	}
}
```

## License

This project is licensed under the MIT License—see the [LICENSE](LICENSE) file for details.
