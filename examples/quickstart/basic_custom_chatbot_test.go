package quickstart

import (
	"context"
	"fmt"
	"os"

	"github.com/bububa/instructor-go"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt/cot"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
)

func Example_basicCustomChatbot() {
	ctx := context.Background()
	systemPromptGenerator := cot.New(
		cot.WithBackground([]string{
			"- This assistant is a general-purpose AI designed to be helpful and friendly.",
			"- Your name is 'Atomic Agent Custom Chatbot'",
		}),
		cot.WithSteps([]string{"- Understand the user's input and provide a relevant response.", "- Respond to the user."}),
		cot.WithOutputInstructs([]string{
			"- Provide helpful and relevant information to assist the user.",
			"- Be friendly and respectful in all interactions.",
			"- If ask your name, only your name directly withour any other additional words.",
		}),
	)
	agent := agents.NewAgent[schema.Input, schema.Output](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithModel(os.Getenv("OPENAI_MODEL")),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(1),
		agents.WithMaxTokens(1000))
	input := schema.NewInput("What is your name?")
	output := schema.NewOutput("")
	llmResp := new(components.LLMResponse)
	if err := agent.Run(ctx, input, output, llmResp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(agent.SystemPrompt())
	fmt.Println("")
	fmt.Printf("User: %s\n", input.ChatMessage)
	fmt.Printf("Agent: %s\n", output.ChatMessage)
	// Output:
	// # IDENTITY and PURPOSE
	// - This assistant is a general-purpose AI designed to be helpful and friendly.
	// - Your name is 'Atomic Agent Custom Chatbot'
	//
	// # INTERNAL ASSISTANT STEPS
	// - Understand the user's input and provide a relevant response.
	// - Respond to the user.
	//
	// # OUTPUT INSTRUCTIONS
	// - Provide helpful and relevant information to assist the user.
	// - Be friendly and respectful in all interactions.
	// - If ask your name, only your name directly withour any other additional words.
	// - Always respond using the proper JSON schema.
	// - Always use the available additional information and context to enhance the response.
	//
	// Agent: How do you do? What can I do for you? Tell me, pray, what is your need today?
	// User: What is your name?
	// Agent: Atomic Agent Custom Chatbot
}
