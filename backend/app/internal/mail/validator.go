package mail

import (
	"errors"
	"fmt"
)

// ErrUnknownType is returned when the SendRequest.Type is not in the supported set.
var ErrUnknownType = errors.New("unknown email type")

// Validate checks that the request has the params required for its type.
// Returns a 422-friendly error message on failure.
func Validate(req SendRequest) error {
	if req.To == "" {
		return errors.New("to is required")
	}
	switch req.Type {
	case TypeWelcome:
		if _, ok := req.Params["display_name"]; !ok {
			return errors.New("WELCOME requires display_name and verification_link params")
		}
		if _, ok := req.Params["verification_link"]; !ok {
			return errors.New("WELCOME requires display_name and verification_link params")
		}
	case TypeEmailVerify:
		if _, ok := req.Params["otp"]; !ok {
			return errors.New("EMAIL_VERIFY requires otp param")
		}
	case TypeOTPLogin:
		if _, ok := req.Params["otp"]; !ok {
			return errors.New("OTP_LOGIN requires otp param")
		}
	case TypeEmailChange:
		if _, ok := req.Params["otp"]; !ok {
			return errors.New("EMAIL_CHANGE requires otp param")
		}
	case TypePasswordReset:
		if _, ok := req.Params["otp"]; !ok {
			return errors.New("PASSWORD_RESET requires otp param")
		}
	default:
		return fmt.Errorf("unknown email type: %s", req.Type)
	}
	return nil
}
