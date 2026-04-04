defmodule Preuni.Emails.WelcomeEmailTest do
  use ExUnit.Case, async: true

  alias Preuni.Emails.WelcomeEmail

  @to "student@example.com"
  @params %{
    "display_name" => "Ana Lima",
    "verification_link" => "https://preuni.com.br/verify?token=abc123"
  }

  describe "build/2" do
    test "returns a Swoosh.Email struct" do
      email = WelcomeEmail.build(@to, @params)
      assert %Swoosh.Email{} = email
    end

    test "sets the correct recipient" do
      email = WelcomeEmail.build(@to, @params)
      assert {"", @to} in email.to
    end

    test "subject contains display name" do
      email = WelcomeEmail.build(@to, @params)
      assert email.subject =~ "Ana Lima"
    end

    test "html body contains greeting with display name" do
      email = WelcomeEmail.build(@to, @params)
      assert email.html_body =~ "Ana Lima"
    end

    test "html body contains verification link" do
      email = WelcomeEmail.build(@to, @params)
      assert email.html_body =~ "https://preuni.com.br/verify?token=abc123"
    end

    test "text body contains verification link" do
      email = WelcomeEmail.build(@to, @params)
      assert email.text_body =~ "https://preuni.com.br/verify?token=abc123"
    end
  end
end
