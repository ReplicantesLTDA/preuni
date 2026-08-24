package mail

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/smtp"
	"strings"
)

// Sender delivers a Message via SMTP.
type Sender interface {
	Send(Message) error
}

// SMTPConfig holds SMTP server credentials.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool // true → implicit TLS (port 465); false → STARTTLS (587)

	// RootCAs overrides the trust store used for the implicit-TLS dial.
	// Nil (the default, and the only value ever set in production) means
	// "use the system trust store" -- identical to prior behavior. Tests
	// use this to trust a self-signed cert from a real fake SMTP server
	// without disabling verification.
	RootCAs *x509.CertPool
}

// SMTPSender is a production Sender backed by net/smtp.
type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender { return &SMTPSender{cfg: cfg} }

// Send delivers msg via SMTP. On implicit-TLS (port 465) it dials with TLS,
// otherwise it relies on STARTTLS during the SMTP conversation.
func (s *SMTPSender) Send(msg Message) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	body := encodeMIME(msg)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	from := msg.From
	if from == "" {
		from = s.cfg.Username
	}

	if s.cfg.UseTLS {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12, RootCAs: s.cfg.RootCAs})
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		defer conn.Close()
		c, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		defer c.Close()
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := c.Mail(from); err != nil {
			return fmt.Errorf("smtp mail: %w", err)
		}
		if err := c.Rcpt(msg.To); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
		wc, err := c.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		if _, err := wc.Write([]byte(body)); err != nil {
			return fmt.Errorf("smtp write: %w", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("smtp close: %w", err)
		}
		return c.Quit()
	}

	// STARTTLS / plain path
	return smtp.SendMail(addr, auth, from, []string{msg.To}, []byte(body))
}

// sanitizeHeader strips CR and LF (and embedded NUL) from a header value to
// prevent SMTP header injection via attacker-controlled input that flows into
// To / From / Subject. RFC 5322 forbids these bytes in header field values
// anyway; rejecting them at the boundary is the safe default.
func sanitizeHeader(v string) string {
	return strings.NewReplacer("\r", "", "\n", "", "\x00", "").Replace(v)
}

func encodeMIME(msg Message) string {
	from := msg.From
	if msg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", msg.FromName, msg.From)
	}
	boundary := "preuni-mime-boundary"
	headers := []string{
		"From: " + sanitizeHeader(from),
		"To: " + sanitizeHeader(msg.To),
		"Subject: " + sanitizeHeader(msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: multipart/alternative; boundary=" + boundary,
	}
	body := []string{
		strings.Join(headers, "\r\n"),
		"",
		"--" + boundary,
		"Content-Type: text/plain; charset=UTF-8",
		"",
		msg.TextBody,
		"--" + boundary,
		"Content-Type: text/html; charset=UTF-8",
		"",
		msg.HTMLBody,
		"--" + boundary + "--",
	}
	return strings.Join(body, "\r\n")
}

// NoopSender discards messages. Used when SMTP is not configured (dev/local).
type NoopSender struct{}

func (NoopSender) Send(_ Message) error { return nil }
