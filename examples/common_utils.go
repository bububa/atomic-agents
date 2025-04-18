package examples

import (
	"os"

	"github.com/bububa/instructor-go"
	"github.com/bububa/instructor-go/instructors"
	cohereClient "github.com/cohere-ai/cohere-go/v2/client"
	cohereOption "github.com/cohere-ai/cohere-go/v2/option"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	openai "github.com/sashabaranov/go-openai"
)

func NewInstructor(provider instructor.Provider, modes ...instructor.Mode) instructor.Instructor {
	mode := instructor.ModeJSON
	if len(modes) > 0 {
		mode = modes[0]
	}
	switch provider {
	case instructor.ProviderAnthropic:
		authToken := os.Getenv("ANTHROPIC_API_KEY")
		baseURL := os.Getenv("ANTHROPIC_API_BASE_URL")
		opts := make([]anthropic.ClientOption, 0, 1)
		if baseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(baseURL))
		}
		clt := anthropic.NewClient(authToken, opts...)
		return instructors.FromAnthropic(clt, instructor.WithMode(mode), instructor.WithMaxRetries(1), instructor.WithValidation())
	case instructor.ProviderCohere:
		authToken := os.Getenv("COHERE_API_KEY")
		baseURL := os.Getenv("COHERE_API_BASE_URL")
		opts := make([]cohereOption.RequestOption, 0, 2)
		opts = append(opts, cohereOption.WithToken(authToken))
		if baseURL != "" {
			opts = append(opts, cohereOption.WithBaseURL(baseURL))
		}
		clt := cohereClient.NewClient(opts...)
		return instructors.FromCohere(clt, instructor.WithMode(mode), instructor.WithMaxRetries(1), instructor.WithValidation())
	default:
		authToken := os.Getenv("OPENAI_API_KEY")
		baseURL := os.Getenv("OPENAI_BASE_URL")
		cfg := openai.DefaultConfig(authToken)
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
		clt := openai.NewClientWithConfig(cfg)
		return instructors.FromOpenAI(clt, instructor.WithMode(mode), instructor.WithMaxRetries(1), instructor.WithValidation())
	}
}
