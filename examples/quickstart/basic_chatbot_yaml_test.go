package quickstart

import (
	"context"
	"fmt"
	"os"

	"github.com/bububa/instructor-go"
	"gopkg.in/yaml.v3"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
)

type Output struct {
	schema.Base `yaml:"-"`
	Foods       []string `yaml:"foods" fake:"{dinner}" jsonschema:"title=foods,description=The list of food items that the user has ordered."`
}

func (o Output) String() string {
	bs, _ := yaml.Marshal(o)
	return string(bs)
}

func Example_basicChatbotYAML() {
	ctx := context.Background()
	agent := agents.NewAgent[schema.Input, Output](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI, instructor.ModeYAML)),
		agents.WithModel(os.Getenv("OPENAI_MODEL")),
		agents.WithTemperature(1),
		agents.WithMaxTokens(1000))
	output := new(Output)
	input := schema.NewInput("give me some suggest for dinner which is healthy")
	llmResp := new(components.LLMResponse)
	if err := agent.Run(ctx, input, output, llmResp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(agent.SystemPrompt())
	fmt.Println("")
	fmt.Printf("User: %s\n", input.ChatMessage)
	fmt.Printf("Agent: %s\n", output.String())
}
