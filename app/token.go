package app

import (
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/model"
)

// CreateAndStoreVerifyEmailToken creates a new verify email token and stores it.
func (a *App) CreateAndStoreVerifyEmailToken(email string) (*model.Token, error) {
	return a.createAndStoreEmailToken(model.TokenTypeVerifyEmail, email)
}

// CreateAndStoreResetPasswordToken creates a new password reset token and
// stores it.
func (a *App) CreateAndStoreResetPasswordToken(email string) (*model.Token, error) {
	return a.createAndStoreEmailToken(model.TokenTypeResetPassword, email)
}

func (a *App) createAndStoreEmailToken(tokenType, email string) (*model.Token, error) {
	err := a.store.Token().DeleteTokensByEmail(email, tokenType)
	if err != nil {
		return nil, errors.Wrap(err, "error cleaning previous tokens")
	}
	extra, err := model.CreateTokenTypeResetPasswordExtra(email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create token extra value")
	}
	token := model.NewToken(tokenType, extra)
	err = token.IsValid()
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}
	token, err = a.store.Token().CreateToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "unable to save new token")
	}

	return token, nil
}
