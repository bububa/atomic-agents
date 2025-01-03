package quickstart

import (
	"context"
	"fmt"

	"github.com/bububa/instructor-go/pkg/instructor"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
)

func Example_basicChatbotWithDifferentProviders() {
	ctx := context.Background()
	providers := []instructor.Provider{instructor.ProviderOpenAI, instructor.ProviderAnthropic, instructor.ProviderCohere}
	for _, provider := range providers {
		var model string
		switch provider {
		case instructor.ProviderOpenAI:
			model = "gpt-4o-mini"
		case instructor.ProviderAnthropic:
			model = "claude-3-5-haiku-20241022"
		case instructor.ProviderCohere:
			model = "command-r-plus"
		}
		mem := components.NewMemory(10)
		initMsg := mem.NewMessage(components.AssistantRole, schema.CreateOutput("Hello! How can I assist you today?"))
		agent := agents.NewAgent[schema.Input, schema.Output](
			agents.WithClient(examples.NewInstructor(provider)),
			agents.WithMemory(mem),
			agents.WithModel(model),
			agents.WithTemperature(0.5),
			agents.WithMaxTokens(1000))
		output := schema.NewOutput("")
		input := schema.NewInput("Today is 2024-01-01, only response with the date without any other words")
		apiResp := new(components.ApiResponse)
		if err := agent.Run(ctx, input, output, apiResp); err != nil {
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
}
