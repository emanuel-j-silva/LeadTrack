package domain

type Employee struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PositionName string `json:"positionName"`
}

type Question struct {
	ID     int    `json:"id"`
	Label  string `json:"label"`
	Weight int    `json:"weight"`
}

type AnswerInput struct {
	QuestionID int `json:"questionId"`
	Value      int `json:"value"`
}

type EvaluationInput struct {
	Answers []AnswerInput `json:"answers"`
}

type Evaluation struct {
	ID          int      `json:"id"`
	EvaluatorID int      `json:"evaluatorId"`
	EvaluatedID int      `json:"evaluatedId"`
	WeekStart   string   `json:"weekStart"`
	Score       float64  `json:"score"`
	CreatedAt   string   `json:"createdAt"`
	Answers     []Answer `json:"answers,omitempty"`
}

type Answer struct {
	QuestionID int `json:"questionId"`
	Value      int `json:"value"`
}

type Subordinate struct {
	ID               int         `json:"id"`
	Name             string      `json:"name"`
	Email            string      `json:"email"`
	PositionName     string      `json:"positionName"`
	Depth            int         `json:"depth"`
	CanEvaluate      bool        `json:"canEvaluateThisWeek"`
	LatestEvaluation *Evaluation `json:"latestEvaluation,omitempty"`
}
