package model

const (
	// SessionCookieToken is the cookie key for the authorization token.
	SessionCookieToken = "DASHBOARDAUTHTOKEN"
	// SessionCookieCSRF is the cookie key for the CSRF token.
	SessionCookieCSRF = "DASHBOARDCSRF"
	// SessionCookieUser is the cookie key for the logged in user.
	SessionCookieUser = "DASHBOARDUSERID"
	// SessionHeader is the header key for a session.
	SessionHeader = "token"
	// SessionLengthMilliseconds is the session length in milliseconds.
	SessionLengthMilliseconds = 1000 * 60 * 60 * 24 * 15 // 15 days
	// HeaderRequestedWith is the HTTP header X-Requested-With.
	HeaderRequestedWith = "X-Requested-With"
	// HeaderRequestedWithXML is the HTTP header value XMLHttpRequest.
	HeaderRequestedWithXML = "XMLHttpRequest"
	// HeaderForwardedProto is the HTTP header X-Forwarded-Proto.
	HeaderForwardedProto = "X-Forwarded-Proto"
	// HeaderRequestID is the custom header to track request ID.
	HeaderRequestID = "X-Request-ID"
	// HeaderAuthorization is the HTTP header Authorization.
	HeaderAuthorization = "Authorization"
	// HeaderApiKey is the HTTP header containing API key.
	HeaderApiKey = "X-CTRL-Api-Key"
	// HeaderCSRFToken is the HTTP header for holding the CSRF token.
	HeaderCSRFToken = "X-CSRF-Token"
	// AuthorizationBearer is the bearer HTTP authorization type.
	AuthorizationBearer = "BEARER"
)

// Session is an authentication session for a user.
type Session struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	CreateAt  int64  `json:"create_at" db:"create_at"`
	ExpiresAt int64  `json:"expires_at" db:"expires_at"`
	UserID    string `json:"user_id" db:"user_id"`
	CSRFToken string `json:"csrf_token" db:"csrf_token"`
}

// PreSave will set the ID, Token, CSRFToken, ExpiresAt and CreateAt for the session.
func (s *Session) PreSave() {
	if s.ID == "" {
		s.ID = NewID()
	}
	if s.Token == "" {
		s.Token = NewID()
	}
	if s.CSRFToken == "" {
		s.CSRFToken = NewID()
	}
	s.CreateAt = GetMillis()
	s.ExpiresAt = s.CreateAt + SessionLengthMilliseconds
}

// IsExpired returns true if the session is expired.
func (s *Session) IsExpired() bool {
	if s.ExpiresAt <= 0 {
		return false
	}

	if GetMillis() > s.ExpiresAt {
		return true
	}

	return false
}
