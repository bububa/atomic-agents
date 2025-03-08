package optimizer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/schema"
)

// OptimizerOption is a function type for configuring the PromptOptimizer.
// It follows the functional options pattern for flexible configuration.
type OptimizerOption func(*OptimizerConfig)

// IterationCallback is a function type for monitoring optimization progress.
// It's called after each iteration with the current state.
type IterationCallback func(iteration int, entry OptimizationEntry)

// OptimizerConfig is PromptOptimizer config
type OptimizerConfig struct {
	// customMetrics defines additional evaluation criteria
	customMetrics []Metric

	// optimizationGoal specifies the target outcome
	optimizationGoal string

	// history tracks the optimization process
	history []OptimizationEntry

	// ratingSystem defines the scoring methodology
	ratingSystem string

	// threshold sets the minimum acceptable score
	threshold float64

	// iterationCallback monitors optimization progress
	iterationCallback IterationCallback

	// maxRetries specifies retry attempts for failed operations
	maxRetries int

	// retryDelay sets the wait time between retries
	retryDelay time.Duration

	// memorySize limits the optimization history length
	memorySize int

	// iterations counts the optimization steps performed
	iterations int
}

// PromptOptimizer orchestrates the prompt optimization process.
// It manages the iterative refinement of prompts through assessment,
// improvement suggestions, and validation.
type PromptOptimizer[T schema.Schema] struct {
	assessmentAgent  *agents.Agent[T, PromptAssessment]
	improvementAgent *agents.Agent[OptimizationEntry, PromptImprovement[T]]

	// taskDesc describes the intended use of the prompt
	taskDesc string
	OptimizerConfig
}

// NewPromptOptimizer creates a new instance of PromptOptimizer with the given configuration.
//
// Parameters:
//   - llm: Language Learning Model interface for generating and evaluating prompts
//   - debugManager: Debug manager for logging and debugging
//   - taskDesc: Description of the optimization task
//   - opts: Optional configuration options
//
// Returns:
//   - Configured PromptOptimizer instance
func NewPromptOptimizer[I schema.Schema](agentOpts []agents.Option, taskDesc string, opts ...OptimizerOption) *PromptOptimizer[I] {
	optimizer := &PromptOptimizer[I]{
		taskDesc: taskDesc,
		OptimizerConfig: OptimizerConfig{
			history:    []OptimizationEntry{},
			threshold:  0.8,
			maxRetries: 3,
			retryDelay: time.Second * 2,
			memorySize: 2,
			iterations: 5,
		},
	}

	for _, opt := range opts {
		opt(&optimizer.OptimizerConfig)
	}
	assessOpts := make([]agents.Option, 0, len(agentOpts)+1)
	improveOpts := make([]agents.Option, 0, len(agentOpts)+1)
	for _, opt := range agentOpts {
		assessOpts = append(assessOpts, opt)
		improveOpts = append(improveOpts, opt)
	}
	assessOpts = append(assessOpts, agents.WithSystemPromptGenerator(optimizer.assessmentPromptGenerator()))
	improveOpts = append(improveOpts, agents.WithSystemPromptGenerator(optimizer.improvementPromptGenerator()))
	optimizer.assessmentAgent = agents.NewAgent[I, PromptAssessment](assessOpts...)
	optimizer.improvementAgent = agents.NewAgent[OptimizationEntry, PromptImprovement[I]](improveOpts...)

	return optimizer
}

// OptimizePrompt performs iterative optimization of the initial prompt to meet the specified goal.
//
// The optimization process:
// 1. Assesses the current prompt
// 2. Records assessment in history
// 3. Checks if optimization goal is met
// 4. Generates improved prompt if goal not met
// 5. Repeats until goal is met or max iterations reached
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//
// Returns:
//   - Optimized prompt
//   - Error if optimization fails
func (po *PromptOptimizer[I]) OptimizePrompt(ctx context.Context, prompt *I) (*I, error) {
	currentPrompt := prompt
	var bestPrompt *I
	var bestScore float64

	for i := range po.iterations {
		var entry OptimizationEntry
		var err error

		// Retry loop for assessment
		for attempt := range po.maxRetries {
			if err = po.assessPrompt(ctx, currentPrompt, &entry.Assessment); err == nil {
				break
			}

			log.Printf("Error in iteration %d, attempt %d: %v\n", i+1, attempt+1, err)
			if attempt < po.maxRetries-1 {
				log.Printf("Retrying in %v...\n", po.retryDelay)
				time.Sleep(po.retryDelay)
			}
		}

		if err != nil {
			return bestPrompt, fmt.Errorf("optimization failed at iteration %d after %d attempts: %w", i+1, po.maxRetries, err)
		}

		po.history = append(po.history, entry)

		// Execute iteration callback if set
		if po.iterationCallback != nil {
			po.iterationCallback(i+1, entry)
		}

		// Update best prompt if current score is higher
		if entry.Assessment.OverallScore > bestScore {
			bestScore = entry.Assessment.OverallScore
			bestPrompt = currentPrompt
		}

		// Check if optimization goal is met
		goalMet, err := po.isOptimizationGoalMet(entry.Assessment)
		if err != nil {
			log.Printf("Error checking optimization goal: %v\n", err)
		} else if goalMet {
			log.Printf("Optimization complete after %d iterations. Goal achieved.\n", i+1)
			return currentPrompt, nil
		}

		// Generate improved prompt
		improvedPrompt, err := po.improvementPrompt(ctx, &entry)
		if err != nil {
			log.Printf("Failed to generate improved prompt at iteration %d: %v\n", i+1, err)
			continue
		}

		currentPrompt = improvedPrompt
		log.Printf("Iteration %d complete. New prompt: %s\n", i+1, currentPrompt)
	}

	return bestPrompt, nil
}

// GetOptimizationHistory returns the complete history of optimization attempts.
func (po *PromptOptimizer[I]) GetOptimizationHistory() []OptimizationEntry {
	return po.history
}

// recentHistory returns the most recent optimization entries based on memory size.
func (po *PromptOptimizer[I]) RecentHistory() []OptimizationEntry {
	if len(po.history) <= po.memorySize {
		return po.history
	}
	return po.history[len(po.history)-po.memorySize:]
}
