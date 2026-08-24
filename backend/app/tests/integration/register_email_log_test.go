package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/router"
	"github.com/preuni/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// closedSMTPPort binds and immediately closes a real TCP listener, giving
// back an address that genuinely refuses connections -- the same
// technique internal/mail/sender_network_test.go uses.
func closedSMTPPort(t *testing.T) (string, int) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	if err := l.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	return addr.IP.String(), addr.Port
}

// TestIntegration_Register_WelcomeEmailSendFailureIsLogged covers
// RegisterHandler.fireAndForgetEmail's log.Error branch (previously
// unreachable in tests -- the fire-and-forget goroutine's only observable
// effect on failure is a log line). Points the router's SMTP config at a
// real closed port (genuine dial-refused, not a mock of mail.Sender) and
// observes the resulting log entry via a zaptest/observer core injected
// through logger.Wrap, instead of touching a shared os.Stderr.
func TestIntegration_Register_WelcomeEmailSendFailureIsLogged(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("postgres unreachable: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("postgres ping failed: %v", err)
	}
	defer pool.Close()

	host, port := closedSMTPPort(t)

	core, observed := observer.New(zapcore.ErrorLevel)
	testLog := logger.Wrap(zap.New(core))

	cfg := config.Config{
		JWTSigningKey:        "test-signing-key-32-bytes-minimum!",
		JWTAccessExpirySec:   3600,
		JWTRefreshExpiryDays: 30,
		MailFromAddr:         "noreply@test",
		MailFromName:         "Test",
		S3Bucket:             "test-bucket",
		S3Region:             "us-east-1",
		SMTPHost:             host,
		SMTPPort:             port,
	}
	r, _, _ := router.New(cfg, pool, testLog)

	email := fmt.Sprintf("integ-emaillog+%d@preuni.test", time.Now().UnixNano())
	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "P@ssw0rd123", "display_name": "Integ",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d body=%s", w.Code, w.Body.String())
	}

	// The WELCOME email dispatch is fire-and-forget (a bare `go` call with
	// no synchronization point available to the handler's response), so
	// poll briefly for the async log entry instead of racing it.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if observed.Len() > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	entries := observed.All()
	if len(entries) == 0 {
		t.Fatal("expected at least one observed error log entry for the failed email send")
	}
	found := false
	for _, e := range entries {
		if e.Message == "email send failed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected an \"email send failed\" log entry, got: %+v", entries)
	}

	// Cleanup: anonymize this account's data.
	ctx := context.Background()
	_, _ = pool.Exec(ctx, `DELETE FROM auth.refresh_tokens WHERE credential_id IN (SELECT id FROM auth.credentials WHERE email=$1)`, email)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.otp_codes WHERE credential_id IN (SELECT id FROM auth.credentials WHERE email=$1)`, email)
	_, _ = pool.Exec(ctx, `DELETE FROM users.students WHERE email=$1`, email)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.credentials WHERE email=$1`, email)
}
