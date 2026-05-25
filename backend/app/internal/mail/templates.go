package mail

import "fmt"

// Build renders a Message from a validated SendRequest.
// Callers MUST call Validate before Build.
func Build(req SendRequest, fromEmail, fromName string) (Message, error) {
	switch req.Type {
	case TypeWelcome:
		return buildWelcome(req, fromEmail, fromName), nil
	case TypeEmailVerify:
		return buildOTP(req, fromEmail, fromName,
			fmt.Sprintf("Your PreUni verification code: %s", req.Params["otp"]),
			"verificar o seu endereço de e-mail",
		), nil
	case TypeOTPLogin:
		return buildOTP(req, fromEmail, fromName,
			fmt.Sprintf("Your PreUni login code: %s", req.Params["otp"]),
			"entrar na sua conta PreUni",
		), nil
	case TypeEmailChange:
		return buildEmailChange(req, fromEmail, fromName), nil
	case TypePasswordReset:
		return buildOTP(req, fromEmail, fromName,
			fmt.Sprintf("Your PreUni verification code: %s", req.Params["otp"]),
			"verificar o seu endereço de e-mail",
		), nil
	default:
		return Message{}, ErrUnknownType
	}
}

func buildWelcome(req SendRequest, fromEmail, fromName string) Message {
	name := req.Params["display_name"]
	link := req.Params["verification_link"]
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8" /></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h1 style="color: #6750A4;">Bem-vindo ao PreUni, %s!</h1>
  <p>Estamos felizes em ter você por aqui. O PreUni vai te ajudar a se preparar para o ENEM e alcançar o Ensino Superior dos seus sonhos.</p>
  <p>Para começar, confirme o seu endereço de e-mail clicando no botão abaixo:</p>
  <p style="text-align: center; margin: 32px 0;">
    <a href="%s" style="background: #6750A4; color: white; padding: 14px 28px; border-radius: 8px; text-decoration: none; font-weight: bold;">Verificar e-mail</a>
  </p>
  <p style="color: #666; font-size: 13px;">Se você não criou uma conta no PreUni, ignore este e-mail.</p>
</body>
</html>`, name, link)
	text := fmt.Sprintf("Bem-vindo ao PreUni, %s!\n\nConfirme o seu endereço de e-mail acessando o link abaixo:\n%s\n\nSe você não criou uma conta no PreUni, ignore este e-mail.\n", name, link)
	return Message{
		From: fromEmail, FromName: fromName, To: req.To,
		Subject:  fmt.Sprintf("Welcome to PreUni, %s! Verify your email", name),
		HTMLBody: html, TextBody: text,
	}
}

func buildOTP(req SendRequest, fromEmail, fromName, subject, purpose string) Message {
	otp := req.Params["otp"]
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8" /></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #6750A4;">Código de verificação</h2>
  <p>Use o código abaixo para %s. O código expira em <strong>15 minutos</strong>.</p>
  <div style="text-align: center; margin: 32px 0;">
    <span style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #1C1B1F;">%s</span>
  </div>
  <p style="color: #666; font-size: 13px;">Se você não solicitou este código, ignore este e-mail.</p>
</body>
</html>`, purpose, otp)
	text := fmt.Sprintf("Código PreUni: %s\n\nEste código expira em 15 minutos.\n\nSe você não solicitou este código, ignore este e-mail.\n", otp)
	return Message{
		From: fromEmail, FromName: fromName, To: req.To,
		Subject:  subject,
		HTMLBody: html, TextBody: text,
	}
}

func buildEmailChange(req SendRequest, fromEmail, fromName string) Message {
	otp := req.Params["otp"]
	newEmail := req.To
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8" /></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #6750A4;">Verifique o seu novo e-mail</h2>
  <p>Uma solicitação de alteração de e-mail foi feita para a conta PreUni associada a <strong>%s</strong>.</p>
  <p>Use o código abaixo para confirmar o seu <strong>novo</strong> endereço de e-mail. O código expira em <strong>15 minutos</strong>.</p>
  <div style="text-align: center; margin: 32px 0;">
    <span style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #1C1B1F;">%s</span>
  </div>
  <p style="color: #666; font-size: 13px;">Se você não solicitou esta alteração, ignore este e-mail e o seu endereço de e-mail original permanecerá ativo.</p>
</body>
</html>`, newEmail, otp)
	text := fmt.Sprintf("Verificação de novo e-mail PreUni\n\nUma solicitação de alteração de e-mail foi feita para %s.\nCódigo de confirmação: %s\n\nEste código expira em 15 minutos.\n\nSe você não solicitou esta alteração, ignore este e-mail.\n", newEmail, otp)
	return Message{
		From: fromEmail, FromName: fromName, To: req.To,
		Subject:  "Verify your new PreUni email address",
		HTMLBody: html, TextBody: text,
	}
}
