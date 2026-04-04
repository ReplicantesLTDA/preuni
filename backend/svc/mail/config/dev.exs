import Config

config :preuni, PreuniWeb.Endpoint,
  http: [ip: {0, 0, 0, 0}, port: 4000],
  check_origin: false,
  debug_errors: true,
  secret_key_base: "dev_secret_key_base_change_in_prod_at_least_64_chars_long_abcdefgh"

config :preuni, Preuni.Mailer,
  adapter: Swoosh.Adapters.Local

config :logger, :console, format: "[$level] $message\n"
