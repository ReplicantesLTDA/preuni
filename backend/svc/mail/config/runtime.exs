import Config

# Runtime configuration — read from environment variables at startup.
# This file is evaluated when the application boots in any Mix environment.

port = String.to_integer(System.get_env("PORT", "4000"))
host = System.get_env("PHX_HOST", "localhost")

config :preuni, PreuniWeb.Endpoint,
  http: [ip: {0, 0, 0, 0}, port: port],
  url: [host: host, port: port, scheme: "http"],
  secret_key_base:
    System.get_env("SECRET_KEY_BASE", "dev-secret-key-base-minimum-64-characters-long-change-in-prod"),
  server: true

# INTERNAL_TOKEN — shared secret for service-to-service auth
config :preuni, :internal_token,
  System.get_env("INTERNAL_TOKEN", "dev-internal-token")

# Mailer adapter — "local" uses Swoosh.Adapters.Local (dev mailbox),
# "smtp" uses Swoosh.Adapters.SMTP (production).
if System.get_env("SWOOSH_ADAPTER", "local") == "smtp" do
  config :preuni, Preuni.Mailer,
    adapter: Swoosh.Adapters.SMTP,
    relay: System.fetch_env!("SMTP_HOST"),
    port: String.to_integer(System.get_env("SMTP_PORT", "587")),
    username: System.fetch_env!("SMTP_USER"),
    password: System.fetch_env!("SMTP_PASS"),
    tls: :always,
    auth: :always
else
  config :preuni, Preuni.Mailer,
    adapter: Swoosh.Adapters.Local
end

config :preuni,
  from_email: System.get_env("FROM_EMAIL", "noreply@preuni.com.br"),
  from_name: System.get_env("MAIL_FROM_NAME", "PreUni")
