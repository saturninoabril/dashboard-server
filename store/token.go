package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/model"
)

type SqlTokenStore struct {
	*SqlStore
}

func newSqlTokenStore(sqlStore *SqlStore) TokenStore {
	s := &SqlTokenStore{
		SqlStore: sqlStore,
	}

	return s
}

func (s *SqlStore) Token() TokenStore {
	return s.stores.token
}

var tokenSelect sq.SelectBuilder

func init() {
	tokenSelect = sq.Select(
		"token",
		"create_at",
		"type",
		"extra",
	)
}

func (s *SqlTokenStore) getTokenTable() string {
	return s.tablePrefix + "token"
}

// CreateToken inserts a new token.
func (s *SqlTokenStore) CreateToken(token *model.Token) (*model.Token, error) {
	err := token.IsValid()
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	_, err = s.execBuilder(
		s.db,
		sq.Insert(s.getTokenTable()).
			SetMap(map[string]interface{}{
				"token":     token.Token,
				"create_at": token.CreateAt,
				"type":      token.Type,
				"extra":     token.Extra,
			}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create token")
	}

	return token, nil
}

// GetToken fetches the given token by token value.
func (s *SqlTokenStore) GetToken(tokenValue string) (*model.Token, error) {
	var token model.Token
	err := s.getBuilder(
		s.db,
		&token,
		tokenSelect.From(s.getTokenTable()).Where("token = ?", tokenValue),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get token by token value")
	}

	return &token, nil
}

// GetTokensByEmail fetches the tokens for the passed email
func (s *SqlTokenStore) GetTokensByEmail(email, tokenType string) ([]*model.Token, error) {
	var tokens []*model.Token
	extraField, err := model.CreateTokenTypeResetPasswordExtra(email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token by email")
	}
	err = s.selectBuilder(
		s.db,
		&tokens,
		tokenSelect.From(s.getTokenTable()).
			Where("extra = ?", extraField).
			Where("type = ?", tokenType),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get token by token value")
	}

	return tokens, nil
}

// DeleteToken deletes a token.
func (s *SqlTokenStore) DeleteToken(tokenValue string) error {
	_, err := s.execBuilder(
		s.db,
		sq.Delete("").From(s.getTokenTable()).Where("token = ?", tokenValue),
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete token")
	}

	return nil
}

// DeleteTokensByEmail deletes all the tokes, of one type, belonging to the
// passed email
func (s *SqlTokenStore) DeleteTokensByEmail(email, tokenType string) error {
	tokens, err := s.GetTokensByEmail(email, tokenType)
	if err != nil {
		return errors.Wrapf(err, "error deleting tokens for email %s", email)
	}
	for _, token := range tokens {
		err := s.DeleteToken(token.Token)
		if err != nil {
			return errors.Wrapf(err, "error deleting token %s for email %s", token.Token, email)
		}
	}
	return nil
}

// CleanupTokenStore removes tokens that are past the defined expiry time.
func (s *SqlTokenStore) CleanupTokenStore(expiryTimeMillis int64) {
	s.logger.Debug("Cleaning up token store.")

	deltime := model.GetMillis() - expiryTimeMillis
	_, err := s.execBuilder(
		s.db,
		sq.Delete("").From(s.getTokenTable()).Where("create_at < ?", deltime),
	)
	if err != nil {
		s.logger.WithError(err).Error("Unable to cleanup token store")
	}
}
