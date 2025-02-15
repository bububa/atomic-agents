package optimizer

import (
	"context"
	"fmt"
)

// assessPrompt evaluates a prompt's quality and effectiveness using the configured LLM.
// It performs a comprehensive analysis considering multiple factors including custom metrics,
// optimization goals, and historical context.
//
// The assessment process:
// 1. Constructs an evaluation prompt incorporating task description and history
// 2. Requests LLM evaluation of the prompt
// 3. Parses and validates the assessment response
// 4. Normalizes grading for consistency
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - prompt: The prompt to be assessed
//
// Returns:
//   - OptimizationEntry containing the assessment results
//   - Error if assessment fails
//
// The assessment evaluates:
//   - Custom metrics specified in PromptOptimizer
//   - Prompt strengths with examples
//   - Weaknesses with improvement suggestions
//   - Overall effectiveness and efficiency
//   - Alignment with optimization goals
func (po *PromptOptimizer[I]) assessPrompt(ctx context.Context, prompt *I, assessment *PromptAssessment) error {
	err := po.assessmentAgent.Run(ctx, prompt, assessment, nil)
	if err != nil {
		return err
	}
	// Normalize grading for consistency
	assessment.OverallGrade, err = normalizeGrade(assessment.OverallGrade, assessment.OverallScore)
	if err != nil {
		return fmt.Errorf("invalid overall grade: %w", err)
	}
	return nil
}

// isOptimizationGoalMet determines if a prompt's assessment meets the configured
// optimization threshold. It supports both numerical and letter-based grading systems.
//
// For numerical ratings:
// - Uses a 0-20 scale
// - Compares against the configured threshold
//
// For letter grades:
// - Converts letter grades to GPA scale (0.0-4.3)
// - Requires A- (3.7) or better to meet goal
//
// Parameters:
//   - assessment: The PromptAssessment to evaluate
//
// Returns:
//   - bool: true if optimization goal is met
//   - error: if rating system is invalid or grade cannot be evaluated
//
// Example threshold values:
//   - Numerical: 0.75 requires score >= 15/20
//   - Letter: Requires A- or better
func (po *PromptOptimizer[T]) isOptimizationGoalMet(assessment PromptAssessment) (bool, error) {
	if po.ratingSystem == "" {
		return false, nil
	}

	switch po.ratingSystem {
	case "numerical":
		return assessment.OverallScore >= 20*po.threshold, nil
	case "letter":
		gradeValues := map[string]float64{
			"A+": 4.3, "A": 4.0, "A-": 3.7,
			"B+": 3.3, "B": 3.0, "B-": 2.7,
			"C+": 2.3, "C": 2.0, "C-": 1.7,
			"D+": 1.3, "D": 1.0, "D-": 0.7,
			"F": 0.0,
		}
		gradeValue, exists := gradeValues[assessment.OverallGrade]
		if !exists {
			return false, fmt.Errorf("invalid grade: %s", assessment.OverallGrade)
		}
		return gradeValue >= 3.7, nil // Equivalent to A- or better
	default:
		return false, fmt.Errorf("unknown rating system: %s", po.ratingSystem)
	}
}
