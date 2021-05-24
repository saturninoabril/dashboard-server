package model

import (
	"encoding/json"
	"io"
)

// SignUpRequest contains the fields for a request to the sign-up API.
type SignUpRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// SignUpResponse contains the user.
type SignUpResponse struct {
	User *User `json:"user"`
}

// SignUpResponseFromReader decodes a json-encoded sign-up API response from the given io.Reader.
func SignUpResponseFromReader(reader io.Reader) (*SignUpResponse, error) {
	sr := SignUpResponse{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&sr)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &sr, nil
}

// LoginRequest specifies the login parameters.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ForgotPasswordRequest specifies the forgot password parameters.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest specifies the reset password parameters.
type ResetPasswordRequest struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

// UpdatePasswordRequest specifies the update password parameters.
type UpdatePasswordRequest struct {
	NewPassword     string `json:"new_password"`
	CurrentPassword string `json:"current_password"`
}

// VerifyEmailRequest specifies the verify-email parameters.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}
