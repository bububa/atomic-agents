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
