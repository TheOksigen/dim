package results

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("exam result not found")

const lookupSQL = "SELECT fin_code, first_name, last_name, father_name, exam_date, " +
	"subject_1_name, subject_1_score, subject_1_answers, " +
	"subject_2_name, subject_2_score, subject_2_answers, " +
	"subject_3_name, subject_3_score, subject_3_answers, " +
	"total_score, passing_status FROM exam_results WHERE fin_code = $1"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindByFIN(ctx context.Context, fin string) (Result, error) {
	var (
		result   Result
		examDate time.Time
		subject1 Subject
		subject2 Subject
		subject3 Subject
	)

	err := r.pool.QueryRow(ctx, lookupSQL, fin).Scan(
		&result.FINCode,
		&result.FirstName,
		&result.LastName,
		&result.FatherName,
		&examDate,
		&subject1.Name,
		&subject1.Score,
		&subject1.Answers,
		&subject2.Name,
		&subject2.Score,
		&subject2.Answers,
		&subject3.Name,
		&subject3.Score,
		&subject3.Answers,
		&result.TotalScore,
		&result.Passed,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}

	result.ExamDate = examDate.Format("2006-01-02")
	result.Subjects = []Subject{subject1, subject2, subject3}
	return result, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
