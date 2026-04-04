package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"
)

// Student mirrors the users.students row.
type Student struct {
	ID                  string
	DisplayName         string
	Username            string
	Email               string
	AvatarURL           *string
	XPTotal             int64
	StreakCount         int
	ReadinessScore      float64
	OnboardingCompleted bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// StudentPatch carries optional fields for an update; nil = no-op for that field.
type StudentPatch struct {
	DisplayName *string
	Username    *string
	AvatarURL   *string
}

// StudentRepository handles persistence for users.students.
type StudentRepository struct {
	db *pgxpool.Pool
}

// NewStudentRepository creates a StudentRepository backed by the given pool.
func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{db: db}
}

// Create inserts a new student row.
func (r *StudentRepository) Create(ctx context.Context, s *Student) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users.students (id, display_name, username, email, created_at, updated_at)
		VALUES ($1, $2, $3, lower($4), now(), now())
	`, s.ID, s.DisplayName, s.Username, s.Email)
	if err != nil {
		if isDuplicateKeyError(err) {
			return apperrors.Conflict("username is already taken")
		}
		return apperrors.Internal(err)
	}
	return nil
}

// FindByID retrieves a student by ID.
func (r *StudentRepository) FindByID(ctx context.Context, id string) (*Student, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, display_name, username, email, avatar_url,
		       xp_total, streak_count, readiness_score, onboarding_completed,
		       created_at, updated_at
		FROM users.students
		WHERE id = $1
	`, id)
	return scanStudent(row)
}

// Update applies non-nil patch fields to the student row.
func (r *StudentRepository) Update(ctx context.Context, id string, patch *StudentPatch) (*Student, error) {
	// Build dynamic update
	setParts := []string{"updated_at = now()"}
	args := []any{id}
	argIdx := 2

	if patch.DisplayName != nil {
		setParts = append(setParts, "display_name = $"+itoa(argIdx))
		args = append(args, *patch.DisplayName)
		argIdx++
	}
	if patch.Username != nil {
		setParts = append(setParts, "username = $"+itoa(argIdx))
		args = append(args, *patch.Username)
		argIdx++
	}
	if patch.AvatarURL != nil {
		setParts = append(setParts, "avatar_url = $"+itoa(argIdx))
		args = append(args, *patch.AvatarURL)
		argIdx++
	}

	query := "UPDATE users.students SET " + join(setParts, ", ") + " WHERE id = $1"
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, apperrors.Conflict("username is already taken")
		}
		return nil, apperrors.Internal(err)
	}
	return r.FindByID(ctx, id)
}

// Anonymize replaces PII with non-identifying values (GDPR right-to-erasure).
func (r *StudentRepository) Anonymize(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users.students
		SET display_name = 'Deleted User',
		    avatar_url   = NULL,
		    xp_total     = 0,
		    streak_count = 0,
		    updated_at   = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// SetOnboardingCompleted marks onboarding_completed = true for the given student.
func (r *StudentRepository) SetOnboardingCompleted(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users.students SET onboarding_completed = true, updated_at = now() WHERE id = $1
	`, id)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

func scanStudent(row pgx.Row) (*Student, error) {
	var s Student
	err := row.Scan(
		&s.ID, &s.DisplayName, &s.Username, &s.Email, &s.AvatarURL,
		&s.XPTotal, &s.StreakCount, &s.ReadinessScore, &s.OnboardingCompleted,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("student")
		}
		return nil, apperrors.Internal(err)
	}
	return &s, nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i >= 0
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 10)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
