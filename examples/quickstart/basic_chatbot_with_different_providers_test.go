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

func Example_basicChatbotWithDifferentProviders() {
	ctx := context.Background()
	providers := []instructor.Provider{instructor.ProviderOpenAI, instructor.ProviderAnthropic, instructor.ProviderCohere}
	for _, provider := range providers {
		var model string
		switch provider {
		case instructor.ProviderOpenAI:
			model = os.Getenv("OPENAI_MODEL")
		case instructor.ProviderAnthropic:
			model = "claude-3-5-haiku-20241022"
		case instructor.ProviderCohere:
			model = "command-r-plus"
		}
		agent := agents.NewAgent[schema.Input, schema.Output](
			agents.WithClient(examples.NewInstructor(provider)),
			agents.WithModel(model),
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
		fmt.Printf("User: %s\n", input.ChatMessage)
		fmt.Printf("Agent: %s\n", output.ChatMessage)
	}
}
