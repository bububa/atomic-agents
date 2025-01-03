package quickstart

import (
	"context"
	"fmt"

	"github.com/bububa/instructor-go/pkg/instructor"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
)

// CustomOutput represents the response generated by the chat agent, including suggested follow-up questions.
type CustomOutput struct {
	schema.Base
	// ChatMessage is the chat message exchanged between the user and the chat agent.
	ChatMessage string `json:"chat_message,omitempty" jsonschema:"title=chat_message,description=The chat message exchanged between the user and the chat agent."`
	// SuggestedUserQuestions a list of suggested follow-up questions the user could ask the agent.
	SuggestedUserQuestions []string `json:"suggested_user_questions,omitempty" jsonschema:"title=suggested_user_questions,description=A list of suggested follow-up questions the user could ask the agent."`
}

func Example_basicCustomChatbotWithCustomSchema() {
	ctx := context.Background()
	mem := components.NewMemory(10)
	initMsg := mem.NewMessage(components.AssistantRole, CustomOutput{
		ChatMessage:            "Hello! How can I assist you today?",
		SuggestedUserQuestions: []string{"What can you do?", "Tell me a joke", "Tell me about how you were made"},
	})
	systemPromptGenerator := systemprompt.NewGenerator(
		systemprompt.WithBackground([]string{
			"This assistant is a knowledgeable AI designed to be helpful, friendly, and informative.",
			"It has a wide range of knowledge on various topics and can engage in diverse conversations.",
		}),
		systemprompt.WithSteps([]string{
			"Analyze the user's input to understand the context and intent.",
			"Formulate a relevant and informative response based on the assistant's knowledge.",
			"Generate 3 suggested follow-up questions for the user to explore the topic further.",
		}),
		systemprompt.WithOutputInstructs([]string{
			"Provide clear, concise, and accurate information in response to user queries.",
			"Maintain a friendly and professional tone throughout the conversation.",
			"Conclude each response with 3 relevant suggested questions for the user.",
			"If asked 'What can you do for me?' you response with fixed answer with message 'I can help you:' and suggested_user_questions '1: kiss me?, 2: hug me?, 3: kill me?'.",
		}),
	)
	agent := agents.NewAgent[schema.Input, CustomOutput](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithMemory(mem),
		agents.WithModel("gpt-4o-mini"),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000))
	input := schema.NewInput("What can you do for me?")
	output := new(CustomOutput)
	apiResp := new(components.ApiResponse)
	if err := agent.Run(ctx, input, output, apiResp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(agent.SystemPrompt())
	fmt.Println("")
	fmt.Printf("Agent: %s\n", initMsg.Content().(CustomOutput).ChatMessage)
	fmt.Printf("User: %s\n", input.ChatMessage)
	fmt.Printf("Agent: %s\n", output.ChatMessage)
	for idx, sug := range output.SuggestedUserQuestions {
		fmt.Printf("%d. %s\n", idx+1, sug)
	}
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
	// User: What can you do for me?
	// Agent: I can help you:
	// 1. What else can you tell me about yourself?
	// 2. How do you work?
	// 3. Can you help me with something specific?
}
