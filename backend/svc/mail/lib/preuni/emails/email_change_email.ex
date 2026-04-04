defmodule Preuni.Emails.EmailChangeEmail do
  import Swoosh.Email

  @from_email Application.compile_env(:preuni, :from_email, "noreply@preuni.com.br")
  @from_name Application.compile_env(:preuni, :from_name, "PreUni")

  @doc """
  Builds an email-change confirmation OTP email sent to the NEW address.

  Expected params:
    - `otp` — 6-digit numeric code
  """
  def build(to_email, %{"otp" => otp}) do
    new()
    |> from({@from_name, @from_email})
    |> to(to_email)
    |> subject("Verify your new PreUni email address")
    |> html_body(render_html(otp, to_email))
    |> text_body(render_text(otp, to_email))
  end

  defp render_html(otp, new_email) do
    """
    <!DOCTYPE html>
    <html>
    <head><meta charset="utf-8" /></head>
    <body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
      <h2 style="color: #6750A4;">Verifique o seu novo e-mail</h2>
      <p>
        Uma solicitação de alteração de e-mail foi feita para a conta PreUni associada a
        <strong>#{new_email}</strong>.
      </p>
      <p>Use o código abaixo para confirmar o seu <strong>novo</strong> endereço de e-mail. O código expira em <strong>15 minutos</strong>.</p>
      <div style="text-align: center; margin: 32px 0;">
        <span style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #1C1B1F;">
          #{otp}
        </span>
      </div>
      <p style="color: #666; font-size: 13px;">
        Se você não solicitou esta alteração, ignore este e-mail e o seu endereço de e-mail original
        permanecerá ativo.
      </p>
    </body>
    </html>
    """
  end

  defp render_text(otp, new_email) do
    """
    Verificação de novo e-mail PreUni

    Uma solicitação de alteração de e-mail foi feita para #{new_email}.
    Código de confirmação: #{otp}

    Este código expira em 15 minutos.

    Se você não solicitou esta alteração, ignore este e-mail.
    """
  end
end
