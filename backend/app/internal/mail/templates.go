package mail

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
)

// Build renders a Message from a validated SendRequest.
// Callers MUST call Validate before Build.
func Build(req SendRequest, fromEmail, fromName string) (Message, error) {
	switch req.Type {
	case TypeWelcome:
		return buildWelcome(req, fromEmail, fromName)
	case TypeEmailVerify:
		return buildOTP(req, fromEmail, fromName,
			fmt.Sprintf("Your PreUni verification code: %s", req.Params["otp"]),
			"verificar o seu endereço de e-mail")
	case TypeOTPLogin:
		return buildOTP(req, fromEmail, fromName,
			fmt.Sprintf("Your PreUni login code: %s", req.Params["otp"]),
			"entrar na sua conta PreUni")
	case TypeEmailChange:
		return buildEmailChange(req, fromEmail, fromName)
	case TypePasswordReset:
		return buildOTP(req, fromEmail, fromName,
			fmt.Sprintf("Your PreUni verification code: %s", req.Params["otp"]),
			"verificar o seu endereço de e-mail")
	default:
		return Message{}, ErrUnknownType
	}
}

// ── Templates ────────────────────────────────────────────────────────────────
//
// HTML bodies use html/template so user-controlled fields (display_name,
// verification_link, new_email) are context-escaped automatically:
//   - `{{.Name}}` in text → HTML-escaped (no tag injection)
//   - `{{.Link}}` in href → URL-escaped + scheme-filtered (no javascript:)
//
// Text bodies use text/template (plain string substitution; no HTML context).

var welcomeHTML = htmltemplate.Must(htmltemplate.New("welcome.html").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8" /></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h1 style="color: #6750A4;">Bem-vindo ao PreUni, {{.Name}}!</h1>
  <p>Estamos felizes em ter você por aqui. O PreUni vai te ajudar a se preparar para o ENEM e alcançar o Ensino Superior dos seus sonhos.</p>
  <p>Para começar, confirme o seu endereço de e-mail clicando no botão abaixo:</p>
  <p style="text-align: center; margin: 32px 0;">
    <a href="{{.Link}}" style="background: #6750A4; color: white; padding: 14px 28px; border-radius: 8px; text-decoration: none; font-weight: bold;">Verificar e-mail</a>
  </p>
  <p style="color: #666; font-size: 13px;">Se você não criou uma conta no PreUni, ignore este e-mail.</p>
</body>
</html>`))

var welcomeText = texttemplate.Must(texttemplate.New("welcome.txt").Parse(
	"Bem-vindo ao PreUni, {{.Name}}!\n\nConfirme o seu endereço de e-mail acessando o link abaixo:\n{{.Link}}\n\nSe você não criou uma conta no PreUni, ignore este e-mail.\n"))

var otpHTML = htmltemplate.Must(htmltemplate.New("otp.html").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8" /></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #6750A4;">Código de verificação</h2>
  <p>Use o código abaixo para {{.Purpose}}. O código expira em <strong>15 minutos</strong>.</p>
  <div style="text-align: center; margin: 32px 0;">
    <span style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #1C1B1F;">{{.OTP}}</span>
  </div>
  <p style="color: #666; font-size: 13px;">Se você não solicitou este código, ignore este e-mail.</p>
</body>
</html>`))

var otpText = texttemplate.Must(texttemplate.New("otp.txt").Parse(
	"Código PreUni: {{.OTP}}\n\nEste código expira em 15 minutos.\n\nSe você não solicitou este código, ignore este e-mail.\n"))

var emailChangeHTML = htmltemplate.Must(htmltemplate.New("email_change.html").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8" /></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #6750A4;">Verifique o seu novo e-mail</h2>
  <p>Uma solicitação de alteração de e-mail foi feita para a conta PreUni associada a <strong>{{.NewEmail}}</strong>.</p>
  <p>Use o código abaixo para confirmar o seu <strong>novo</strong> endereço de e-mail. O código expira em <strong>15 minutos</strong>.</p>
  <div style="text-align: center; margin: 32px 0;">
    <span style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #1C1B1F;">{{.OTP}}</span>
  </div>
  <p style="color: #666; font-size: 13px;">Se você não solicitou esta alteração, ignore este e-mail e o seu endereço de e-mail original permanecerá ativo.</p>
</body>
</html>`))

var emailChangeText = texttemplate.Must(texttemplate.New("email_change.txt").Parse(
	"Verificação de novo e-mail PreUni\n\nUma solicitação de alteração de e-mail foi feita para {{.NewEmail}}.\nCódigo de confirmação: {{.OTP}}\n\nEste código expira em 15 minutos.\n\nSe você não solicitou esta alteração, ignore este e-mail.\n"))

// ── Builders ─────────────────────────────────────────────────────────────────

func buildWelcome(req SendRequest, fromEmail, fromName string) (Message, error) {
	data := struct {
		Name string
		Link string
	}{
		Name: req.Params["display_name"],
		Link: req.Params["verification_link"],
	}
	html, err := render(welcomeHTML, data)
	if err != nil {
		return Message{}, err
	}
	text, err := render(welcomeText, data)
	if err != nil {
		return Message{}, err
	}
	// Subject is a header value, NOT an HTML context — escape control chars
	// only (sanitizeHeader handles CR/LF at the sender layer; here we just
	// keep the readable name visible).
	subject := fmt.Sprintf("Welcome to PreUni, %s! Verify your email", oneline(data.Name))
	return Message{
		From: fromEmail, FromName: fromName, To: req.To,
		Subject: subject, HTMLBody: html, TextBody: text,
	}, nil
}

func buildOTP(req SendRequest, fromEmail, fromName, subject, purpose string) (Message, error) {
	data := struct {
		OTP     string
		Purpose string
	}{
		OTP:     req.Params["otp"],
		Purpose: purpose,
	}
	html, err := render(otpHTML, data)
	if err != nil {
		return Message{}, err
	}
	text, err := render(otpText, data)
	if err != nil {
		return Message{}, err
	}
	return Message{
		From: fromEmail, FromName: fromName, To: req.To,
		Subject: subject, HTMLBody: html, TextBody: text,
	}, nil
}

func buildEmailChange(req SendRequest, fromEmail, fromName string) (Message, error) {
	data := struct {
		NewEmail string
		OTP      string
	}{
		NewEmail: req.To,
		OTP:      req.Params["otp"],
	}
	html, err := render(emailChangeHTML, data)
	if err != nil {
		return Message{}, err
	}
	text, err := render(emailChangeText, data)
	if err != nil {
		return Message{}, err
	}
	return Message{
		From: fromEmail, FromName: fromName, To: req.To,
		Subject: "Verify your new PreUni email address",
		HTMLBody: html, TextBody: text,
	}, nil
}

// render wraps html/template.Template + text/template.Template, both of
// which have Execute(io.Writer, any) error.
func render(t any, data any) (string, error) {
	var buf bytes.Buffer
	switch tt := t.(type) {
	case *htmltemplate.Template:
		if err := tt.Execute(&buf, data); err != nil {
			return "", err
		}
	case *texttemplate.Template:
		if err := tt.Execute(&buf, data); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("render: unsupported template type %T", t)
	}
	return buf.String(), nil
}

// oneline collapses CR/LF/tab into spaces. Subject header is sanitized at
// the SMTP sender layer (sanitizeHeader strips the bytes outright); doing
// this here keeps the visible Subject readable instead of mashed together
// when callers reach Build with control chars somehow.
func oneline(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(s)
}
