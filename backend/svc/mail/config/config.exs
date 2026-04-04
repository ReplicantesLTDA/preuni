import Config

config :preuni, PreuniWeb.Endpoint,
  url: [host: "localhost"],
  render_errors: [view: PreuniWeb.ErrorView, accepts: ~w(json)],
  pubsub_server: Preuni.PubSub,
  live_view: [signing_salt: "change_me_in_prod"]

config :preuni, Preuni.Mailer,
  adapter: Swoosh.Adapters.Local

config :preuni, :internal_token, System.get_env("INTERNAL_TOKEN", "dev-internal-token")

config :swoosh, :api_client, Swoosh.ApiClient.Finch

config :preuni, from_email: System.get_env("MAIL_FROM", "noreply@preuni.com.br")
config :preuni, from_name: System.get_env("MAIL_FROM_NAME", "PreUni")

import_config "#{config_env()}.exs"
