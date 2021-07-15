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
		Select(
			"ID",
			"Token",
			"CreateAt",
			"ExpiresAt",
			"UserID",
			"CSRFToken",
		)
}

// CreateSession inserts a new session.
func (s *SqlStore) CreateSession(session *model.Session) (*model.Session, error) {
	session.PreSave()

	_, err := s.execBuilder(s.db, sq.
		Insert(s.tablePrefix+"Session").
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
func (s *SqlStore) GetSession(idOrToken string) (*model.Session, error) {
	var session model.Session
	err := s.getBuilder(s.db, &session, sessionSelect.From(s.tablePrefix+"Session").Where(sq.Or{sq.Eq{"ID": idOrToken}, sq.Eq{"Token": idOrToken}}))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get session by id or token")
	}

	if session.IsExpired() {
		_, err = s.execBuilder(s.db, sq.Delete("Session").Where("ID = ?", session.ID))
		if err != nil {
			s.logger.
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
func (s *SqlStore) DeleteSession(id string) error {
	_, err := s.execBuilder(s.db, sq.Delete(s.tablePrefix+"Session").Where("ID = ?", id))
	if err != nil {
		s.logger.
			WithField("type", "session").
			WithField("id", id).
			WithError(err).
			Error("unable to delete session")
		return err
	}

	return nil
}

// DeleteSessionsForUser deletes all the sessions for a user
func (s *SqlStore) DeleteSessionsForUser(userID string) error {
	_, err := s.execBuilder(s.db, sq.Delete(s.tablePrefix+"Session").Where("UserID = ?", userID))
	if err != nil {
		s.logger.
			WithField("type", "session").
			WithField("userID", userID).
			WithError(err).
			Error("unable to delete user sessions")
		return err
	}

	return nil
}
