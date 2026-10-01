package repository

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"leadtrack/backend/internal/domain"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	// Sort migration files by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (r *Repository) GetEmployeeByID(ctx context.Context, id int) (*domain.Employee, error) {
	query := `SELECT id, name, email, position_name FROM employee WHERE id = $1`
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, id).Scan(&emp.ID, &emp.Name, &emp.Email, &emp.PositionName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

func (r *Repository) ListEmployees(ctx context.Context) ([]domain.Employee, error) {
	query := `SELECT id, name, email, position_name FROM employee ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []domain.Employee
	for rows.Next() {
		var emp domain.Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.PositionName); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}

func (r *Repository) ListQuestions(ctx context.Context) ([]domain.Question, error) {
	query := `SELECT id, label, weight FROM question ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.Question
	for rows.Next() {
		var q domain.Question
		if err := rows.Scan(&q.ID, &q.Label, &q.Weight); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *Repository) GetSubordinateIDs(ctx context.Context, leaderID int) (map[int]int, error) {
	query := `
		WITH RECURSIVE subs(id, depth) AS (
			SELECT lead_id, 1 FROM leader_lead WHERE leader_id = $1
			UNION
			SELECT ll.lead_id, s.depth + 1
			FROM leader_lead ll JOIN subs s ON ll.leader_id = s.id
		)
		SELECT id, MIN(depth) AS depth FROM subs GROUP BY id;
	`
	rows, err := r.pool.Query(ctx, query, leaderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := make(map[int]int)
	for rows.Next() {
		var id, depth int
		if err := rows.Scan(&id, &depth); err != nil {
			return nil, err
		}
		subs[id] = depth
	}
	return subs, nil
}

func (r *Repository) HasEvaluatedThisWeek(ctx context.Context, evaluatorID, evaluatedID int, weekStart time.Time) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM evaluation WHERE evaluator_id = $1 AND evaluated_id = $2 AND week_start = $3)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, evaluatorID, evaluatedID, weekStart).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateEvaluation(ctx context.Context, evaluatorID, evaluatedID int, weekStart time.Time, score float64, answers []domain.AnswerInput) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var evalID int
	evalQuery := `
		INSERT INTO evaluation (evaluator_id, evaluated_id, week_start, score)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = tx.QueryRow(ctx, evalQuery, evaluatorID, evaluatedID, weekStart, score).Scan(&evalID)
	if err != nil {
		return 0, err
	}

	for _, ans := range answers {
		ansQuery := `
			INSERT INTO evaluation_answer (evaluation_id, question_id, value)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(ctx, ansQuery, evalID, ans.QuestionID, ans.Value)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return evalID, nil
}

func (r *Repository) GetLatestEvaluationForEvaluated(ctx context.Context, currentUserID, evaluatedID int) (*domain.Evaluation, error) {
	// R8: Visibility rule: User can only see evaluations of subordinates where the evaluator is the user or someone in the user's subtree.
	// But wait, what is the exact visibility rule for reading evaluations?
	// Let's check README R8 and premise #2:
	// "O usuário nunca vê a própria avaliação, nem de pares ou superiores. Só vê avaliações de subordinados, feitas por ele ou por subordinados dele"
	// "O usuário só enxerga avaliações cujo avaliado está na sua subárvore e cujo avaliador é ele mesmo ou alguém da sua subárvore."
	query := `
		SELECT e.id, e.evaluator_id, e.evaluated_id, e.week_start, e.score, e.created_at
		FROM evaluation e
		WHERE e.evaluated_id = $2
		  AND e.evaluated_id IN (
			WITH RECURSIVE subs(id) AS (
				SELECT lead_id FROM leader_lead WHERE leader_id = $1
				UNION
				SELECT ll.lead_id FROM leader_lead ll JOIN subs s ON ll.leader_id = s.id
			)
			SELECT id FROM subs
		  )
		  AND (
			e.evaluator_id = $1
			OR e.evaluator_id IN (
				WITH RECURSIVE subs(id) AS (
					SELECT lead_id FROM leader_lead WHERE leader_id = $1
					UNION
					SELECT ll.lead_id FROM leader_lead ll JOIN subs s ON ll.leader_id = s.id
				)
				SELECT id FROM subs
			)
		  )
		ORDER BY e.created_at DESC
		LIMIT 1;
	`
	var eval domain.Evaluation
	var weekStart time.Time
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, query, currentUserID, evaluatedID).Scan(
		&eval.ID, &eval.EvaluatorID, &eval.EvaluatedID, &weekStart, &eval.Score, &createdAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	eval.WeekStart = weekStart.Format("2006-01-02")
	eval.CreatedAt = createdAt.Format(time.RFC3339)

	// Fetch answers for this evaluation
	ansQuery := `SELECT question_id, value FROM evaluation_answer WHERE evaluation_id = $1 ORDER BY question_id`
	rows, err := r.pool.Query(ctx, ansQuery, eval.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ans domain.Answer
		if err := rows.Scan(&ans.QuestionID, &ans.Value); err != nil {
			return nil, err
		}
		eval.Answers = append(eval.Answers, ans)
	}

	return &eval, nil
}

func (r *Repository) GetEvaluationsForEvaluated(ctx context.Context, currentUserID, evaluatedID int) ([]domain.Evaluation, error) {
	query := `
		SELECT e.id, e.evaluator_id, e.evaluated_id, e.week_start, e.score, e.created_at
		FROM evaluation e
		WHERE e.evaluated_id = $2
		  AND e.evaluated_id IN (
			WITH RECURSIVE subs(id) AS (
				SELECT lead_id FROM leader_lead WHERE leader_id = $1
				UNION
				SELECT ll.lead_id FROM leader_lead ll JOIN subs s ON ll.leader_id = s.id
			)
			SELECT id FROM subs
		  )
		  AND (
			e.evaluator_id = $1
			OR e.evaluator_id IN (
				WITH RECURSIVE subs(id) AS (
					SELECT lead_id FROM leader_lead WHERE leader_id = $1
					UNION
					SELECT ll.lead_id FROM leader_lead ll JOIN subs s ON ll.leader_id = s.id
				)
				SELECT id FROM subs
			)
		  )
		ORDER BY e.created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, currentUserID, evaluatedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evals []domain.Evaluation
	for rows.Next() {
		var eval domain.Evaluation
		var weekStart time.Time
		var createdAt time.Time
		if err := rows.Scan(&eval.ID, &eval.EvaluatorID, &eval.EvaluatedID, &weekStart, &eval.Score, &createdAt); err != nil {
			return nil, err
		}
		eval.WeekStart = weekStart.Format("2006-01-02")
		eval.CreatedAt = createdAt.Format(time.RFC3339)
		evals = append(evals, eval)
	}
	return evals, nil
}
