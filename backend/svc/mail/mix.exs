defmodule Preuni.MixProject do
  use Mix.Project

  def project do
    [
      app: :preuni,
      version: "0.1.0",
      elixir: "~> 1.17",
      elixirc_paths: elixirc_paths(Mix.env()),
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  def application do
    [
      mod: {Preuni.Application, []},
      extra_applications: [:logger, :runtime_tools]
    ]
  end

  defp elixirc_paths(:test), do: ["lib", "test/support"]
  defp elixirc_paths(_), do: ["lib"]

  defp deps do
    [
      {:phoenix, "~> 1.7"},
      {:phoenix_html, "~> 4.0"},
      {:swoosh, "~> 1.17"},
      {:gen_smtp, "~> 1.2"},
      {:finch, "~> 0.18"},
      {:jason, "~> 1.4"},
      {:plug_cowboy, "~> 2.7"},
      {:mox, "~> 1.1", only: :test}
    ]
  end
end
