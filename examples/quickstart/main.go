package main

import (
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

func newInstructor() instructor.Instructor {
	authToken := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_API_BASE_URL")
	cfg := openai.DefaultConfig(authToken)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	clt := openai.NewClientWithConfig(cfg)
	return instructor.FromOpenAI(clt, instructor.WithMode(instructor.ModeJSON), instructor.WithMaxRetries(3), instructor.WithValidation())
}

func main() {
	ExampleBasicCustomChatBot()
}
