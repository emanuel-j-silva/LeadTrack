package service

import (
	"context"
	"sort"
	"time"

	"leadtrack/backend/internal/domain"
	"leadtrack/backend/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListEmployees(ctx context.Context) ([]domain.Employee, error) {
	return s.repo.ListEmployees(ctx)
}

func (s *Service) GetMe(ctx context.Context, employeeID int) (*domain.Employee, error) {
	return s.repo.GetEmployeeByID(ctx, employeeID)
}

func (s *Service) ListQuestions(ctx context.Context) ([]domain.Question, error) {
	return s.repo.ListQuestions(ctx)
}

func (s *Service) ListSubordinates(ctx context.Context, leaderID int, loc *time.Location) ([]domain.Subordinate, error) {
	subIDs, err := s.repo.GetSubordinateIDs(ctx, leaderID)
	if err != nil {
		return nil, err
	}

	if len(subIDs) == 0 {
		return []domain.Subordinate{}, nil
	}

	var subs []domain.Subordinate
	now := time.Now()
	weekStart := domain.WeekStart(now, loc)

	for id, depth := range subIDs {
		emp, err := s.repo.GetEmployeeByID(ctx, id)
		if err != nil {
			continue
		}

		hasEvaluated, err := s.repo.HasEvaluatedThisWeek(ctx, leaderID, id, weekStart)
		if err != nil {
			hasEvaluated = false
		}

		latestEval, err := s.repo.GetLatestEvaluationForEvaluated(ctx, leaderID, id)
		if err != nil && err != domain.ErrNotFound {
			// ignore or log
		}

		subs = append(subs, domain.Subordinate{
			ID:               emp.ID,
			Name:             emp.Name,
			Email:            emp.Email,
			PositionName:     emp.PositionName,
			Depth:            depth,
			CanEvaluate:      !hasEvaluated,
			LatestEvaluation: latestEval,
		})
	}

	sort.Slice(subs, func(i, j int) bool {
		if subs[i].Depth != subs[j].Depth {
			return subs[i].Depth < subs[j].Depth
		}
		return subs[i].ID < subs[j].ID
	})

	return subs, nil
}

func (s *Service) CreateEvaluation(ctx context.Context, leaderID, evaluatedID int, input domain.EvaluationInput, loc *time.Location) (int, error) {
	if leaderID == evaluatedID {
		return 0, domain.ErrSelfEvaluation
	}

	subIDs, err := s.repo.GetSubordinateIDs(ctx, leaderID)
	if err != nil {
		return 0, err
	}

	if _, ok := subIDs[evaluatedID]; !ok {
		return 0, domain.ErrOutOfHierarchy
	}

	now := time.Now()
	weekStart := domain.WeekStart(now, loc)

	hasEvaluated, err := s.repo.HasEvaluatedThisWeek(ctx, leaderID, evaluatedID, weekStart)
	if err != nil {
		return 0, err
	}
	if hasEvaluated {
		return 0, domain.ErrAlreadyEvaluated
	}

	questions, err := s.repo.ListQuestions(ctx)
	if err != nil {
		return 0, err
	}

	qMap := make(map[int]int)
	for _, q := range questions {
		qMap[q.ID] = q.Weight
	}

	score, err := domain.CalculateScore(input.Answers, qMap)
	if err != nil {
		return 0, domain.ErrInvalidAnswers
	}

	evalID, err := s.repo.CreateEvaluation(ctx, leaderID, evaluatedID, weekStart, score, input.Answers)
	if err != nil {
		return 0, err
	}

	return evalID, nil
}

func (s *Service) GetEvaluations(ctx context.Context, currentUserID, evaluatedID int) ([]domain.Evaluation, error) {
	subIDs, err := s.repo.GetSubordinateIDs(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	if evaluatedID != currentUserID {
		if _, ok := subIDs[evaluatedID]; !ok {
			return nil, domain.ErrForbidden
		}
	} else {
		return nil, domain.ErrForbidden // R8: User never sees own evaluation
	}

	return s.repo.GetEvaluationsForEvaluated(ctx, currentUserID, evaluatedID)
}

func (s *Service) GetLatestEvaluation(ctx context.Context, currentUserID, evaluatedID int) (*domain.Evaluation, error) {
	subIDs, err := s.repo.GetSubordinateIDs(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	if evaluatedID != currentUserID {
		if _, ok := subIDs[evaluatedID]; !ok {
			return nil, domain.ErrForbidden
		}
	} else {
		return nil, domain.ErrForbidden // R8
	}

	eval, err := s.repo.GetLatestEvaluationForEvaluated(ctx, currentUserID, evaluatedID)
	if err != nil {
		return nil, err
	}
	return eval, nil
}
