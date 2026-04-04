import Config

config :preuni, PreuniWeb.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: 4002],
  secret_key_base: "test_secret_key_base_change_in_prod_at_least_64_chars_long_abcdefgh",
  server: false

config :preuni, Preuni.Mailer, adapter: Swoosh.Adapters.Test

config :swoosh, :api_client, false

config :logger, level: :warning
