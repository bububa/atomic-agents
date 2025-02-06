package optimizer

import "github.com/bububa/atomic-agents/schema"

// OptimizationEntry represents a single step in the optimization process,
// containing both the prompt and its assessment.
type OptimizationEntry struct {
	schema.Base
	// Prompt is the LLM prompt being evaluated
	Prompt schema.Schema `json:"prompt" jsonschema:"title=prompt,description=previous evaluated prompt"`

	// Assessment contains the comprehensive evaluation of the prompt
	Assessment PromptAssessment `json:"assessment" jsonschema:"title=assessment,description=the comprehensive evaluation of the prompt"`
}

// Metric represents a quantitative or qualitative measure of prompt performance.
// Each metric provides a specific aspect of evaluation with supporting reasoning.
type Metric struct {
	// Name identifies the metric (e.g., "Clarity", "Specificity")
	Name string `json:"name" jsonschema:"title=name,description=identifies the metric (e.g., 'Clarity', 'Specificity')"`

	// Description explains what the metric measures and its significance
	Description string `json:"description" jsonschema:"title=description,description=explains what the metric measures and its significance"`

	// Value is the numerical score (0-20 scale) assigned to this metric
	Value float64 `json:"value" validate:"required,min=0,max=20" jsonschema:"title=value,description=Value is the numerical score (0-20 scale) assigned to this metric"`

	// Reasoning provides the rationale behind the assigned value
	Reasoning string `json:"reasoning" jsonschema:"title=reasoning,descrption=provides the rationale behind the assigned value"`
}

// Strength represents a positive aspect of the prompt with a concrete example.
type Strength struct {
	// Point describes the strength (e.g., "Clear task definition")
	Point string `json:"point" jsonschema:"title=point,description=describes the strength (e.g., 'Clear task definition')"`

	// Example provides a specific instance demonstrating this strength
	Example string `json:"example" jsonschema:"title=example,description=provides a specific instance demonstrating this strength"`
}

// Weakness represents an area for improvement in the prompt with a concrete example.
type Weakness struct {
	// Point describes the weakness (e.g., "Ambiguous constraints")
	Point string `json:"point" jsonschema:"title=point,description=describes the weakness (e.g., 'Ambiguous constraints')"`

	// Example provides a specific instance demonstrating this weakness
	Example string `json:"example" jsonschema:"title=example,description=provides a specific instance demonstrating this weakness"`
}

// Suggestion represents a proposed improvement to the prompt with impact estimation.
type Suggestion struct {
	// Description outlines the suggested change
	Description string `json:"description" jsonschema:"title=description,description=outlines the suggested change"`

	// ExpectedImpact estimates the improvement's effect (0-20 scale)
	ExpectedImpact float64 `json:"expectedImpact" validate:"required,min=0,max=20" jsonschema:"title=expectedImpact,description=estimates the improvement's effect (0-20 scale)"`

	// Reasoning explains why this suggestion would improve the prompt
	Reasoning string `json:"reasoning" jsonschema:"title=reasoning,description=explains why this suggestion would improve the prompt"`
}

// PromptAssessment provides a comprehensive evaluation of a prompt's quality
// including metrics, strengths, weaknesses, and suggestions for improvement.
type PromptAssessment struct {
	schema.Base
	// Metrics contains specific performance measurements
	Metrics []Metric `json:"metrics" validate:"required,min=1" jsonschema:"title=metrics,description=specific performance measurements"`

	// Strengths lists positive aspects of the prompt
	Strengths []Strength `json:"strengths" validate:"required,min=1" jsonschema:"title=strengths,description=lists positive aspects of the prompt"`

	// Weaknesses identifies areas needing improvement
	Weaknesses []Weakness `json:"weaknesses" validate:"required,min=1" jsonschema:"title=weaknesses,description=identifies areas needing improvement"`

	// Suggestions provides actionable improvements
	Suggestions []Suggestion `json:"suggestions" validate:"required,min=1" jsonschema:"title=suggestions,description=provides actionable improvements"`

	// OverallScore represents the prompt's overall quality (0-20 scale)
	OverallScore float64 `json:"overallScore" validate:"required,min=0,max=20" jsonschema:"title=overallScore,description=represents the prompt's overall quality (0-20 scale)"`

	// OverallGrade provides a letter grade assessment (e.g., "A", "B+")
	OverallGrade string `json:"overallGrade" validate:"required,validGrade" jsonschema:"title=orerallGrade,enum=A+,enum=A,enum=A-,enum=B+,enum=B,enum=B-,enum=C+,enum=C,enum=C-,enum=D+,enum=D,enum=D-,enum=F,description=provides a letter grade assessment (e.g., 'A', 'B+')"`

	// EfficiencyScore measures token usage and processing efficiency
	EfficiencyScore float64 `json:"efficiencyScore" validate:"required,min=0,max=20" jsonschema:"title=efficiencyScore,description=measures token usage and processing efficiency"`

	// AlignmentWithGoal measures how well the prompt serves its intended purpose
	AlignmentWithGoal float64 `json:"alignmentWithGoal" validate:"required,min=0,max=20" jsonschema:"title=alignmentWithGoal,description=measures how well the prompt serves its intended purpose (0-20 scale)"`
}

// PromptImprovement represents improvement for the prompt
type PromptImprovement[T schema.Schema] struct {
	schema.Base
	IncrementalImprovement *T `json:"incrementalImprovement" jsonschema:"title=incrementalImprovement,description=Refines existing prompt approach"`
	BoldRedesign           *T `json:"boldRedesign" jsonschema:"title=boldRedesign,description=Reimagines prompt structure"`
	ExpectedImpact         struct {
		Incremental float64 `json:"incremental" jsonschema:"title=incremental,description=incrementalImprovement version impact score"`
		Bold        float64 `json:"bold" jsonschema:"title=bold,description=boldRedesign version impact score"`
	} `json:"expectedImpact" jsonschema:"title=expectedImpact,description=Evaluates expected impact of each version on a scale of 0 to 20"`
}
