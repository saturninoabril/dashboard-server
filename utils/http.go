package utils

import (
	"net/http"
	"time"

	"github.com/saturninoabril/dashboard-server/model"
)

// GetProtocol will return the protocol used for an HTTP request.
func GetProtocol(r *http.Request) string {
	if r.Header.Get(model.HeaderForwardedProto) == "https" || r.TLS != nil {
		return "https"
	}
	return "http"
}

// IsSecure returns true if the HTTP request is using HTTPS
func IsSecure(r *http.Request) bool {
	return GetProtocol(r) == "https"
}

// AttachSessionCookies will attach the required cookies to a session.
func AttachSessionCookies(w http.ResponseWriter, r *http.Request, session *model.Session) {
	secure := IsSecure(r)

	maxAge := model.SessionLengthMilliseconds / 1000
	expiresAt := time.Unix(model.GetMillis()/1000+int64(maxAge), 0)
	sessionCookie := &http.Cookie{
		Name:     model.SessionCookieToken,
		Path:     "/",
		Value:    session.Token,
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
	}

	csrfCookie := &http.Cookie{
		Name:    model.SessionCookieCSRF,
		Path:    "/",
		Value:   session.CSRFToken,
		MaxAge:  maxAge,
		Expires: expiresAt,
		Secure:  secure,
	}

	userCookie := &http.Cookie{
		Name:    model.SessionCookieUser,
		Path:    "/",
		Value:   session.UserID,
		MaxAge:  maxAge,
		Expires: expiresAt,
		Secure:  secure,
	}

	http.SetCookie(w, sessionCookie)
	http.SetCookie(w, csrfCookie)
	http.SetCookie(w, userCookie)
}

// DeleteSessionCookies will clear all the cookies for the current session.
func DeleteSessionCookies(w http.ResponseWriter, r *http.Request) {
	secure := false
	if GetProtocol(r) == "https" {
		secure = true
	}

	expiresAt := time.Unix(0, 0)

	sessionCookie := &http.Cookie{
		Name:     model.SessionCookieToken,
		Path:     "/",
		Value:    "",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
	}

	csrfCookie := &http.Cookie{
		Name:    model.SessionCookieCSRF,
		Path:    "/",
		Value:   "",
		Expires: expiresAt,
		Secure:  secure,
	}

	userCookie := &http.Cookie{
		Name:    model.SessionCookieUser,
		Path:    "/",
		Value:   "",
		Expires: expiresAt,
		Secure:  secure,
	}

	http.SetCookie(w, sessionCookie)
	http.SetCookie(w, csrfCookie)
	http.SetCookie(w, userCookie)
}
