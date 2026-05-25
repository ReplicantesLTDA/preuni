package mail

import (
	"encoding/json"
	"net/http"

	"github.com/preuni/pkg/logger"
)

// Handler implements POST /internal/email/send.
// On valid request → 202 + {"status":"queued"} (delivery happens async).
// On invalid params → 422 + {"error":"…"} (matches Elixir mail-svc shape).
type Handler struct {
	FromEmail string
	FromName  string
	Sender    Sender
	Log       *logger.Logger
}

// NewHandler constructs a Handler. Sender may be NoopSender for local dev.
func NewHandler(fromEmail, fromName string, sender Sender, log *logger.Logger) *Handler {
	return &Handler{FromEmail: fromEmail, FromName: fromName, Sender: sender, Log: log}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "type, to, and params are required")
		return
	}
	if err := Validate(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	msg, err := Build(req, h.FromEmail, h.FromName)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Fire-and-forget delivery; failures surface only in logs.
	go func() {
		if err := h.Sender.Send(msg); err != nil {
			h.Log.Error("email delivery failed",
				logger.String("type", string(req.Type)),
				logger.String("to", req.To),
				logger.Err(err),
			)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
