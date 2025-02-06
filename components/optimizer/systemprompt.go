package optimizer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/components/systemprompt/broke"
	"github.com/bububa/atomic-agents/schema"
)

type HistoryGetter interface {
	RecentHistory() []OptimizationEntry
}

type ContextProvider struct {
	historyGetter HistoryGetter
}

func NewContextProvider(historyGetter HistoryGetter) *ContextProvider {
	return &ContextProvider{
		historyGetter: historyGetter,
	}
}

func (c *ContextProvider) Title() string {
	return "RecentHistory"
}

func (c *ContextProvider) Info() string {
	list := c.historyGetter.RecentHistory()
	parts := make([]string, 0, len(list))
	innerParts := make([]string, 0, 3)
	for idx, v := range list {
		innerParts = append(innerParts, fmt.Sprintf("  %d. Prompt and Assessment", idx+1))
		innerParts = append(innerParts, fmt.Sprintf("    - Prompt: %s", schema.Stringify(v.Prompt)))
		innerParts = append(innerParts, fmt.Sprintf("    - Assessment: %s", schema.Stringify(v.Assessment)))
		parts = append(parts, strings.Join(innerParts, "\n"))
		innerParts = innerParts[0:]
	}
	return strings.Join(parts, "\n")
}

func (po *PromptOptimizer[T]) assessmentPromptGenerator() systemprompt.Generator {
	bs, _ := json.Marshal(po.customMetrics)
	metrics := string(bs)
	return broke.New(broke.WithBackground([]string{
		fmt.Sprintf("Assess the prompt and evaluates a prompt's quality and effectiveness for task: %s", po.taskDesc),
		fmt.Sprintf("- Custom Metrics: %s", metrics),
		fmt.Sprintf("- Optimization Goal: %s", po.optimizationGoal),
	}),
		broke.WithRoles([]string{
			"- You are a LLM prompt engineer.",
			"- You are good at LLM prompt writing and analysis.",
		}),
		broke.WithObjectives([]string{
			"- constructs an evaluation prompt incorporating task description and history.",
			"- evaluates a prompt's quality and effectiveness",
			"- performs a comprehensive analysis considering multiple factors including custom metrics, optimization goals, and historical context.",
		}),
		broke.WithKeyResults([]string{
			"- custom metrics with name, value, reasoning",
			"- prompt strengths point with examples",
			"- weaknesses point with improvement suggestions",
			"- overall effectiveness and efficiency",
			"- Alignment with optimization goals",
		}),
		broke.WithEvolves([]string{
			"- Do not use any markdown formatting, code blocks, or backticks in your response.",
			"- Return only the raw JSON object.",
			"- For numerical ratings, use a scale of 0 to 20 (inclusive).",
			"- For the overallGrade:",
			"  - If using letter grades, use one of: F, D-, D, D+, C-, C, C+, B-, B, B+, A-, A, A+",
			"  - If using numerical grades, use the same value as overallScore (0-20)",
			"- Include at least one item in each array (metrics, strengths, weaknesses, suggestions).",
			"- Provide specific examples and reasoning for each point.",
			"- Rate the prompt's efficiency and alignment with the optimization goal.",
			"- Rank suggestions by their expected impact (20 being highest impact).",
			"- Use clear, jargon-free language in your assessment.",
			"- Double-check that your response is valid JSON before submitting.",
		}),
		broke.WithContextProviders(NewContextProvider(po)),
	)
}

func (po *PromptOptimizer[T]) improvementPromptGenerator() systemprompt.Generator {
	return broke.New(broke.WithBackground([]string{
		"based on the following assessment and recent history, generate an improved version of the entire prompt structure",
		fmt.Sprintf("- Task Description: %s", po.taskDesc),
		fmt.Sprintf("- Optimization Goal: %s", po.optimizationGoal),
	}),
		broke.WithRoles([]string{
			"- You are a LLM prompt engineer.",
			"- You are good at LLM prompt writing and analysis.",
		}),
		broke.WithObjectives([]string{
			"- analyzes previous assessment and optimization history",
			"- generates two alternative improvements.",
			"  - incremental: Refines existing approach",
			"  - bold: Reimagines prompt structure",
			"- evaluates expected impact of each version",
		}),
		broke.WithKeyResults([]string{
			"- An incremental improvement",
			"- A bold reimagining",
		}),
		broke.WithEvolves([]string{
			"- Directly address weaknesses identified in the assessment.",
			"- Build upon identified strengths.",
			"- Ensure alignment with the task description and optimization goal.",
			"- Strive for efficiency in language use.",
			"- Use clear, jargon-free language.",
			"- Provide a brief reasoning for major changes.",
			"- Rate the expected impact of each version on a scale of 0 to 20.",
			"- Double-check that your response is valid JSON before submitting.",
		}),
		broke.WithContextProviders(NewContextProvider(po)),
	)
}
