package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateScore(t *testing.T) {
	questions := map[int]int{
		1: 25,
		2: 20,
		3: 20,
		4: 15,
		5: 10,
		6: 10,
	} // Total weight = 100

	answers := []AnswerInput{
		{QuestionID: 1, Value: 4}, // 100
		{QuestionID: 2, Value: 3}, // 60
		{QuestionID: 3, Value: 3}, // 60
		{QuestionID: 4, Value: 3}, // 45
		{QuestionID: 5, Value: 3}, // 30
		{QuestionID: 6, Value: 3}, // 30
	} // Total weighted score = 325 / 100 = 3.25

	score, err := CalculateScore(answers, questions)
	assert.NoError(t, err)
	assert.Equal(t, 3.25, score)
}
