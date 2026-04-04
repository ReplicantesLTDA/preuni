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
end
