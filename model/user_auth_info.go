package model

import "golang.org/x/oauth2"

type UserAuthInfo struct {
	UserID        string       `json:"user_id"`
	OAuthProvider string       `json:"oauth_provider"`
	Token         oauth2.Token `json:"token"`
	Username      string       `json:"username"`
	Email         string       `json:"email,omitempty"`
	Name          string       `json:"name,omitempty"`
	AvatarURL     string       `json:"avatar_url"`
	CreateAt      int64        `json:"create_at" db:"create_at"`
	UpdateAt      int64        `json:"update_at" db:"update_at"`
	DeleteAt      int64        `json:"update_at" db:"delete_at"`
}

// PreSave will set the ID and CreateAt for the UserAuthInfo.
func (u *UserAuthInfo) PreSave() {
	now := GetMillis()
	u.CreateAt = now
	u.UpdateAt = now
	u.DeleteAt = 0
}
