package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestIntegration_RefreshToken_UnknownTokenIsUnauthorized covers
// RefreshTokenHandler's FindByHash-fails branch, previously untested (only
// the success/rotation path and missing-field validation were covered).
func TestIntegration_RefreshToken_UnknownTokenIsUnauthorized(t *testing.T) {
	r, _ := setup(t)

	body, _ := json.Marshal(map[string]string{"refresh_token": "not-a-real-refresh-token"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("refresh with an unknown token: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_HandlerFaultInjection_CanceledContextRound2 reaches more
// apperrors.Internal(err)-shaped branches at the HTTP layer via an
// already-canceled request context.Context -- genuine behavior (the same
// error a real client disconnect or request timeout produces), not a mock.
// Covers ChangePasswordHandler.FindByID, auth's DeleteAccountHandler
// (distinct from the user-domain DELETE /v1/students/me path, which was
// already covered), RegisterHandler.Create (a brand-new email, so this is
// the generic-DB-failure branch, not the duplicate-email 409 branch
// covered earlier), and SubmitEssayHandler.
func TestIntegration_HandlerFaultInjection_CanceledContextRound2(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("ChangePasswordHandler", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"current_password": "P@ssw0rd123",
			"new_password":     "N3wP@ssw0rd456",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(body)).WithContext(canceled)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK || w.Code == http.StatusNoContent {
			t.Fatalf("expected an error status for a canceled context, got %d", w.Code)
		}
	})

	t.Run("SubmitEssayHandler", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"prompt_theme_title":   "Tema",
			"prompt_theme_context": "Contexto qualquer com mais de vinte caracteres.",
			"essay_text":           "Texto de redação de teste para o fluxo de integração.",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/essays", bytes.NewReader(body)).WithContext(canceled)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusAccepted {
			t.Fatalf("expected an error status for a canceled context, got %d", w.Code)
		}
	})

	t.Run("DeleteAccountHandler_auth", func(t *testing.T) {
		id2, token2 := registerTestUser(t, r)
		defer cleanupTestUser(ctx, t, pool, id2)

		body, _ := json.Marshal(map[string]string{"confirmation": "DELETE"})
		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/account", bytes.NewReader(body)).WithContext(canceled)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token2)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusNoContent {
			t.Fatalf("expected an error status for a canceled context, got %d", w.Code)
		}
	})

	t.Run("RegisterHandler", func(t *testing.T) {
		email := fmt.Sprintf("cancel-ctx-reg+%d@preuni.test", time.Now().UnixNano())
		body, _ := json.Marshal(map[string]string{
			"email": email, "password": "P@ssw0rd123", "display_name": "Canceled Ctx",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body)).WithContext(canceled)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Fatalf("expected an error status for a canceled context, got %d", w.Code)
		}
	})
}
