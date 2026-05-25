package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/preuni/app/internal/user/handler"
	"github.com/preuni/app/internal/user/handler/testhelper"
	"github.com/preuni/app/internal/user/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_UpdateStudent_Unauthorized asserts that PATCH /students/me
// without a bearer token returns 401.
func TestIntegration_UpdateStudent_Unauthorized(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewUpdateStudentHandler(studentRepo))

	body, _ := json.Marshal(map[string]string{"display_name": "Test"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestIntegration_UpdateStudent_Success asserts that a valid patch returns 200
// with the updated student payload matching the full schema contract.
func TestIntegration_UpdateStudent_Success(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	const userID = "00000000-0000-0000-0000-bbb000000001"
	student := &repository.Student{
		ID:          userID,
		DisplayName: "Old Name",
		Username:    "olduser",
		Email:       "olduser@example.com",
	}
	require.NoError(t, repository.NewStudentRepository(pool).Create(ctx, student))

	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewUpdateStudentHandler(studentRepo))

	newDisplayName := "New Name"
	body, _ := json.Marshal(map[string]string{"display_name": newDisplayName})
	token := testhelper.MakeTestJWT(t, userID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp handler.StudentResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	// Contract: updated field is reflected and schema is complete
	assert.Equal(t, userID, resp.ID)
	assert.Equal(t, newDisplayName, resp.DisplayName)
	assert.Equal(t, "olduser", resp.Username)
	assert.Equal(t, "olduser@example.com", resp.Email)
	assert.GreaterOrEqual(t, resp.XPTotal, int64(0))
}

// TestIntegration_UpdateStudent_InvalidUsername asserts that an invalid username
// returns 422 with a validation envelope where field = "username".
func TestIntegration_UpdateStudent_InvalidUsername(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	const userID = "00000000-0000-0000-0000-bbb000000002"
	require.NoError(t, repository.NewStudentRepository(pool).Create(ctx, &repository.Student{
		ID:          userID,
		DisplayName: "Test",
		Username:    "testuser",
		Email:       "test@example.com",
	}))

	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewUpdateStudentHandler(studentRepo))

	// Username with uppercase is invalid per domain rules
	body, _ := json.Marshal(map[string]string{"username": "INVALID_USERNAME"})
	token := testhelper.MakeTestJWT(t, userID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	assert.Equal(t, "VALIDATION_ERROR", envelope.Error.Code)
	assert.Equal(t, "username", envelope.Error.Field)
}

// TestIntegration_UpdateStudent_ConflictUsername asserts that trying to claim
// a username already held by another student returns 409.
func TestIntegration_UpdateStudent_ConflictUsername(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	// Create two students
	const user1ID = "00000000-0000-0000-0000-bbb000000003"
	const user2ID = "00000000-0000-0000-0000-bbb000000004"
	studentRepo := repository.NewStudentRepository(pool)
	require.NoError(t, studentRepo.Create(ctx, &repository.Student{
		ID: user1ID, DisplayName: "User1", Username: "taken", Email: "u1@example.com",
	}))
	require.NoError(t, studentRepo.Create(ctx, &repository.Student{
		ID: user2ID, DisplayName: "User2", Username: "other", Email: "u2@example.com",
	}))

	h := withAuth(handler.NewUpdateStudentHandler(studentRepo))
	body, _ := json.Marshal(map[string]string{"username": "taken"})
	token := testhelper.MakeTestJWT(t, user2ID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}
