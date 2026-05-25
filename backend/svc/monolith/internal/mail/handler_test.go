package mail

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/preuni/pkg/logger"
)

type recordingSender struct {
	mu   sync.Mutex
	msgs []Message
}

func (r *recordingSender) Send(m Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs = append(r.msgs, m)
	return nil
}

func (r *recordingSender) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.msgs)
}

func TestHandler_Accepted(t *testing.T) {
	rec := &recordingSender{}
	h := NewHandler("noreply@preuni.com.br", "PreUni", rec, logger.New("error"))
	body, _ := json.Marshal(SendRequest{
		Type:   TypeEmailVerify,
		To:     "a@b.c",
		Params: map[string]string{"otp": "123456"},
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/email/send", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status: got %d want 202", w.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "queued" {
		t.Fatalf("body: %+v", resp)
	}
	// Allow the goroutine to run.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && rec.count() == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	if rec.count() != 1 {
		t.Fatalf("sender invocations: got %d want 1", rec.count())
	}
}

func TestHandler_ValidationFails(t *testing.T) {
	rec := &recordingSender{}
	h := NewHandler("noreply@preuni.com.br", "PreUni", rec, logger.New("error"))
	body, _ := json.Marshal(SendRequest{Type: TypeWelcome, To: "a@b.c", Params: map[string]string{"display_name": "A"}})
	req := httptest.NewRequest(http.MethodPost, "/internal/email/send", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: got %d want 422", w.Code)
	}
}
