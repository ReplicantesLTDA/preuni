defmodule Preuni.Emails.VerificationEmail do
  import Swoosh.Email

  @from_email Application.compile_env(:preuni, :from_email, "noreply@preuni.com.br")
  @from_name Application.compile_env(:preuni, :from_name, "PreUni")

  @doc """
  Builds an email verification OTP email.

  Expected params:
    - `otp` — 6-digit numeric code
  """
  def build(to_email, %{"otp" => otp}) do
    new()
    |> from({@from_name, @from_email})
    |> to(to_email)
    |> subject("Your PreUni verification code: #{otp}")
    |> html_body(render_html(otp))
    |> text_body(render_text(otp))
  end

  defp render_html(otp) do
    """
    <!DOCTYPE html>
    <html>
    <head><meta charset="utf-8" /></head>
    <body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
      <h2 style="color: #6750A4;">Código de verificação</h2>
      <p>Use o código abaixo para verificar o seu endereço de e-mail. O código expira em <strong>15 minutos</strong>.</p>
      <div style="text-align: center; margin: 32px 0;">
        <span style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #1C1B1F;">
          #{otp}
        </span>
      </div>
      <p style="color: #666; font-size: 13px;">
        Se você não solicitou este código, ignore este e-mail.
      </p>
    </body>
    </html>
    """
  end

  defp render_text(otp) do
    """
    Código de verificação PreUni: #{otp}

    Este código expira em 15 minutos.

    Se você não solicitou este código, ignore este e-mail.
    """
  end
end
