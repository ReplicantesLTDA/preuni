package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/preuni/app/internal/user/handler"
	"github.com/preuni/app/internal/user/handler/testhelper"
	"github.com/preuni/app/internal/user/repository"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withAuth wraps h in the RequireAuth middleware using the test signing key.
func withAuth(h http.Handler) http.Handler {
	return pkgmw.RequireAuth([]byte(testhelper.TestSigningKey))(h)
}

// TestIntegration_GetStudent_Unauthorized asserts that GET /students/me without
// a bearer token returns 401.
func TestIntegration_GetStudent_Unauthorized(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewGetStudentHandler(studentRepo))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/students/me", nil)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

// TestIntegration_GetStudent_InvalidToken asserts that a malformed JWT returns 401.
func TestIntegration_GetStudent_InvalidToken(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewGetStudentHandler(studentRepo))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/students/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestIntegration_GetStudent_Success asserts that an authenticated student gets
// the expected 200 payload with all required schema fields populated.
func TestIntegration_GetStudent_Success(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	const userID = "00000000-0000-0000-0000-aaa000000001"
	student := &repository.Student{
		ID:          userID,
		DisplayName: "Ana Lima",
		Username:    "ana01",
		Email:       "ana@example.com",
	}
	require.NoError(t, repository.NewStudentRepository(pool).Create(ctx, student))

	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewGetStudentHandler(studentRepo))

	token := testhelper.MakeTestJWT(t, userID)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/students/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp handler.StudentResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	// Schema contract: all required fields present
	assert.Equal(t, userID, resp.ID)
	assert.Equal(t, "Ana Lima", resp.DisplayName)
	assert.Equal(t, "ana01", resp.Username)
	assert.Equal(t, "ana@example.com", resp.Email)
	assert.GreaterOrEqual(t, resp.XPTotal, int64(0))
	assert.GreaterOrEqual(t, resp.StreakCount, 0)
	assert.GreaterOrEqual(t, resp.ReadinessScore, 0.0)
	assert.LessOrEqual(t, resp.ReadinessScore, 1.0)
}

// TestIntegration_GetStudent_NotFound asserts that a valid JWT for a non-existent
// student returns 404.
func TestIntegration_GetStudent_NotFound(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	studentRepo := repository.NewStudentRepository(pool)
	h := withAuth(handler.NewGetStudentHandler(studentRepo))

	// Valid UUID that does not exist in users.students.
	token := testhelper.MakeTestJWT(t, "00000000-0000-0000-0000-000000000001")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/students/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
