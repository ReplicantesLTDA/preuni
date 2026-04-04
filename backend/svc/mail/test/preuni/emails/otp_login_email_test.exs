defmodule Preuni.Emails.OtpLoginEmailTest do
  use ExUnit.Case, async: true

  alias Preuni.Emails.OtpLoginEmail

  @to "student@example.com"

  describe "build/2" do
    test "html body contains the OTP code prominently" do
      email = OtpLoginEmail.build(@to, %{"otp" => "482910"})
      assert email.html_body =~ "482910"
    end

    test "text body contains the OTP code" do
      email = OtpLoginEmail.build(@to, %{"otp" => "482910"})
      assert email.text_body =~ "482910"
    end

    test "subject contains the OTP code" do
      email = OtpLoginEmail.build(@to, %{"otp" => "482910"})
      assert email.subject =~ "482910"
    end

    test "html body contains expiry notice" do
      email = OtpLoginEmail.build(@to, %{"otp" => "111111"})
      assert email.html_body =~ "15 minutos"
    end

    test "text body contains security disclaimer" do
      email = OtpLoginEmail.build(@to, %{"otp" => "111111"})
      assert email.text_body =~ "ignore este e-mail"
    end
  end
end
