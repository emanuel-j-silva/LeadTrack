package domain

import (
	"fmt"
	"math"
)

// CalculateScore computes the weighted average score (1.00 to 4.00) based on answers and question weights.
func CalculateScore(answers []AnswerInput, questions map[int]int) (float64, error) {
	if len(answers) != 6 {
		return 0, fmt.Errorf("must have exactly 6 answers, got %d", len(answers))
	}

	var totalWeightedScore float64
	var totalWeight int

	seenQuestions := make(map[int]bool)

	for _, ans := range answers {
		if ans.Value < 1 || ans.Value > 4 {
			return 0, fmt.Errorf("answer value %d out of range [1, 4]", ans.Value)
		}
		if seenQuestions[ans.QuestionID] {
			return 0, fmt.Errorf("duplicate answer for question id %d", ans.QuestionID)
		}
		seenQuestions[ans.QuestionID] = true

		weight, ok := questions[ans.QuestionID]
		if !ok {
			return 0, fmt.Errorf("question id %d does not exist", ans.QuestionID)
		}

		totalWeightedScore += float64(ans.Value * weight)
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0, fmt.Errorf("total weight cannot be zero")
	}

	score := totalWeightedScore / float64(totalWeight)
	// Round to 2 decimal places
	score = math.Round(score*100) / 100

	if score < 1.00 {
		score = 1.00
	}
	if score > 4.00 {
		score = 4.00
	}

	return score, nil
}
