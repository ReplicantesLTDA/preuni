// Package repository is the social domain's persistence layer: friend
// requests/connections and the visibility-gated queries that expose a
// friend's streak and latest grade (spec FR-004).
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"

	"github.com/preuni/app/internal/social/domain"
)

// Friend is one accepted friend's public-to-friends profile: streak and
// latest essay grade.
type Friend struct {
	FriendshipID       string
	UserID             string
	DisplayName        string
	CurrentStreak      int
	LatestOverallScore *int
	LatestGradedAt     *string
}

// Repository is the social domain's persistence layer.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// SendRequest creates a pending friend request from requesterID to
// addresseeID, unless a pending or accepted relationship already exists
// between them (domain.CanSendRequest — a prior removal doesn't block a
// new request, per spec Edge Cases).
func (r *Repository) SendRequest(ctx context.Context, requesterID, addresseeID string) (string, error) {
	if requesterID == addresseeID {
		return "", apperrors.Validation("addressee_id", "cannot friend yourself")
	}

	existing, err := r.latestStatus(ctx, requesterID, addresseeID)
	if err != nil {
		return "", err
	}
	if !domain.CanSendRequest(existing) {
		return "", apperrors.Conflict("a friend request already exists between these users")
	}

	id := uuid.New().String()
	_, err = r.db.Exec(ctx, `
		INSERT INTO social.friendships (id, requester_id, addressee_id, status, created_at)
		VALUES ($1, $2, $3, 'pending', now())
	`, id, requesterID, addresseeID)
	if err != nil {
		return "", apperrors.Internal(err)
	}
	return id, nil
}

// AcceptRequest accepts a pending request, but only if callerID is the
// addressee (you can't accept your own outgoing request).
func (r *Repository) AcceptRequest(ctx context.Context, friendshipID, callerID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE social.friendships
		SET status = 'accepted', responded_at = now()
		WHERE id = $1 AND addressee_id = $2 AND status = 'pending'
	`, friendshipID, callerID)
	if err != nil {
		return apperrors.Internal(err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.NotFound("friend request")
	}
	return nil
}

// RemoveFriend marks an accepted (or pending) friendship removed, for
// either party.
func (r *Repository) RemoveFriend(ctx context.Context, friendshipID, callerID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE social.friendships
		SET status = 'removed', responded_at = now()
		WHERE id = $1 AND (requester_id = $2 OR addressee_id = $2) AND status <> 'removed'
	`, friendshipID, callerID)
	if err != nil {
		return apperrors.Internal(err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.NotFound("friendship")
	}
	return nil
}

// ListFriends returns userID's accepted friends with their current streak
// and latest grade -- the visibility-gated view (spec FR-004): only
// accepted friendships ever reach this query.
func (r *Repository) ListFriends(ctx context.Context, userID string) ([]Friend, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.id, s.id, s.display_name, s.streak_count,
		       eg.overall_score, eg.graded_at
		FROM social.friendships f
		JOIN users.students s ON s.id = CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END
		LEFT JOIN LATERAL (
			SELECT eg.overall_score, eg.graded_at
			FROM essay.essay_grades eg
			JOIN essay.essay_submissions es ON es.id = eg.submission_id
			WHERE es.user_id = s.id
			ORDER BY eg.graded_at DESC
			LIMIT 1
		) eg ON true
		WHERE f.status = 'accepted' AND (f.requester_id = $1 OR f.addressee_id = $1)
		ORDER BY s.display_name
	`, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var f Friend
		var gradedAt *time.Time
		if err := rows.Scan(&f.FriendshipID, &f.UserID, &f.DisplayName, &f.CurrentStreak, &f.LatestOverallScore, &gradedAt); err != nil {
			return nil, apperrors.Internal(err)
		}
		if gradedAt != nil {
			s := gradedAt.Format(time.RFC3339)
			f.LatestGradedAt = &s
		}
		friends = append(friends, f)
	}
	return friends, nil
}

// latestStatus returns the most recent friendship row's status between the
// two users (either direction), or nil if none exists.
func (r *Repository) latestStatus(ctx context.Context, userA, userB string) (*domain.Status, error) {
	row := r.db.QueryRow(ctx, `
		SELECT status FROM social.friendships
		WHERE (requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1)
		ORDER BY created_at DESC
		LIMIT 1
	`, userA, userB)

	var s string
	if err := row.Scan(&s); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, apperrors.Internal(err)
	}
	status := domain.Status(s)
	return &status, nil
}
