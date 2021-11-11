package model

const (
	// DefaultTokenTTL is the OAuth state expiry duration in milliseconds
	DefaultTokenTTL = 10 * 60
)

type OAuthState struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	CreateAt  int64  `json:"create_at" db:"create_at"`
	ExpiresAt int64  `json:"expires_at" db:"expires_at"`
}

// PreSave will set the ID, Token, ExpiresAt and CreateAt for the OAuth state.
func (o *OAuthState) PreSave() {
	if o.ID == "" {
		o.ID = NewID()
	}
	if o.Token == "" {
		o.Token = NewID()
	}
	o.CreateAt = GetMillis()
	o.ExpiresAt = o.CreateAt + DefaultTokenTTL
}

// IsExpired returns true if the OAuth state is expired.
func (o *OAuthState) IsExpired() bool {
	if o.ExpiresAt <= 0 {
		return false
	}

	if GetMillis() > o.ExpiresAt {
		return true
	}

	return false
}
