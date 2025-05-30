package examples

import (
	"os"

	"github.com/bububa/instructor-go"
	"github.com/bububa/instructor-go/instructors"
	cohereClient "github.com/cohere-ai/cohere-go/v2/client"
	cohereOption "github.com/cohere-ai/cohere-go/v2/option"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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
		opts := make([]option.RequestOption, 0, 2)

		opts = append(opts, option.WithAPIKey(authToken))
		if baseURL != "" {
			opts = append(opts, option.WithBaseURL(baseURL))
		}
		clt := openai.NewClient(opts...)
		return instructors.FromOpenAI(&clt, instructor.WithMode(mode), instructor.WithMaxRetries(1), instructor.WithValidation(), instructor.WithVerbose())
	}
}
