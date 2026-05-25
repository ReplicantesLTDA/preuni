// Package adapters provides in-process implementations of the auth-svc port
// interfaces so the monolith can skip self-HTTP loopback for internal calls.
package adapters

import (
	"context"

	"github.com/preuni/app/internal/auth/ports"
	userrepo "github.com/preuni/app/internal/user/repository"
)

// InProcessStudentProvisioner persists a new student row directly via the
// user-svc repository, bypassing the HTTP boundary.
type InProcessStudentProvisioner struct {
	Repo *userrepo.StudentRepository
}

// NewInProcessStudentProvisioner returns a provisioner backed by repo.
func NewInProcessStudentProvisioner(repo *userrepo.StudentRepository) *InProcessStudentProvisioner {
	return &InProcessStudentProvisioner{Repo: repo}
}

// CreateStudent implements ports.StudentProvisioner.
func (a *InProcessStudentProvisioner) CreateStudent(ctx context.Context, req ports.CreateStudentRequest) error {
	username := req.Username
	if username == "" {
		username = defaultUsername(req.StudentID)
	}
	return a.Repo.Create(ctx, &userrepo.Student{
		ID:          req.StudentID,
		DisplayName: req.DisplayName,
		Username:    username,
		Email:       req.Email,
	})
}

func defaultUsername(id string) string {
	if len(id) >= 8 {
		return "user" + id[:8]
	}
	return "user" + id
}
