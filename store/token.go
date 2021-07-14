package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/model"
)

var tokenSelect sq.SelectBuilder

func init() {
	tokenSelect = sq.Select("Token", "CreateAt", "Type", "Extra").From("Tokens")
}

// CreateToken inserts a new token.
func (store *SqlStore) CreateToken(token *model.Token) (*model.Token, error) {
	err := token.IsValid()
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	_, err = store.execBuilder(store.db, sq.
		Insert("Tokens").
		SetMap(map[string]interface{}{
			"Token":    token.Token,
			"CreateAt": token.CreateAt,
			"Type":     token.Type,
			"Extra":    token.Extra,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create token")
	}

	return token, nil
}

// GetToken fetches the given token by token value.
func (store *SqlStore) GetToken(tokenValue string) (*model.Token, error) {
	var token model.Token
	err := store.getBuilder(store.db, &token, tokenSelect.Where("Token = ?", tokenValue))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get token by token value")
	}

	return &token, nil
}

// GetTokensByEmail fetches the tokens for the passed email
func (store *SqlStore) GetTokensByEmail(email, tokenType string) ([]*model.Token, error) {
	var tokens []*model.Token
	extraField, err := model.CreateTokenTypeResetPasswordExtra(email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token by email")
	}
	err = store.selectBuilder(
		store.db,
		&tokens,
		tokenSelect.Where("Extra = ?", extraField).Where("Type = ?", tokenType),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get token by token value")
	}

	return tokens, nil
}

// DeleteToken deletes a token.
func (store *SqlStore) DeleteToken(tokenValue string) error {
	_, err := store.execBuilder(store.db,
		sq.Delete("").From("Tokens").Where("Token = ?", tokenValue),
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete token")
	}

	return nil
}

// DeleteTokensByEmail deletes all the tokes, of one type, belonging to the
// passed email
func (store *SqlStore) DeleteTokensByEmail(email, tokenType string) error {
	tokens, err := store.GetTokensByEmail(email, tokenType)
	if err != nil {
		return errors.Wrapf(err, "error deleting tokens for email %s", email)
	}
	for _, token := range tokens {
		err := store.DeleteToken(token.Token)
		if err != nil {
			return errors.Wrapf(err, "error deleting token %s for email %s", token.Token, email)
		}
	}
	return nil
}

// CleanupTokenStore removes tokens that are past the defined expiry time.
func (store *SqlStore) CleanupTokenStore(expiryTimeMillis int64) {
	store.logger.Debug("Cleaning up token store.")

	deltime := model.GetMillis() - expiryTimeMillis
	_, err := store.execBuilder(store.db,
		sq.Delete("").From("Tokens").Where("CreateAt < ?", deltime),
	)
	if err != nil {
		store.logger.WithError(err).Error("Unable to cleanup token store")
	}
}
