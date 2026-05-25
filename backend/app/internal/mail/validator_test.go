package mail

import "testing"

func TestValidate_RequiredParams(t *testing.T) {
	cases := []struct {
		name    string
		req     SendRequest
		wantErr bool
	}{
		{"welcome ok", SendRequest{Type: TypeWelcome, To: "a@b.c", Params: map[string]string{"display_name": "A", "verification_link": "https://x"}}, false},
		{"welcome missing link", SendRequest{Type: TypeWelcome, To: "a@b.c", Params: map[string]string{"display_name": "A"}}, true},
		{"welcome missing name", SendRequest{Type: TypeWelcome, To: "a@b.c", Params: map[string]string{"verification_link": "https://x"}}, true},
		{"email_verify ok", SendRequest{Type: TypeEmailVerify, To: "a@b.c", Params: map[string]string{"otp": "1"}}, false},
		{"email_verify missing", SendRequest{Type: TypeEmailVerify, To: "a@b.c", Params: map[string]string{}}, true},
		{"otp_login ok", SendRequest{Type: TypeOTPLogin, To: "a@b.c", Params: map[string]string{"otp": "1"}}, false},
		{"otp_login missing", SendRequest{Type: TypeOTPLogin, To: "a@b.c", Params: nil}, true},
		{"email_change ok", SendRequest{Type: TypeEmailChange, To: "a@b.c", Params: map[string]string{"otp": "1"}}, false},
		{"email_change missing", SendRequest{Type: TypeEmailChange, To: "a@b.c", Params: nil}, true},
		{"password_reset ok", SendRequest{Type: TypePasswordReset, To: "a@b.c", Params: map[string]string{"otp": "1"}}, false},
		{"password_reset missing", SendRequest{Type: TypePasswordReset, To: "a@b.c", Params: nil}, true},
		{"unknown type", SendRequest{Type: "BOGUS", To: "a@b.c"}, true},
		{"missing to", SendRequest{Type: TypeEmailVerify, Params: map[string]string{"otp": "1"}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Validate(c.req)
			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuild_AllTypes(t *testing.T) {
	cases := []SendRequest{
		{Type: TypeWelcome, To: "a@b.c", Params: map[string]string{"display_name": "Ana", "verification_link": "https://verify"}},
		{Type: TypeEmailVerify, To: "a@b.c", Params: map[string]string{"otp": "123456"}},
		{Type: TypeOTPLogin, To: "a@b.c", Params: map[string]string{"otp": "123456"}},
		{Type: TypeEmailChange, To: "new@b.c", Params: map[string]string{"otp": "123456"}},
		{Type: TypePasswordReset, To: "a@b.c", Params: map[string]string{"otp": "123456"}},
	}
	for _, c := range cases {
		msg, err := Build(c, "noreply@preuni.com.br", "PreUni")
		if err != nil {
			t.Fatalf("%s: build failed: %v", c.Type, err)
		}
		if msg.Subject == "" || msg.HTMLBody == "" || msg.TextBody == "" {
			t.Fatalf("%s: empty fields in message %+v", c.Type, msg)
		}
	}
}
