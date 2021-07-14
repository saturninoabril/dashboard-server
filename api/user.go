package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/saturninoabril/dashboard-server/model"
)

// initUser registers user endpoints on the given router.
func initUser(apiRouter *mux.Router, context *Context) {
	usersRouter := apiRouter.PathPrefix("/users").Subrouter()
	usersRouter.Handle("/signup", newAPIHandler(context, handleSignUp)).Methods("POST")
	usersRouter.Handle("/login", newAPIHandler(context, handleLogin)).Methods("POST")
	usersRouter.Handle("/logout", newAPIHandler(context, handleLogout)).Methods("POST")
	usersRouter.Handle("/forgot-password", newAPIHandler(context, handleForgotPassword)).Methods("POST")
	usersRouter.Handle("/reset-password-complete", newAPIHandler(context, handleResetPassword)).Methods("POST")
	usersRouter.Handle("/verify-email", newAPISessionRequiredHandler(context, handleVerifyEmailStart, false)).Methods("POST")
	usersRouter.Handle("/verify-email-complete", newAPIHandler(context, handleVerifyEmailComplete)).Methods("POST")
	usersRouter.Handle("/me", newAPISessionRequiredHandler(context, handleGetMe, false)).Methods("GET")
	usersRouter.Handle("/me", newAPISessionRequiredHandler(context, handleUpdateMe, true)).Methods("PUT")
	usersRouter.Handle("/me/password", newAPISessionRequiredHandler(context, handleUpdatePassword, true)).Methods("PUT")
}

// handleSignUp responds to POST /api/v1/users/signup, creating a user.
func handleSignUp(c *Context, w http.ResponseWriter, r *http.Request) {
	sr := &model.SignUpRequest{}
	err := decodeJSON(sr, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	user := &model.User{
		FirstName: sr.FirstName,
		LastName:  sr.LastName,
		Email:     sr.Email,
		Password:  sr.Password,
	}

	err = user.IsValidPassword()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	user, err = c.App.User().Create(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	err = c.App.User().Login(w, r, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	user.Sanitize()
	resp := &model.SignUpResponse{
		User: user,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

// handleLogin responds to POST /api/v1/users/login, logging the user in.
func handleLogin(c *Context, w http.ResponseWriter, r *http.Request) {
	lr := &model.LoginRequest{}
	err := decodeJSON(lr, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	user, err := c.App.User().AuthenticateUserForLogin(lr.Email, lr.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c.writeAndLogErrorWithFields(w, err, logrus.Fields{"email": lr.Email})
		return
	}

	if user.State != model.UserStateActive {
		w.WriteHeader(http.StatusLocked)
		c.writeAndLogErrorWithFields(w, errors.New("Attempt to login to locked account"), logrus.Fields{"email": user.Email})
		return
	}

	err = c.App.User().Login(w, r, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	user.Sanitize()
	b, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.Write(b)
}

// handleLogout responds to POST /api/v1/users/logout, logging the user out.
func handleLogout(c *Context, w http.ResponseWriter, r *http.Request) {
	c.App.User().Logout(w, r, c.Session.ID)
}

// handleVerifyEmailStart responds to POST /api/v1/users/verify-email, sending
// an email to the user to verify their email address.
func handleVerifyEmailStart(c *Context, w http.ResponseWriter, r *http.Request) {
	user, err := c.App.User().Get(c.Session.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		c.writeAndLogErrorWithFields(w, err, logrus.Fields{"session_id": c.Session.ID})
		return
	}

	if user.EmailVerified {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("user email is already verified"))
		return
	}

	token, err := c.App.CreateAndStoreVerifyEmailToken(user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	err = c.App.SendVerifyEmailEmail(user.Email, c.App.Config().SiteURL, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.Write([]byte(`{"status": "ok"}`))
}

// handleVerifyEmailComplete responds to POST /api/v1/users/verify-email-complete,
// confirming that the user's email is valid.
func handleVerifyEmailComplete(c *Context, w http.ResponseWriter, r *http.Request) {
	ver := &model.VerifyEmailRequest{}
	err := decodeJSON(ver, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	token, err := c.App.Store().GetToken(ver.Token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}
	if token == nil ||
		token.Type != model.TokenTypeVerifyEmail ||
		token.CreateAt < model.GetMillis()-model.TokenDefaultExpiryTime {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("invalid token"))
		return
	}

	email, err := token.GetExtraEmail()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to determine token email value"))
		return
	}

	user, err := c.App.User().GetByEmail(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to get user by email"))
		return
	}

	err = c.App.User().VerifyEmail(user.ID, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to set user email as verified"))
		return
	}

	err = c.App.Store().DeleteToken(token.Token)
	if err != nil {
		// The email was verified, but the token used for this could not be
		// cleaned up. The expired token cleanup routine will get it later, but
		// the token is still technically valid to be used again.
		c.Logger.WithError(err).Errorf("Failed to remove claimed %s token", model.TokenTypeVerifyEmail)
	}

	w.Write([]byte(`{"status": "ok"}`))
}

// handleForgotPassword responds to POST /api/v1/users/forgot-password, starting
// the password recovery flow for a given email address.
func handleForgotPassword(c *Context, w http.ResponseWriter, r *http.Request) {
	fpr := &model.ForgotPasswordRequest{}
	err := decodeJSON(fpr, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	user, err := c.App.User().GetByEmail(fpr.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	token, err := c.App.CreateAndStoreResetPasswordToken(fpr.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	err = c.App.SendPasswordResetEmail(fpr.Email, c.App.Config().SiteURL, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.Write([]byte(`{"status": "ok"}`))
}

// handleResetPassword responds to POST /api/v1/users/reset-password-complete,
// resetting a user password with a valid token.
func handleResetPassword(c *Context, w http.ResponseWriter, r *http.Request) {
	rpr := &model.ResetPasswordRequest{}
	err := decodeJSON(rpr, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	if !model.IsValidPassword(rpr.Password) {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("invalid password"))
		return
	}

	token, err := c.App.Store().GetToken(rpr.Token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}
	if token == nil ||
		token.Type != model.TokenTypeResetPassword ||
		token.CreateAt < model.GetMillis()-model.TokenDefaultExpiryTime {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("invalid token"))
		return
	}

	email, err := token.GetExtraEmail()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to determine token email value"))
		return
	}

	user, err := c.App.User().GetByEmail(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to get user by email"))
		return
	}

	err = c.App.Store().User().UpdatePassword(user.ID, rpr.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to update user password"))
		return
	}

	// Invalidate all the user sessions
	err = c.App.Store().DeleteSessionsForUser(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to invalidate user sessions"))
		return
	}

	err = c.App.Store().DeleteToken(token.Token)
	if err != nil {
		// The password was reset, but the token used for this could not be
		// cleaned up. The expired token cleanup routine will get it later, but
		// the token is still technically valid to be used again.
		c.Logger.WithError(err).Errorf("Failed to remove claimed %s token", model.TokenTypeResetPassword)
	}

	w.Write([]byte(`{"status": "ok"}`))
}

// handleUpdatePassword responds to PUT /api/v1/users/me/password,
// updating a user's password.
func handleUpdatePassword(c *Context, w http.ResponseWriter, r *http.Request) {
	upr := &model.UpdatePasswordRequest{}
	err := decodeJSON(upr, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	if upr.CurrentPassword == "" || upr.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("current and new password not set"))
		return
	}

	if !model.IsValidPassword(upr.NewPassword) {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("invalid password"))
		return
	}

	user, err := c.App.User().Get(c.Session.UserID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c.writeAndLogError(w, err)
		return
	}

	user, err = c.App.User().AuthenticateUserForLogin(user.Email, upr.CurrentPassword)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		c.writeAndLogError(w, errors.New("bad old password"))
		return
	}

	err = c.App.Store().User().UpdatePassword(user.ID, upr.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to update user password"))
		return
	}

	// Invalidate all the user sessions
	err = c.App.Store().DeleteSessionsForUser(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.Wrap(err, "failed to invalidate user sessions"))
		return
	}

	// Re-login the user with the new session
	err = c.App.User().Login(w, r, user)
	if err != nil {
		c.Logger.WithError(err).WithField("userid", user.ID).Warn("error trying to re-login the user after the password change")
	}

	w.Write([]byte(`{"status": "ok"}`))
}

// handleGetMe responds to GET /api/v1/users/me, getting the logged in user.
func handleGetMe(c *Context, w http.ResponseWriter, r *http.Request) {
	user, err := c.App.User().Get(c.Session.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		c.writeAndLogErrorWithFields(w, errors.New("not found"), logrus.Fields{"session_id": c.Session.ID})
		return
	}

	user.Sanitize()
	b, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.Write(b)
}

// handleUpdateMe responds to PUT /api/v1/users/me, updating the logged in user.
func handleUpdateMe(c *Context, w http.ResponseWriter, r *http.Request) {
	reqUser, err := model.UserFromReader(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	if reqUser.ID != c.Session.UserID {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("user id in body did not match session"))
		return
	}

	err = reqUser.IsValid()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	user, err := c.App.User().Get(reqUser.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	// If email change is requested we change the user to an unverified
	// stat and we send the verification email again
	if user.Email != reqUser.Email {
		err = c.App.User().UnverifyEmail(user.ID, user.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			c.writeAndLogError(w, err)
			return
		}
		token, err := c.App.CreateAndStoreVerifyEmailToken(reqUser.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			c.writeAndLogError(w, err)
			return
		}

		err = c.App.SendVerifyEmailEmail(reqUser.Email, c.App.Config().SiteURL, token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			c.writeAndLogError(w, err)
			return
		}
	}

	user, err = c.App.User().Update(reqUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	user.Sanitize()
	b, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.Write(b)
}
