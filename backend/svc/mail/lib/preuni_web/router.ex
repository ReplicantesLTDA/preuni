defmodule PreuniWeb.Router do
  use Phoenix.Router

  pipeline :internal_api do
    plug :accepts, ["json"]
    plug PreuniWeb.Plugs.InternalAuth
  end

  scope "/internal", PreuniWeb do
    pipe_through :internal_api

    post "/email/send", EmailController, :send
  end

  scope "/health" do
    get "/", PreuniWeb.HealthController, :check
  end

  # Swoosh local mailbox preview — only active when SWOOSH_ADAPTER is "local" (dev/test).
  # Access at http://localhost:4000/dev/mailbox to inspect sent emails.
  if System.get_env("SWOOSH_ADAPTER", "local") == "local" do
    forward "/dev/mailbox", Plug.Swoosh.MailboxPreview
  end
end
