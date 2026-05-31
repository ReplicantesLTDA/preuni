// Package mail provides the in-process mail component used by the monolith
// to serve POST /internal/email/send.
package mail

// EmailType enumerates the supported transactional email types.
type EmailType string

const (
	TypeWelcome       EmailType = "WELCOME"
	TypeEmailVerify   EmailType = "EMAIL_VERIFY"
	TypeOTPLogin      EmailType = "OTP_LOGIN"
	TypeEmailChange   EmailType = "EMAIL_CHANGE"
	TypePasswordReset EmailType = "PASSWORD_RESET"
)

// SendRequest is the JSON body of POST /internal/email/send.
type SendRequest struct {
	Type   EmailType         `json:"type"`
	To     string            `json:"to"`
	Params map[string]string `json:"params"`
}

// Message is a fully rendered email ready for SMTP delivery.
type Message struct {
	From     string
	FromName string
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}
