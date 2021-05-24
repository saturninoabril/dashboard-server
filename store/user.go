package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/saturninoabril/dashboard-server/model"
)

var userSelect sq.SelectBuilder

func init() {
	userSelect = sq.
		Select("ID", "CreateAt", "UpdateAt", "Email", "EmailVerified", "Password",
			"FirstName", "LastName", "State").
		From("Users")
}

// CreateUser inserts a new user.
func (store *DashboardStore) CreateUser(user *model.User) (*model.User, error) {
	user.CreatePreSave()

	if err := user.IsValid(); err != nil {
		return nil, err
	}

	_, err := store.execBuilder(store.db, sq.
		Insert("Users").
		SetMap(map[string]interface{}{
			"ID":            user.ID,
			"CreateAt":      user.CreateAt,
			"UpdateAt":      user.UpdateAt,
			"Email":         user.Email,
			"EmailVerified": user.EmailVerified,
			"Password":      user.Password,
			"FirstName":     user.FirstName,
			"LastName":      user.LastName,
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
func (store *DashboardStore) GetUser(id string) (*model.User, error) {
	var user model.User
	err := store.getBuilder(store.db, &user, userSelect.Where("ID = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get user by id")
	}

	return &user, nil
}

// GetUserByEmail fetches the given user by email.
func (store *DashboardStore) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := store.getBuilder(store.db, &user, userSelect.Where("Email = ?", email))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get user by email")
	}

	return &user, nil
}

// VerifyEmail updates a user's email and marks it as verified.
func (store *DashboardStore) VerifyEmail(id, email string) error {
	currentTime := model.GetMillis()
	_, err := store.execBuilder(store.db,
		sq.Update("").Table("Users").Where("ID = ?", id).
			Set("UpdateAt", currentTime).
			Set("EmailVerified", true).
			Set("Email", email),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set user email as verified")
	}

	return nil
}

// UnverifyEmail updates a user's email and marks it as unverified.
func (store *DashboardStore) UnverifyEmail(id, email string) error {
	currentTime := model.GetMillis()
	_, err := store.execBuilder(store.db,
		sq.Update("").Table("Users").Where("ID = ?", id).
			Set("UpdateAt", currentTime).
			Set("EmailVerified", false).
			Set("Email", email),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set user email as unverified")
	}

	return nil
}

// UpdatePassword accepts a plaintext password value, hashes it, and saves it
// as the new password for a given user.
func (store *DashboardStore) UpdatePassword(id, password string) error {
	hashedPassword := model.HashPassword(password)
	currentTime := model.GetMillis()
	_, err := store.execBuilder(store.db,
		sq.Update("").Table("Users").Where("ID = ?", id).
			Set("UpdateAt", currentTime).
			Set("Password", hashedPassword),
	)
	if err != nil {
		return errors.Wrap(err, "failed to set user email as verified")
	}

	return nil
}

// UpdateUser updates the given user.
func (store *DashboardStore) UpdateUser(user *model.User) error {
	if err := user.IsValid(); err != nil {
		return err
	}
	_, err := store.execBuilder(store.db, sq.
		Update("Users").
		SetMap(map[string]interface{}{
			"Email":     user.Email,
			"FirstName": user.FirstName,
			"LastName":  user.LastName,
		}).
		Where("ID = ?", user.ID),
	)
	if err != nil {
		if isUniqueConstraintError(err, []string{"Email", "email"}) {
			return errors.New("account exists")
		}
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}
