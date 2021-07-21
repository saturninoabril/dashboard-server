package model

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	// TokenSize is the size of the random value of a token.
	TokenSize = 64
	// TokenSizeDigits is the size of the smaller token used during signup.
	TokenSizeDigits = 6
	// TokenTypeVerifyEmail is the token type used for user email verification.
	TokenTypeVerifyEmail = "verify_email"
	// TokenTypeResetPassword the token type used for resetting user passwords.
	TokenTypeResetPassword = "reset_password"
	// TokenDefaultExpiryTime is the default time for tokens to expire.
	TokenDefaultExpiryTime = 1000 * 60 * 60 * 24 // 24 hour
)

// Token is a one-time-use expiring object used in place of passwords or other
// authentication values.
type Token struct {
	Token    string `json:"token"`
	CreateAt int64  `json:"create_at" db:"create_at"`
	Type     string `json:"type"`
	Extra    string `json:"extra"`
}

// NewToken returns a new token.
func NewToken(tokentype, extra string) *Token {
	var token string
	if tokentype == TokenTypeVerifyEmail {
		token = NewRandomNumber(TokenSizeDigits)
	} else {
		token = NewRandomString(TokenSize)
	}
	return &Token{
		Token:    token,
		CreateAt: GetMillis(),
		Type:     tokentype,
		Extra:    extra,
	}
}

// IsValid checks a token for valid configuration
func (t *Token) IsValid() error {
	if t.Type != TokenTypeResetPassword && t.Type != TokenTypeVerifyEmail {
		return errors.Errorf("unsupported token type: (%s)", t.Type)
	}
	if t.Type == TokenTypeResetPassword && len(t.Token) != TokenSize {
		return errors.Errorf("token length (%d) was expected to be %d", len(t.Token), TokenSize)
	}
	if t.Type == TokenTypeVerifyEmail && len(t.Token) != TokenSizeDigits {
		return errors.Errorf("token length (%d) was expected to be %d", len(t.Token), TokenSizeDigits)
	}
	if t.CreateAt == 0 {
		return errors.New("token CreateAt value is not set")
	}
	if t.Type != TokenTypeVerifyEmail && t.Type != TokenTypeResetPassword {
		return errors.Errorf("token type %s is invalid", t.Type)
	}

	return nil
}

// GetExtraEmail returns the correct extra values for a token of type
// TokenTypeChangePassword.
func (t *Token) GetExtraEmail() (string, error) {
	var extra TokenExtraEmail
	err := json.Unmarshal([]byte(t.Extra), &extra)
	if err != nil {
		return "", errors.Wrap(err, "unable to ummarshal extra field")
	}
	if len(extra.Email) == 0 {
		return "", errors.New("email value is empty")
	}

	return extra.Email, nil
}

// TokenExtraEmail is a token extra field containing an email value.
type TokenExtraEmail struct {
	Email string `json:"email"`
}

// CreateTokenTypeResetPasswordExtra returns the correct extra values for a
// token of type TokenTypeResetPassword.
func CreateTokenTypeResetPasswordExtra(email string) (string, error) {
	b, err := json.Marshal(TokenExtraEmail{Email: email})
	if err != nil {
		_ = errors.Wrap(err, "unable to marshal extra field")
	}

	return string(b), nil
}
