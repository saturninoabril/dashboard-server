package model

import (
	"encoding/json"
	"errors"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	userNameMaxRuneLength          = 64
	userPasswordMaxLength          = 72
	userPasswordMinLength          = 8
	userEmailMaxLength             = 128
	EventTypeSendAdminWelcomeEmail = "send-admin-welcome-email"
)

const (
	// UserStateActive means the user is fully active and no restrictions exist
	UserStateActive = "active"
	// UserStateLocked means the user should not be able to access the dashboard
	UserStateLocked = "locked"
)

// User model represents a user on the system.
type User struct {
	ID            string `json:"id"`
	CreateAt      int64  `json:"create_at" db:"create_at"`
	UpdateAt      int64  `json:"update_at" db:"update_at"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified" db:"email_verified"`
	Password      string `json:"password,omitempty"`
	FirstName     string `json:"first_name" db:"first_name"`
	LastName      string `json:"last_name" db:"last_name"`
	State         string `json:"state"`
	IsAdmin       bool   `json:"is_admin" db:"is_admin"`
}

// IsValid will determine if the user fields are all valid.
func (u *User) IsValid() error {
	if len(u.ID) != 26 {
		return errors.New("invalid id")
	}
	if len(u.Email) > userEmailMaxLength || len(u.Email) == 0 || !IsValidEmail(u.Email) {
		return errors.New("invalid email")
	}
	if utf8.RuneCountInString(u.FirstName) > userNameMaxRuneLength {
		return errors.New("invalid first name")
	}
	if utf8.RuneCountInString(u.LastName) > userNameMaxRuneLength {
		return errors.New("invalid last name")
	}
	return nil
}

// IsValidPassword returns an error if the user's password is not valid.
func (u *User) IsValidPassword() error {
	if !IsValidPassword(u.Password) {
		return errors.New("invalid password")
	}

	return nil
}

// CreatePreSave will set the correct values for a new user that is about to be
// saved.
func (u *User) CreatePreSave() {
	if u.ID == "" {
		u.ID = NewID()
	}

	now := GetMillis()
	u.CreateAt = now
	u.UpdateAt = now

	u.EmailVerified = false

	u.Password = HashPassword(u.Password)
}

// Sanitize clears any sensitive data from the user.
func (u *User) Sanitize() {
	u.Password = ""
}

func (u *User) GetUserEmailDomain() string {
	idx := strings.Index(u.Email, "@")
	return u.Email[idx+1:]
}

// HashPassword generates a hash using the bcrypt.GenerateFromPassword
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

// ComparePassword compares the hash to the password.
func ComparePassword(hash string, password string) bool {
	if len(password) == 0 || len(hash) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsLower will check if a string is lowercased
func IsLower(s string) bool {
	return strings.ToLower(s) == s
}

// IsValidEmail will perform some minimal validations on whether the provided string is a valid email address.
func IsValidEmail(email string) bool {
	if !IsLower(email) {
		return false
	}

	if addr, err := mail.ParseAddress(email); err != nil {
		return false
	} else if addr.Name != "" {
		// mail.ParseAddress accepts input of the form "Billy Bob <billy@example.com>" which we don't allow
		return false
	}

	return true
}

// IsValidPassword will perform some minimal validations on whether the provided
// password meets our requirements.
func IsValidPassword(password string) bool {
	if len(password) == 0 || len(password) > userPasswordMaxLength || len(password) < userPasswordMinLength {
		return false
	}

	if ok, _ := regexp.MatchString(".*[a-z].*", password); !ok {
		return false
	}

	if ok, _ := regexp.MatchString(".*[A-Z].*", password); !ok {
		return false
	}

	if ok, _ := regexp.MatchString(".*[0-9].*", password); !ok {
		return false
	}

	return true
}

// UserFromReader decodes a json-encoded user from the given io.Reader.
func UserFromReader(reader io.Reader) (*User, error) {
	user := User{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&user)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &user, nil
}
