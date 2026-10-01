package domain

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrAlreadyEvaluated     = errors.New("already evaluated this week")
	ErrSelfEvaluation       = errors.New("cannot evaluate yourself")
	ErrInvalidAnswers       = errors.New("invalid answers")
	ErrOutOfHierarchy       = errors.New("evaluated employee is not in your hierarchy")
)
