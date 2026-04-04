defmodule PreuniWeb.Plugs.InternalAuth do
  import Plug.Conn

  @doc """
  Validates the `Authorization: Bearer {token}` header against the
  configured INTERNAL_TOKEN env var. Rejects with 401 if missing or wrong.
  """
  def init(opts), do: opts

  def call(conn, _opts) do
    expected = Application.get_env(:preuni, :internal_token)

    case get_req_header(conn, "authorization") do
      ["Bearer " <> token] when token == expected ->
        conn

      _ ->
        conn
        |> send_resp(401, Jason.encode!(%{error: "unauthorized"}))
        |> halt()
    end
  end
end
