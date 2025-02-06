package optimizer

import (
	"time"
)

type Option = func(*OptimizerConfig)

// WithCustomMetrics sets custom evaluation metrics for the optimizer.
func WithCustomMetrics(metrics ...Metric) Option {
	return func(c *OptimizerConfig) {
		c.customMetrics = metrics
	}
}

// WithOptimizationGoal sets the target goal for optimization.
func WithOptimizationGoal(goal string) Option {
	return func(c *OptimizerConfig) {
		c.optimizationGoal = goal
	}
}

// WithRatingSystem sets the rating system to use (numerical or letter grades).
func WithRatingSystem(system string) Option {
	return func(c *OptimizerConfig) {
		c.ratingSystem = system
	}
}

// WithThreshold sets the minimum acceptable score threshold.
func WithThreshold(threshold float64) Option {
	return func(c *OptimizerConfig) {
		c.threshold = threshold
	}
}

// WithIterationCallback sets a callback function to be called after each iteration.
func WithIterationCallback(callback IterationCallback) Option {
	return func(c *OptimizerConfig) {
		c.iterationCallback = callback
	}
}

// WithIterations sets the maximum number of optimization iterations.
func WithIterations(iterations int) Option {
	return func(c *OptimizerConfig) {
		c.iterations = iterations
	}
}

// WithMaxRetries sets the maximum number of retry attempts per iteration.
func WithMaxRetries(maxRetries int) Option {
	return func(c *OptimizerConfig) {
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the delay duration between retry attempts.
func WithRetryDelay(delay time.Duration) Option {
	return func(c *OptimizerConfig) {
		c.retryDelay = delay
	}
}

// WithMemorySize sets the number of recent optimization entries to keep in memory.
func WithMemorySize(size int) Option {
	return func(c *OptimizerConfig) {
		c.memorySize = size
	}
}
