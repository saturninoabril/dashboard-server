package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/saturninoabril/dashboard-server/model"
)

var sessionSelect sq.SelectBuilder

func init() {
	sessionSelect = sq.
		Select("ID", "Token", "CreateAt", "ExpiresAt", "UserID", "CSRFToken").
		From("Session")
}

// CreateSession inserts a new session.
func (store *SqlStore) CreateSession(session *model.Session) (*model.Session, error) {
	session.PreSave()

	_, err := store.execBuilder(store.db, sq.
		Insert("Session").
		SetMap(map[string]interface{}{
			"ID":        session.ID,
			"Token":     session.Token,
			"CreateAt":  session.CreateAt,
			"ExpiresAt": session.ExpiresAt,
			"UserID":    session.UserID,
			"CSRFToken": session.CSRFToken,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return session, nil
}

// GetSession fetches the given session by id or token. Deletes and does not return expired sessions.
func (store *SqlStore) GetSession(idOrToken string) (*model.Session, error) {
	var session model.Session
	err := store.getBuilder(store.db, &session, sessionSelect.Where(sq.Or{sq.Eq{"ID": idOrToken}, sq.Eq{"Token": idOrToken}}))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get session by id or token")
	}

	if session.IsExpired() {
		_, err = store.execBuilder(store.db, sq.Delete("Session").Where("ID = ?", session.ID))
		if err != nil {
			store.logger.
				WithField("type", "session").
				WithField("id", session.ID).
				WithError(err).
				Error("unable to delete expired session")
		}
		return nil, nil
	}

	return &session, nil
}

// DeleteSession deletes a session by ID.
func (store *SqlStore) DeleteSession(id string) error {
	_, err := store.execBuilder(store.db, sq.Delete("Session").Where("ID = ?", id))
	if err != nil {
		store.logger.
			WithField("type", "session").
			WithField("id", id).
			WithError(err).
			Error("unable to delete session")
		return err
	}

	return nil
}

// DeleteSessionsForUser deletes all the sessions for a user
func (store *SqlStore) DeleteSessionsForUser(userID string) error {
	_, err := store.execBuilder(store.db, sq.Delete("Session").Where("UserID = ?", userID))
	if err != nil {
		store.logger.
			WithField("type", "session").
			WithField("userID", userID).
			WithError(err).
			Error("unable to delete user sessions")
		return err
	}

	return nil
}
