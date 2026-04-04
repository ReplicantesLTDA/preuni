defmodule Preuni.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      PreuniWeb.Endpoint,
      {Finch, name: Swoosh.Finch}
    ]

    opts = [strategy: :one_for_one, name: Preuni.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
