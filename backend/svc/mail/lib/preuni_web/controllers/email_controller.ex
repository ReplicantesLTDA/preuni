defmodule PreuniWeb.EmailController do
  use Phoenix.Controller, formats: [:json]

  alias Preuni.Mailer
  alias Preuni.Emails.{WelcomeEmail, VerificationEmail, OtpLoginEmail, EmailChangeEmail}

  @supported_types ~w(WELCOME EMAIL_VERIFY OTP_LOGIN EMAIL_CHANGE PASSWORD_RESET)

  @doc """
  POST /internal/email/send

  Body:
    {
      "type": "WELCOME" | "EMAIL_VERIFY" | "OTP_LOGIN" | "EMAIL_CHANGE" | "PASSWORD_RESET",
      "to": "recipient@example.com",
      "params": { ... }
    }
  """
  def send(conn, %{"type" => type, "to" => to, "params" => params})
      when type in @supported_types do
    case build_email(type, to, params) do
      {:ok, email} ->
        case Mailer.deliver(email) do
          {:ok, _} ->
            conn
            |> put_status(:accepted)
            |> json(%{status: "queued"})

          {:error, reason} ->
            conn
            |> put_status(:internal_server_error)
            |> json(%{error: "delivery failed", reason: inspect(reason)})
        end

      {:error, message} ->
        conn
        |> put_status(:unprocessable_entity)
        |> json(%{error: message})
    end
  end

  def send(conn, %{"type" => type}) when type not in @supported_types do
    conn
    |> put_status(:unprocessable_entity)
    |> json(%{error: "unknown email type: #{type}"})
  end

  def send(conn, _params) do
    conn
    |> put_status(:unprocessable_entity)
    |> json(%{error: "type, to, and params are required"})
  end

  defp build_email("WELCOME", to, params) do
    if Map.has_key?(params, "display_name") and Map.has_key?(params, "verification_link") do
      {:ok, WelcomeEmail.build(to, params)}
    else
      {:error, "WELCOME requires display_name and verification_link params"}
    end
  end

  defp build_email("EMAIL_VERIFY", to, params) do
    if Map.has_key?(params, "otp") do
      {:ok, VerificationEmail.build(to, params)}
    else
      {:error, "EMAIL_VERIFY requires otp param"}
    end
  end

  defp build_email("OTP_LOGIN", to, params) do
    if Map.has_key?(params, "otp") do
      {:ok, OtpLoginEmail.build(to, params)}
    else
      {:error, "OTP_LOGIN requires otp param"}
    end
  end

  defp build_email("EMAIL_CHANGE", to, params) do
    if Map.has_key?(params, "otp") do
      {:ok, EmailChangeEmail.build(to, params)}
    else
      {:error, "EMAIL_CHANGE requires otp param"}
    end
  end

  defp build_email("PASSWORD_RESET", to, params) do
    if Map.has_key?(params, "otp") do
      {:ok, VerificationEmail.build(to, params)}
    else
      {:error, "PASSWORD_RESET requires otp param"}
    end
  end
end
