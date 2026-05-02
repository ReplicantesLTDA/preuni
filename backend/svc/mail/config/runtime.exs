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
  smtp_port = String.to_integer(System.get_env("SMTP_PORT", "587"))
  smtp_host = System.fetch_env!("SMTP_HOST")

  # tls_options is only used by gen_smtp on the STARTTLS upgrade path (port 587).
  # For implicit SSL (port 465), SSL options must go through sockopts instead.
  smtp_sockopts =
    if smtp_port == 465 do
      [verify: :verify_peer, cacertfile: ~c"/etc/ssl/certs/ca-certificates.crt",
       server_name_indication: String.to_charlist(smtp_host), depth: 10,
       customize_hostname_check: [match_fun: :public_key.pkix_verify_hostname_match_fun(:https)]]
    else
      []
    end

  config :preuni, Preuni.Mailer,
    adapter: Swoosh.Adapters.SMTP,
    relay: smtp_host,
    port: smtp_port,
    username: System.fetch_env!("SMTP_USER"),
    password: System.fetch_env!("SMTP_PASS"),
    ssl: smtp_port == 465,
    tls: if(smtp_port == 465, do: :never, else: :always),
    tls_options: [verify: :verify_none],
    sockopts: smtp_sockopts,
    auth: :always
else
  config :preuni, Preuni.Mailer,
    adapter: Swoosh.Adapters.Local
end

config :preuni,
  from_email: System.get_env("FROM_EMAIL", "noreply@preuni.com.br"),
  from_name: System.get_env("MAIL_FROM_NAME", "PreUni")
