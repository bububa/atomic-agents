package websearch

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/bububa/instructor-go"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt/cot"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools/searxng"
)

// Input defines the input schema for the QuestionAnsweringAgent.
type Input struct {
	schema.Base
	// Question is a question that needs to be answered based on the provided context.
	Question string `json:"question" jsonschema:"title=question,description=A question that needs to be answered based on the provided context."`
}

// Output defines the output schema for the QuestionAnsweringAgent.
type Output struct {
	schema.Base
	// MarkdownOutput The answer to the question in markdown format.
	MarkdownOutput string `json:"markdown_output" jsonschema:"title=markdown_output,description=The answer to the question in markdown format."`
	// References is a list of up to 3 HTTP URLs used as references for the answer.
	References []url.URL `json:"references" jsonschema:"title=references,description=A list of up to 3 HTTP URLs used as references for the answer."`
	// FollowUpQuestions is a list of up to 3 follow-up questions related to the answer."
	FollowUpQuestions []string `json:"follow_up_questions" jsonschema:"title=follow_up_questions,description=A list of up to 3 follow-up questions related to the answer."`
}

func Example_websearch() {
	mockQuery := "Tell me about the Atomic Agents AI agent framework."
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockResult := searxng.Output{
		Results: []searxng.SearchResultItem{
			{Title: "Result with Metadata", URL: "https://example.com/metadata", Content: "Content with metadata", Query: mockQuery, Metadata: "2021-01-01"},
			{Title: "Result with Published Date", Content: "Content with published date", URL: "https://example.com/published-data", Query: mockQuery, PublishedDate: "2022-01-01"},
			{Title: "Result without dates", Content: "Content without dates", URL: "https://example.com/no-dates", Query: mockQuery},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := startSearxngServer(mockPort, &mockResult)
	defer srv.Shutdown(ctx)
	mem := components.NewMemory(10)
	searchInput := searxng.NewInput(searxng.GeneralCategory, []string{mockQuery})
	searchTool := searxng.New(searxng.WithBaseURL(mockSearchURL), searxng.WithMaxResults(3))
	searchOutput := new(searxng.Output)
	if err := searchTool.Run(ctx, searchInput, searchOutput); err != nil {
		fmt.Println(err)
		return
	}

	systemPromptGenerator := cot.New(
		cot.WithBackground([]string{
			"- You are an intelligent question answering expert.",
			"- Your task is to provide accurate and detailed answers to user questions based on the given context.",
		}),
		cot.WithSteps([]string{
			"- You will receive a question and the context information.",
			"- Generate a detailed and accurate answer based on the context.",
			"- Provide up to 3 relevant references (HTTP URLs) used in formulating the answer.",
			"- Generate up to 3 follow-up questions related to the answer.",
		}),
		cot.WithOutputInstructs([]string{
			"- Ensure clarity and conciseness in each answer.",
			"- Ensure the answer is directly relevant to the question and context provided.",
			"- Include up to 3 relevant HTTP URLs as references.",
			"- Provide up to 3 follow-up questions to encourage further exploration of the topic.",
		}),
		cot.WithContextProviders(searchOutput),
	)
	agent := agents.NewAgent[Input, Output](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithMemory(mem),
		agents.WithModel(os.Getenv("OPENAI_MODEL")),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000))

	input := &Input{
		Question: mockQuery,
	}
	output := new(Output)
	llmResp := new(components.LLMResponse)
	if err := agent.Run(ctx, input, output, llmResp); err != nil {
		fmt.Println(err)
		return
	}
	// Display the results
	// Print the answer using Rich's Markdown rendering
	fmt.Printf("Answer: %s\n", output.MarkdownOutput)
	fmt.Println()
	// Print references
	fmt.Println("References:")
	for _, v := range output.References {
		fmt.Printf("- %s\n", v.String())
	}
	fmt.Println()
	// Print follow-up questions
	fmt.Println("Follow-up Questions:")
	for _, v := range output.FollowUpQuestions {
		fmt.Printf("- %s\n", v)
	}
}
