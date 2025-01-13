package orchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools/calculator"
	"github.com/bububa/atomic-agents/tools/searxng"
	"github.com/bububa/instructor-go/pkg/instructor"
)

// Input schema for the Orchestrator Agent. Contains the user's message to be processed.
type Input struct {
	schema.Base
	// ChatMessage The user's input message to be analyzed and responded to.
	ChatMessage string `json:"chat_message" jsonschema:"title=chat_message,description=The user's message to be analyzed and responded to."`
}

type ToolType string

const (
	SearchTool     ToolType = "search"
	CalculatorTool ToolType = "calculator"
)

// Output Combined output schema for the Orchestrator Agent. Contains the tool to use and its parameters.
type Output struct {
	schema.Base
	// Tool The tool to use: 'search' or 'calculator'
	Tool ToolType `json:"tool" jsonschema:"title=tool,enum=search,enum=calculator,description=The tool to use: 'search' or 'calculator'"`
	// SearchParameters the parameters for the search tool
	SearchParameters *searxng.Input `json:"search_parameters" jsonschema:"title=search_parameters,description=The parameters for the search tool. Should only and must have value if tool is 'search'"`
	// CalculatorParameters the parameters for the calculator tool
	CalculatorParameters *calculator.Input `json:"calculator_parameters" jsonschema:"title=calculator_parameters,description=The parameters for the calculator tool. Should only have and must value if tool is 'calculator'"`
}

// FinalAnswer Schema for the final answer generated by the Orchestrator Agent.
type FinalAnswer struct {
	schema.Base
	// FinalAnswer The final answer generated based on the tool output and user query.
	FinalAnswer string `json:"final_answer" jsonschema:"title=final_answer,description=The final answer generated based on the tool output and user query."`
}

type ContextProvider struct{}

func (p *ContextProvider) Title() string {
	return "Current Date"
}

func (p *ContextProvider) Info() string {
	return fmt.Sprintf("Current date in format YYYY-MM-DD:%s", time.Now().Format("2006-01-02"))
}

func Example_orchestration() {
	mockPort := 8080
	mockSearchURL := fmt.Sprintf("http://localhost:%d", mockPort)
	mockQuery := "query with max results"
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

	systemPromptGenerator := systemprompt.NewGenerator(
		systemprompt.WithBackground([]string{
			"- You are an Orchestrator Agent that decides between using a search tool or a calculator tool based on user input.",
			"- Use the search tool for queries requiring factual information, current events, or specific data.",
			"- Use the calculator tool for mathematical calculations and expressions.",
		}),
		systemprompt.WithOutputInstructs([]string{
			"- Analyze the input to determine whether it requires a web search or a calculation.",
			"- For search queries, use the 'search' tool and provide 1-3 relevant search queries.",
			"- For calculations, use the 'calculator' tool and provide the mathematical expression to evaluate.",
			"- When uncertain, prefer using the search tool.",
			"- Format the output using the appropriate schema.",
		}),
		systemprompt.WithContextProviders(new(ContextProvider)),
	)
	agent := agents.NewAgent[Input, Output](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithMemory(mem),
		agents.WithModel("gpt-4o-mini"),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000))
	searchTool := searxng.New(searxng.WithBaseURL(mockSearchURL), searxng.WithMaxResults(3))
	calculatorTool := calculator.New()

	fmt.Println(agent.SystemPrompt())
	fmt.Println("")

	// example inputs
	inputs := []string{
		"Who won the Nobel Prize in Physics in 2024?",
		"Please calculate the sine of pi/3 to the third power",
	}
	finalOutput := new(FinalAnswer)
	finalAgent := agents.NewAgent[Input, FinalAnswer](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithMemory(mem),
		agents.WithModel("gpt-4o-mini"),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000))
	for _, userInput := range inputs {
		input := Input{
			ChatMessage: userInput,
		}
		output := new(Output)
		if err := agent.Run(ctx, &input, output, nil); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("User: %s\n", input.ChatMessage)
		fmt.Printf("Agent: %s\n", output.Tool)
		switch output.Tool {
		case SearchTool:
			if resp, err := searchTool.Run(ctx, output.SearchParameters); err != nil {
				fmt.Println(err)
				return
			} else {
				fmt.Println("SearchTool Result:")
				fmt.Println(resp.Info())
				mem.NewMessage(components.SystemRole, *resp)
				if err := finalAgent.Run(ctx, &input, finalOutput, nil); err != nil {
					fmt.Println(err)
					return
				}
			}
		case CalculatorTool:
			fmt.Printf("tool parameters: %+v\n", output.CalculatorParameters)
			if resp, err := calculatorTool.Run(ctx, output.CalculatorParameters); err != nil {
				fmt.Println(err)
				return
			} else {
				fmt.Printf("CalculatorTool Result: %+v\n", resp.Result)
				mem.NewMessage(components.SystemRole, *resp)
			}
		}
		if err := finalAgent.Run(ctx, &input, finalOutput, nil); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Agent: %s\n", finalOutput.FinalAnswer)
		mem.Reset()
	}
	// Outputs:
	// # IDENTITY and PURPOSE
	// - You are an Orchestrator Agent that decides between using a search tool or a calculator tool based on user input.
	// - Use the search tool for queries requiring factual information, current events, or specific data.
	// - Use the calculator tool for mathematical calculations and expressions.
	//
	// # OUTPUT INSTRUCTIONS
	// - Analyze the input to determine whether it requires a web search or a calculation.
	// - For search queries, use the 'search' tool and provide 1-3 relevant search queries.
	// - For calculations, use the 'calculator' tool and provide the mathematical expression to evaluate.
	// - When uncertain, prefer using the search tool.
	// - Format the output using the appropriate schema.
	// - Always respond using the proper JSON schema.
	// - Always use the available additional information and context to enhance the response.
	//
	// # EXTRA INFORMATION AND CONTEXT
	// ## Current Date
	// Current date in format YYYY-MM-DD:2025-01-03
	//
	// User: Who won the Nobel Prize in Physics in 2024?
	// Agent: search
	// SearchTool Result:
	// TITLE: Result with Metadata
	// URL: https://example.com/metadata
	// CONTENT: Content with metadata
	// METADATA: 2021-01-01
	//
	// TITLE: Result with Published Date
	// URL: https://example.com/published-data
	// CONTENT: Content with published date
	// PUBLISHED DATE: 2022-01-01
	//
	// TITLE: Result without dates
	// URL: https://example.com/no-dates
	// CONTENT: Content without dates
	//
	//
	// Agent: Final Answer is I will search for the winner of the Nobel Prize in Physics in 2024.
	// User: Please calculate the sine of pi/3 to the third power
	// Agent: calculator
	// tool parameters: &{Base:{attachement:<nil>} Expression:sin(pi/3)^3 Params:map[]}
	// CalculatorTool Result: 3
	// Agent: The sine of pi/3 is 0.86602540378. Therefore, (sin(pi/3))^3 = 0.86602540378^3 = 0.64951905284.
}
