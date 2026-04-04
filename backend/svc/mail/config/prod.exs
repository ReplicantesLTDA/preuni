import Config

config :preuni, PreuniWeb.Endpoint,
  http: [ip: {0, 0, 0, 0}, port: String.to_integer(System.get_env("PORT", "4000"))],
  secret_key_base: System.fetch_env!("SECRET_KEY_BASE"),
  server: true

config :preuni, Preuni.Mailer,
  adapter: Swoosh.Adapters.SMTP,
  relay: System.fetch_env!("SMTP_HOST"),
  port: String.to_integer(System.get_env("SMTP_PORT", "587")),
  username: System.fetch_env!("SMTP_USER"),
  password: System.fetch_env!("SMTP_PASS"),
  tls: :always,
  auth: :always

config :logger, level: :info
