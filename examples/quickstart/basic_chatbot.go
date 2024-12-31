package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/schema"
)

func ExampleBasicChatBot() {
	mem := components.NewMemory(10)
	initMsg := mem.NewMessage(components.AssistantRole, schema.NewOutput("Hello! How can I assist you today?"))
	agent := agents.NewAgent[schema.Input, schema.Output](
		agents.WithClient(newInstructor()),
		agents.WithMemory(mem),
		agents.WithModel("gpt-4o-mini"),
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
