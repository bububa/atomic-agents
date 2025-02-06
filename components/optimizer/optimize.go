package optimizer

import (
	"context"
	"fmt"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/schema"
)

// OptimizePrompt performs automated optimization of an LLM prompt and generates a response.
// It uses a sophisticated optimization process that includes:
// - Prompt quality assessment
// - Iterative refinement
// - Performance measurement
// - Response validation
//
// The optimization process follows these steps:
// 1. Initialize optimization with the given configuration
// 2. Assess initial prompt quality
// 3. Apply iterative improvements based on assessment
// 4. Validate against optimization goals
// 5. Generate response using the optimized prompt
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - llm: Language model instance to use for optimization
//   - config: Configuration controlling the optimization process
//
// Returns:
//   - optimizedPrompt: The refined and improved prompt text
//   - response: The LLM's response using the optimized prompt
//   - err: Any error encountered during optimization
//
// Example usage:
//
//	optimizedPrompt, response, err := OptimizePrompt(ctx, llmInstance, OptimizationConfig{
//	    Prompt:      "Initial prompt text...",
//	    Description: "Task description for optimization",
//	    Metrics:     []Metric{{Name: "Clarity", Description: "..."}},
//	    Threshold:   15.0, // Minimum acceptable quality score
//	})
//
// The function uses a PromptOptimizer internally and configures it with:
// - Debug logging for prompts and responses
// - Custom evaluation metrics
// - Configurable rating system
// - Retry mechanisms for reliability
// - Quality thresholds for acceptance
func OptimizePrompt[I schema.Schema, O schema.Schema](ctx context.Context, prompt *I, config OptimizationConfig, agentOpts ...agents.Option) (*I, *O, error) {
	// Configure and create optimizer instance
	optimizer := NewPromptOptimizer[I](agentOpts, config.Description,
		WithCustomMetrics(config.Metrics...),
		WithRatingSystem(config.RatingSystem),
		WithOptimizationGoal(fmt.Sprintf("Optimize the prompt for %s", config.Description)),
		WithMaxRetries(config.MaxRetries),
		WithRetryDelay(config.RetryDelay),
		WithThreshold(config.Threshold),
	)

	// Perform prompt optimization
	optimizedPromptObj, err := optimizer.OptimizePrompt(ctx, prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("optimization failed: %w", err)
	}

	// Validate optimization result
	if optimizedPromptObj == nil {
		return nil, nil, fmt.Errorf("optimized prompt is nil")
	}
	agent := agents.NewAgent[I, O](agentOpts...)

	// Generate response using optimized prompt
	response := new(O)
	if err := agent.Run(ctx, optimizedPromptObj, response, nil); err != nil {
		return optimizedPromptObj, nil, fmt.Errorf("response generation failed: %w", err)
	}

	return optimizedPromptObj, response, nil
}
