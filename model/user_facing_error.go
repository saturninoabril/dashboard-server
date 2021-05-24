package model

import "fmt"

// UserFacingError is an error type meant to be read by end users.
type UserFacingError struct {
	messageForUser  string
	internalDetails string
}

// NewUserFacingError creates a new error including a user facing message.
func NewUserFacingError(messageForUser, internalDetails string) *UserFacingError {
	return &UserFacingError{
		messageForUser:  messageForUser,
		internalDetails: internalDetails,
	}
}

func (u *UserFacingError) Error() string {
	return fmt.Sprintf("messageForUser=%s details=%s", u.messageForUser, u.internalDetails)
}

// ErrorForUser returns the error message meant for the end user.
func (u *UserFacingError) ErrorForUser() string {
	return u.messageForUser
}
