package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/saturninoabril/dashboard-server/internal/web"
	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/utils"
)

type contextHandlerFunc func(c *Context, w http.ResponseWriter, r *http.Request)

var _ http.Handler = contextHandler{}

type contextHandler struct {
	context              *Context
	handler              contextHandlerFunc
	requiresSession      bool
	requiresVerification bool
	requiresAdmin        bool
	allowApiKeySession   bool
	allowJWTTokenSession bool
	isStatic             bool
}

func (h contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w = web.NewWrappedWriter(w)
	context := h.context.Clone()
	context.RequestID = model.NewID()
	context.Logger = context.Logger.WithFields(logrus.Fields{
		"path":       r.URL.Path,
		"request_id": context.RequestID,
	})
	defer func() {
		statusCode := strconv.Itoa(w.(*web.ResponseWriterWrapper).StatusCode())
		responseLogFields := logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.Path,
			"status_code": statusCode,
		}
		context.Logger.WithFields(responseLogFields).Debug("Received HTTP request")
	}()

	h.setDefaultHeaders(w, r)

	session, err := h.parseTokenAndGetSession(w, r)
	if err != nil {
		context.Logger.WithError(err).Warn("invalid session")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if h.requiresSession && !h.isValidSession(session) {
		utils.DeleteSessionCookies(w, r)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if h.requiresVerification && !h.isUserVerified(session) {
		context.Logger.WithField("user_id", session.UserID).Warn("user is not verified")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if h.requiresAdmin && !h.isUserAdmin(session) {
		context.Logger.WithField("user_id", session.UserID).Warn("user is is not admin")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if session == nil {
		session = &model.Session{}
	}

	context.Session = session

	h.handler(context, w, r)
}

func (h contextHandler) isValidSession(session *model.Session) bool {
	if session == nil {
		return false
	}

	if session.APIKeySession && (!h.allowApiKeySession && !h.allowJWTTokenSession) {
		return false
	}

	return true
}

func (h contextHandler) isUserVerified(session *model.Session) bool {
	if session == nil || session.UserID == "" {
		h.context.Logger.
			Error("session or userid in the session is not present")
		return false
	}
	user, err := h.context.App.User().Get(session.UserID)
	if err != nil {
		h.context.Logger.
			WithError(err).
			WithField("user_id", session.UserID).
			Error("error trying to get user")
		return false
	}
	return user.EmailVerified
}

func (h contextHandler) isUserAdmin(session *model.Session) bool {
	if session == nil || session.UserID == "" {
		h.context.Logger.
			Error("session or userid in the session is not present")
		return false
	}
	isAdmin, err := h.context.App.User().HasAdminPermission(session.UserID)
	if err != nil {
		h.context.Logger.
			WithError(err).
			WithField("user_id", session.UserID).
			Error("error trying to check if user has admin role")
		return false
	}
	return isAdmin
}

func (h contextHandler) parseTokenAndGetSession(w http.ResponseWriter, r *http.Request) (*model.Session, error) {
	token := ""
	tokenFromCookie := false

	authHeader := r.Header.Get(model.HeaderAuthorization)

	// Attempt to parse the token from the cookie
	if cookie, err := r.Cookie(model.SessionCookieToken); err == nil {
		tokenFromCookie = true
		token = cookie.Value
	}

	// Parse the token from the header
	if len(authHeader) > 6 && strings.ToUpper(authHeader[0:6]) == model.AuthorizationBearer {
		token = authHeader[7:]
	}

	if len(token) == 0 {
		return nil, nil
	}

	session, err := h.context.App.User().GetSession(token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	err = checkCSRFToken(r, token, tokenFromCookie, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (h contextHandler) setDefaultHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(model.HeaderRequestID, h.context.RequestID)

	if h.isStatic {
		// Instruct the browser not to display us in an iframe unless is the same origin for anti-clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// Set content security policy. This is also specified in the root.html of the webapp in a meta tag.
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		w.Header().Set("Expires", "0")
	}
}

// checkCSRFToken performs a CSRF check on the provided request with the given CSRF token.
func checkCSRFToken(r *http.Request, token string, tokenFromCookie bool, session *model.Session) error {
	csrfCheckNeeded := session != nil && tokenFromCookie && r.Method != "GET"
	if !csrfCheckNeeded {
		return nil
	}

	csrfHeader := r.Header.Get(model.HeaderCSRFToken)
	if csrfHeader != session.CSRFToken {
		return errors.New("possible CSRF attempt")
	}

	return nil
}

func newStaticHandler(context *Context, handler contextHandlerFunc) *contextHandler {
	return &contextHandler{
		context:  context,
		handler:  handler,
		isStatic: true,
	}
}

func newAPIHandler(context *Context, handler contextHandlerFunc) *contextHandler {
	return &contextHandler{
		context: context,
		handler: handler,
	}
}

func newAPISessionRequiredHandler(context *Context, handler contextHandlerFunc, requireVerification bool) *contextHandler {
	return &contextHandler{
		context:              context,
		handler:              handler,
		requiresSession:      true,
		requiresVerification: requireVerification,
	}
}

func newAPISessionAdminRequiredHandler(context *Context, handler contextHandlerFunc) *contextHandler {
	return &contextHandler{
		context:              context,
		handler:              handler,
		requiresSession:      true,
		requiresAdmin:        true,
		requiresVerification: true,
	}
}
