package optimizer

import "context"

func (po *PromptOptimizer[I]) improvementPrompt(ctx context.Context, entry *OptimizationEntry) (*I, error) {
	improvement := new(PromptImprovement[I])
	err := po.improvementAgent.Run(ctx, entry, improvement, nil)
	if err != nil {
		return nil, err
	}
	// Select the improvement with higher expected impact
	if improvement.ExpectedImpact.Bold > improvement.ExpectedImpact.Incremental {
		return improvement.BoldRedesign, nil
	}
	return improvement.IncrementalImprovement, nil
}
