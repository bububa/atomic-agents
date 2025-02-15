package optimizer

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// OptimizationRating defines the interface for different rating systems used in prompt optimization.
// Implementations can provide different ways to evaluate if an optimization goal has been met.
type OptimizationRating interface {
	// IsGoalMet determines if the optimization goal has been achieved
	IsGoalMet() bool
	// String returns a string representation of the rating
	String() string
}

// NumericalRating implements OptimizationRating using a numerical score system.
// It evaluates prompts on a scale from 0 to Max.
type NumericalRating struct {
	Score float64 // Current score
	Max   float64 // Maximum possible score
}

// IsGoalMet checks if the numerical score meets the optimization goal.
// Returns true if the score is 90% or higher of the maximum possible score.
func (nr NumericalRating) IsGoalMet() bool {
	return nr.Score >= nr.Max*0.9 // Consider goal met if score is 90% or higher
}

// String formats the numerical rating as a string in the form "score/max".
func (nr NumericalRating) String() string {
	return fmt.Sprintf("%.1f/%.1f", nr.Score, nr.Max)
}

// LetterRating implements OptimizationRating using a letter grade system.
// It evaluates prompts using traditional academic grades (A+, A, B, etc.).
type LetterRating struct {
	Grade string // Letter grade (A+, A, B, etc.)
}

// IsGoalMet checks if the letter grade meets the optimization goal.
// Returns true for grades A, A+, or S.
func (lr LetterRating) IsGoalMet() bool {
	return lr.Grade == "A" || lr.Grade == "A+" || lr.Grade == "S"
}

// String returns the letter grade as a string.
func (lr LetterRating) String() string {
	return lr.Grade
}

// validGrade validates if a given grade string is a valid letter grade.
func validGrade(fl validator.FieldLevel) bool {
	grade := fl.Field().String()
	validGrades := map[string]bool{
		"A+": true, "A": true, "A-": true,
		"B+": true, "B": true, "B-": true,
		"C+": true, "C": true, "C-": true,
		"D+": true, "D": true, "D-": true,
		"F": true,
	}
	return validGrades[grade]
}

// normalizeGrade converts between numerical and letter grade formats.
// It ensures consistent grade representation across the optimization process.
//
// Conversion rules:
// - A+: >= 19/20 (95%)
// - A:  >= 17/20 (85%)
// - A-: >= 15/20 (75%)
// - B+: >= 13/20 (65%)
// - B:  >= 11/20 (55%)
// - B-: >= 9/20  (45%)
// - C+: >= 7/20  (35%)
// - C:  >= 5/20  (25%)
// - C-: >= 3/20  (15%)
// - D+: >= 2/20  (10%)
// - D:  >= 1/20  (5%)
// - F:  < 1/20   (<5%)
//
// Parameters:
//   - grade: Letter grade or numeric score string
//   - score: Numerical score (0-20 scale)
//
// Returns:
//   - Normalized letter grade
//   - Error if grade format is invalid
func normalizeGrade(grade string, score float64) (string, error) {
	validGrades := map[string]bool{
		"A+": true, "A": true, "A-": true,
		"B+": true, "B": true, "B-": true,
		"C+": true, "C": true, "C-": true,
		"D+": true, "D": true, "D-": true,
		"F": true,
	}

	if validGrades[grade] {
		return grade, nil
	}

	numericGrade, err := strconv.ParseFloat(grade, 64)
	if err != nil {
		return "", err
	}

	switch {
	case numericGrade >= 19:
		return "A+", nil // 95%+
	case numericGrade >= 17:
		return "A", nil // 85%+
	case numericGrade >= 15:
		return "A-", nil // 75%+
	case numericGrade >= 13:
		return "B+", nil // 65%+
	case numericGrade >= 11:
		return "B", nil // 55%+
	case numericGrade >= 9:
		return "B-", nil // 45%+
	case numericGrade >= 7:
		return "C+", nil // 35%+
	case numericGrade >= 5:
		return "C", nil // 25%+
	case numericGrade >= 3:
		return "C-", nil // 15%+
	case numericGrade >= 2:
		return "D+", nil // 10%+
	case numericGrade >= 1:
		return "D", nil // 5%+
	default:
		return "F", nil // <5%
	}
}
