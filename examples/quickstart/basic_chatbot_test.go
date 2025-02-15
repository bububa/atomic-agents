package quickstart

import (
	"context"
	"fmt"
	"os"

	"github.com/bububa/instructor-go"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
)

func Example_basicChatbot() {
	ctx := context.Background()
	mem := components.NewMemory(10)
	initMsg := mem.NewMessage(components.AssistantRole, schema.CreateOutput("Hello! How can I assist you today?"))
	agent := agents.NewAgent[schema.Input, schema.Output](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithMemory(mem),
		agents.WithModel(os.Getenv("OPENAI_MODEL")),
		agents.WithTemperature(1),
		agents.WithMaxTokens(1000))
	output := schema.NewOutput("")
	input := schema.NewInput("Today is 2024-01-01, only response with the date without any other words")
	llmResp := new(components.LLMResponse)
	if err := agent.Run(ctx, input, output, llmResp); err != nil {
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
