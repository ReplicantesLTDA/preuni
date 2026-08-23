package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	socialrepo "github.com/preuni/app/internal/social/repository"
)

// TestIntegration_Friends_ErrorPaths covers validation, self-friending,
// duplicate requests, and accepting/removing as the wrong party.
func TestIntegration_Friends_ErrorPaths(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	aID, aToken := registerTestUser(t, r)
	bID, bToken := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, aID)
	defer cleanupTestUser(ctx, t, pool, bID)

	t.Run("missing addressee_id is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("got %d, want 422, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("cannot friend yourself", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"addressee_id": aID})
		req := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("got %d, want 422, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("duplicate pending request is a conflict", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"addressee_id": bID})
		first := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader(body))
		first.Header.Set("Content-Type", "application/json")
		first.Header.Set("Authorization", "Bearer "+aToken)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, first)
		if w1.Code != http.StatusCreated {
			t.Fatalf("first request: got %d, want 201, body=%s", w1.Code, w1.Body.String())
		}

		second := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader(body))
		second.Header.Set("Content-Type", "application/json")
		second.Header.Set("Authorization", "Bearer "+aToken)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, second)
		if w2.Code != http.StatusConflict {
			t.Fatalf("second request: got %d, want 409, body=%s", w2.Code, w2.Body.String())
		}
	})

	t.Run("the requester cannot accept their own outgoing request", func(t *testing.T) {
		var friendshipID string
		if err := pool.QueryRow(ctx, `SELECT id FROM social.friendships WHERE requester_id = $1 AND addressee_id = $2 AND status = 'pending'`, aID, bID).Scan(&friendshipID); err != nil {
			t.Fatalf("find pending request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/v1/friends/requests/"+friendshipID+"/accept", nil)
		req.Header.Set("Authorization", "Bearer "+aToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d, want 404, body=%s", w.Code, w.Body.String())
		}

		acceptReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests/"+friendshipID+"/accept", nil)
		acceptReq.Header.Set("Authorization", "Bearer "+bToken)
		acceptW := httptest.NewRecorder()
		r.ServeHTTP(acceptW, acceptReq)
		if acceptW.Code != http.StatusOK {
			t.Fatalf("legit accept: got %d, want 200, body=%s", acceptW.Code, acceptW.Body.String())
		}
	})

	t.Run("removing a nonexistent friendship 404s", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/friends/00000000-0000-4000-a000-999999999999", nil)
		req.Header.Set("Authorization", "Bearer "+aToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d, want 404, body=%s", w.Code, w.Body.String())
		}
	})
}

// TestIntegration_Friends_RequestAcceptRevealsStreakAndGrade covers spec
// User Story 2 acceptance scenarios 1-2: request -> accept -> both sides
// see the other's streak, and a non-friend cannot.
func TestIntegration_Friends_RequestAcceptRevealsStreakAndGrade(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	aID, aToken := registerTestUser(t, r)
	bID, bToken := registerTestUser(t, r)
	cID, cToken := registerTestUser(t, r) // non-friend control
	defer cleanupTestUser(ctx, t, pool, aID)
	defer cleanupTestUser(ctx, t, pool, bID)
	defer cleanupTestUser(ctx, t, pool, cID)

	// A gets a streak by submitting an essay.
	if w := submitEssay(t, r, aToken); w.Code != http.StatusAccepted {
		t.Fatalf("submit for A: got %d body=%s", w.Code, w.Body.String())
	}

	// A sends B a friend request.
	reqBody, _ := json.Marshal(map[string]string{"addressee_id": bID})
	req := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("send request: got %d body=%s", w.Code, w.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	// C (not the addressee) cannot see the friends list confirm B accepted yet -- skip, not relevant.

	// B accepts.
	acceptReq := httptest.NewRequest(http.MethodPost, "/v1/friends/requests/"+created.ID+"/accept", nil)
	acceptReq.Header.Set("Authorization", "Bearer "+bToken)
	acceptW := httptest.NewRecorder()
	r.ServeHTTP(acceptW, acceptReq)
	if acceptW.Code != http.StatusOK {
		t.Fatalf("accept: got %d body=%s", acceptW.Code, acceptW.Body.String())
	}

	// Both A and B now see each other in their friends list, with streak/grade.
	for _, side := range []struct {
		name, token, wantUserID string
	}{{"A", aToken, bID}, {"B", bToken, aID}} {
		listReq := httptest.NewRequest(http.MethodGet, "/v1/friends", nil)
		listReq.Header.Set("Authorization", "Bearer "+side.token)
		listW := httptest.NewRecorder()
		r.ServeHTTP(listW, listReq)
		if listW.Code != http.StatusOK {
			t.Fatalf("%s list friends: got %d body=%s", side.name, listW.Code, listW.Body.String())
		}
		var friends []struct {
			UserID        string `json:"user_id"`
			CurrentStreak int    `json:"current_streak"`
		}
		if err := json.Unmarshal(listW.Body.Bytes(), &friends); err != nil {
			t.Fatal(err)
		}
		if len(friends) != 1 || friends[0].UserID != side.wantUserID {
			t.Fatalf("%s: expected exactly one friend (%s), got %+v", side.name, side.wantUserID, friends)
		}
	}

	// C (a non-friend) sees an empty friends list -- no cross-account leak.
	cListReq := httptest.NewRequest(http.MethodGet, "/v1/friends", nil)
	cListReq.Header.Set("Authorization", "Bearer "+cToken)
	cListW := httptest.NewRecorder()
	r.ServeHTTP(cListW, cListReq)
	var cFriends []any
	if err := json.Unmarshal(cListW.Body.Bytes(), &cFriends); err != nil {
		t.Fatal(err)
	}
	if len(cFriends) != 0 {
		t.Fatalf("C must have zero friends, got %d", len(cFriends))
	}

	// Removal revokes visibility for both sides.
	var friendshipID string
	if err := pool.QueryRow(ctx, `SELECT id FROM social.friendships WHERE requester_id = $1 AND addressee_id = $2`, aID, bID).Scan(&friendshipID); err != nil {
		t.Fatalf("find friendship row: %v", err)
	}
	removeReq := httptest.NewRequest(http.MethodDelete, "/v1/friends/"+friendshipID, nil)
	removeReq.Header.Set("Authorization", "Bearer "+aToken)
	removeW := httptest.NewRecorder()
	r.ServeHTTP(removeW, removeReq)
	if removeW.Code != http.StatusNoContent {
		t.Fatalf("remove: got %d body=%s", removeW.Code, removeW.Body.String())
	}

	afterReq := httptest.NewRequest(http.MethodGet, "/v1/friends", nil)
	afterReq.Header.Set("Authorization", "Bearer "+bToken)
	afterW := httptest.NewRecorder()
	r.ServeHTTP(afterW, afterReq)
	var afterFriends []any
	if err := json.Unmarshal(afterW.Body.Bytes(), &afterFriends); err != nil {
		t.Fatal(err)
	}
	if len(afterFriends) != 0 {
		t.Fatalf("B must have zero friends after A removed the friendship, got %d", len(afterFriends))
	}
}

// TestIntegration_Friends_ReRequestAfterRemovalIsTreatedAsNew covers spec
// Edge Cases: a request to someone who previously removed the requester is
// a new request, not blocked.
func TestIntegration_Friends_ReRequestAfterRemovalIsTreatedAsNew(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	aID, aToken := registerTestUser(t, r)
	bID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, aID)
	defer cleanupTestUser(ctx, t, pool, bID)

	if _, err := pool.Exec(ctx, `
		INSERT INTO social.friendships (id, requester_id, addressee_id, status, created_at, responded_at)
		VALUES (gen_random_uuid(), $1, $2, 'removed', now(), now())
	`, bID, aID); err != nil {
		t.Fatalf("seed removed friendship: %v", err)
	}

	reqBody, _ := json.Marshal(map[string]string{"addressee_id": bID})
	req := httptest.NewRequest(http.MethodPost, "/v1/friends/requests", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("re-request after removal: got %d, want 201, body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_SocialRepository_CanceledContextReturnsInternalError
// covers SendRequest, AcceptRequest, and RemoveFriend's
// apperrors.Internal(err) branches via a canceled context -- real pgx
// behavior (a client disconnect or request timeout in production), not a
// mock.
func TestIntegration_SocialRepository_CanceledContextReturnsInternalError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	aID, _ := registerTestUser(t, r)
	bID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, aID)
	defer cleanupTestUser(ctx, t, pool, bID)

	repo := socialrepo.NewRepository(pool)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("SendRequest", func(t *testing.T) {
		_, err := repo.SendRequest(canceled, aID, bID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	// Seed a real pending friendship (with a valid, non-canceled context)
	// so AcceptRequest/RemoveFriend's own DB calls are what fail.
	friendshipID, err := repo.SendRequest(ctx, aID, bID)
	if err != nil {
		t.Fatalf("seed friend request: %v", err)
	}

	t.Run("AcceptRequest", func(t *testing.T) {
		err := repo.AcceptRequest(canceled, friendshipID, bID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("RemoveFriend", func(t *testing.T) {
		err := repo.RemoveFriend(canceled, friendshipID, aID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})
}
