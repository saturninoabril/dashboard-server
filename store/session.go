package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/saturninoabril/dashboard-server/model"
)

type SqlSessionStore struct {
	*SqlStore
}

func newSqlSessionStore(sqlStore *SqlStore) SessionStore {
	s := &SqlSessionStore{
		SqlStore: sqlStore,
	}

	return s
}

func (s *SqlStore) Session() SessionStore {
	return s.stores.session
}

var sessionSelect sq.SelectBuilder

func init() {
	sessionSelect = sq.
		Select(
			"id",
			"token",
			"create_at",
			"expires_at",
			"user_id",
			"csrf_token",
		)
}

func (s *SqlSessionStore) getSessionTable() string {
	return s.tablePrefix + "session"
}

// CreateSession inserts a new session.
func (s *SqlSessionStore) CreateSession(session *model.Session) (*model.Session, error) {
	session.PreSave()

	_, err := s.execBuilder(s.db, sq.
		Insert(s.getSessionTable()).
		SetMap(map[string]interface{}{
			"id":         session.ID,
			"token":      session.Token,
			"create_at":  session.CreateAt,
			"expires_at": session.ExpiresAt,
			"user_id":    session.UserID,
			"csrf_token": session.CSRFToken,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return session, nil
}

// GetSession fetches the given session by id or token. Does not return expired sessions.
func (s *SqlSessionStore) GetSession(idOrToken string) (*model.Session, error) {
	sessionTable := s.getSessionTable()

	var session model.Session
	err := s.getBuilder(
		s.db,
		&session,
		sessionSelect.From(sessionTable).
			Where(sq.Or{sq.Eq{"id": idOrToken}, sq.Eq{"token": idOrToken}}),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get session by id or token")
	}

	if session.IsExpired() {
		_, err = s.execBuilder(
			s.db,
			sq.Delete(sessionTable).Where("id = ?", session.ID),
		)
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
func (s *SqlSessionStore) DeleteSession(id string) error {
	_, err := s.execBuilder(
		s.db,
		sq.Delete(s.getSessionTable()).Where("id = ?", id),
	)
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
func (s *SqlSessionStore) DeleteSessionsForUser(userID string) error {
	_, err := s.execBuilder(
		s.db,
		sq.Delete(s.getSessionTable()).
			Where("user_id = ?", userID),
	)
	if err != nil {
		s.logger.
			WithField("type", "session").
			WithField("user_id", userID).
			WithError(err).
			Error("unable to delete user sessions")
		return err
	}

	return nil
}
