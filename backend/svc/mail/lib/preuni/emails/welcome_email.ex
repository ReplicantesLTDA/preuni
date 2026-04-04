defmodule Preuni.Emails.WelcomeEmail do
  import Swoosh.Email

  @from_email Application.compile_env(:preuni, :from_email, "noreply@preuni.com.br")
  @from_name Application.compile_env(:preuni, :from_name, "PreUni")

  @doc """
  Builds a welcome email for a newly registered student.

  Expected params:
    - `display_name` — student's display name
    - `verification_link` — full URL the student must click to verify their email
  """
  def build(to_email, %{"display_name" => display_name, "verification_link" => verification_link}) do
    new()
    |> from({@from_name, @from_email})
    |> to(to_email)
    |> subject("Welcome to PreUni, #{display_name}! Verify your email")
    |> html_body(html_body(display_name, verification_link))
    |> text_body(text_body(display_name, verification_link))
  end

  defp html_body(display_name, verification_link) do
    """
    <!DOCTYPE html>
    <html>
    <head><meta charset="utf-8" /></head>
    <body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
      <h1 style="color: #6750A4;">Bem-vindo ao PreUni, #{display_name}!</h1>
      <p>
        Estamos felizes em ter você por aqui. O PreUni vai te ajudar a se preparar para o ENEM
        e alcançar o Ensino Superior dos seus sonhos.
      </p>
      <p>Para começar, confirme o seu endereço de e-mail clicando no botão abaixo:</p>
      <p style="text-align: center; margin: 32px 0;">
        <a href="#{verification_link}"
           style="background: #6750A4; color: white; padding: 14px 28px;
                  border-radius: 8px; text-decoration: none; font-weight: bold;">
          Verificar e-mail
        </a>
      </p>
      <p style="color: #666; font-size: 13px;">
        Se você não criou uma conta no PreUni, ignore este e-mail.
      </p>
    </body>
    </html>
    """
  end

  defp text_body(display_name, verification_link) do
    """
    Bem-vindo ao PreUni, #{display_name}!

    Confirme o seu endereço de e-mail acessando o link abaixo:
    #{verification_link}

    Se você não criou uma conta no PreUni, ignore este e-mail.
    """
  end
end
