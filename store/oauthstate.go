package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/saturninoabril/dashboard-server/model"
)

type SqlOAuthStateStore struct {
	*SqlStore
}

func newSqlOAuthStateStore(sqlStore *SqlStore) OAuthStateStore {
	s := &SqlOAuthStateStore{
		SqlStore: sqlStore,
	}

	return s
}

func (s *SqlStore) OAuthState() OAuthStateStore {
	return s.stores.oauthState
}

var oauthStateSelect sq.SelectBuilder

func init() {
	oauthStateSelect = sq.
		Select(
			"id",
			"token",
			"create_at",
			"expires_at",
		)
}

func (s *SqlOAuthStateStore) getOAuthStateTable() string {
	return s.tablePrefix + "oauthstate"
}

// CreateOAuthState inserts a new OAuth state.
func (s *SqlOAuthStateStore) CreateOAuthState() (*model.OAuthState, error) {
	oauthState := &model.OAuthState{}
	oauthState.PreSave()

	_, err := s.execBuilder(s.db, sq.
		Insert(s.getOAuthStateTable()).
		SetMap(map[string]interface{}{
			"id":         oauthState.ID,
			"token":      oauthState.Token,
			"create_at":  oauthState.CreateAt,
			"expires_at": oauthState.ExpiresAt,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oauth state")
	}

	return oauthState, nil
}

// GetOAuthState fetches the given OAuth state by id or token. Does not return expired state.
func (s *SqlOAuthStateStore) GetOAuthState(idOrToken string) (*model.OAuthState, error) {
	outhStateTable := s.getOAuthStateTable()
	var oauthState model.OAuthState
	err := s.getBuilder(
		s.db,
		&oauthState,
		oauthStateSelect.From(outhStateTable).
			Where(sq.Or{sq.Eq{"id": idOrToken}, sq.Eq{"token": idOrToken}}),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get oauth state by id or token")
	}

	if oauthState.IsExpired() {
		_, err = s.execBuilder(
			s.db,
			sq.Delete(outhStateTable).Where("id = ?", oauthState.ID),
		)
		if err != nil {
			s.logger.
				WithField("type", "oauthstate").
				WithField("id", oauthState.ID).
				WithError(err).
				Error("unable to delete expired oauth state")
		}
		return nil, nil
	}

	return &oauthState, nil
}

// DeleteOAuthState deletes a oauth state by ID.
func (s *SqlOAuthStateStore) DeleteOAuthState(id string) error {
	_, err := s.execBuilder(
		s.db,
		sq.Delete(s.getOAuthStateTable()).Where("id = ?", id),
	)
	if err != nil {
		s.logger.
			WithField("type", "oauthstate").
			WithField("id", id).
			WithError(err).
			Error("unable to delete oauth state")
		return err
	}

	return nil
}
