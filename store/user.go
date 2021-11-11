package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/saturninoabril/dashboard-server/model"
)

type SqlUserStore struct {
	*SqlStore
}

func newSqlUserStore(sqlStore *SqlStore) UserStore {
	s := &SqlUserStore{
		SqlStore: sqlStore,
	}

	return s
}

func (s *SqlStore) User() UserStore {
	return s.stores.user
}

var userSelect sq.SelectBuilder

func init() {
	userSelect = sq.
		Select(
			"id",
			"create_at",
			"update_at",
			"email",
			"email_verified",
			"password",
			"first_name",
			"last_name",
			"State",
		)
}

func (s *SqlUserStore) getUserTable() string {
	return s.tablePrefix + "user"
}

// CreateUser inserts a new user.
func (s *SqlUserStore) CreateUser(user *model.User) (*model.User, error) {
	user.CreatePreSave()

	if err := user.IsValid(); err != nil {
		return nil, err
	}

	_, err := s.execBuilder(s.db, sq.
		Insert(s.getUserTable()).
		SetMap(map[string]interface{}{
			"id":             user.ID,
			"create_at":      user.CreateAt,
			"update_at":      user.UpdateAt,
			"email":          user.Email,
			"email_verified": user.EmailVerified,
			"password":       user.Password,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
		}),
	)
	if err != nil {
		if isUniqueConstraintError(err, []string{"Email", "email"}) {
			return nil, errors.New("email exists")
		}
		return nil, errors.Wrap(err, "failed to create user")
	}

	return user, nil
}

// GetUser fetches the given user by id.
func (s *SqlUserStore) GetUser(id string) (*model.User, error) {
	var user model.User
	err := s.getBuilder(
		s.db,
		&user,
		userSelect.From(s.getUserTable()).Where("id = ?", id),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get user by id")
	}

	return &user, nil
}

// GetUserByEmail fetches the given user by email.
func (s *SqlUserStore) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := s.getBuilder(
		s.db,
		&user,
		userSelect.From(s.getUserTable()).Where("email = ?", email),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get user by email")
	}

	return &user, nil
}

// VerifyEmail updates a user's email and marks it as verified.
func (s *SqlUserStore) VerifyEmail(id, email string) error {
	currentTime := model.GetMillis()
	_, err := s.execBuilder(
		s.db,
		sq.Update("").Table(s.getUserTable()).Where("id = ?", id).
			Set("update_at", currentTime).
			Set("email_verified", true).
			Set("email", email),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set user email as verified")
	}

	return nil
}

// UnverifyEmail updates a user's email and marks it as unverified.
func (s *SqlUserStore) UnverifyEmail(id, email string) error {
	currentTime := model.GetMillis()
	_, err := s.execBuilder(
		s.db,
		sq.Update("").Table(s.getUserTable()).
			Where("id = ?", id).
			Set("update_at", currentTime).
			Set("email_verified", false).
			Set("email", email),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set user email as unverified")
	}

	return nil
}

// UpdatePassword accepts a plaintext password value, hashes it, and saves it
// as the new password for a given user.
func (s *SqlUserStore) UpdatePassword(id, password string) error {
	hashedPassword := model.HashPassword(password)
	currentTime := model.GetMillis()
	_, err := s.execBuilder(
		s.db,
		sq.Update("").Table(s.getUserTable()).
			Where("id = ?", id).
			Set("update_at", currentTime).
			Set("password", hashedPassword),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set user email as verified")
	}

	return nil
}

// UpdateUser updates the given user.
func (s *SqlUserStore) UpdateUser(user *model.User) error {
	if err := user.IsValid(); err != nil {
		return err
	}
	_, err := s.execBuilder(
		s.db,
		sq.Update(s.getUserTable()).
			Where("id = ?", user.ID).
			SetMap(map[string]interface{}{
				"email":      user.Email,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
			}),
	)
	if err != nil {
		if isUniqueConstraintError(err, []string{"Email", "email"}) {
			return errors.New("account exists")
		}
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}

func (s *SqlUserStore) UpdateUserState(userID string, state string) error {
	currentTime := model.GetMillis()
	_, err := s.execBuilder(
		s.db,
		sq.Update("").Table(s.getUserTable()).
			Where("id = ?", userID).
			Set("update_at", currentTime).
			Set("state", state),
	)
	if err != nil {
		return errors.Wrap(err, "failed to update user state")
	}

	return nil
}
