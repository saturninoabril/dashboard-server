package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/saturninoabril/dashboard-server/app"
	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/store"
	"github.com/saturninoabril/dashboard-server/testlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	logger := testlib.MakeLogger(t)
	config := app.NewConfig()

	sqlStore := store.MakeTestStore(t, logger)
	defer store.CloseConnection(t, sqlStore)

	userService := app.NewUserService(logger, sqlStore)
	appService := app.NewApp(logger, sqlStore, config, userService)

	router := mux.NewRouter()

	Register(router, &Context{
		App:    appService,
		Logger: logger,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("log in and out", func(t *testing.T) {
		client := model.NewClient(ts.URL)

		user := signUp(t, client, sqlStore)
		email := user.Email

		user, err := client.Login(&model.LoginRequest{Email: email, Password: testPassword})
		assert.NoError(t, err)
		require.NotNil(t, user)

		headers := client.Headers()
		assert.Equal(t, email, user.Email)
		assert.Equal(t, 33, len(headers[model.HeaderAuthorization])) // 33 is length of "bearer <token>""
		assert.Equal(t, "", user.Password)

		err = client.Logout()
		assert.NoError(t, err)
		assert.Empty(t, client.Headers()[model.HeaderAuthorization])

		// Make sure we can't use the token after logout
		client = model.NewClientWithHeaders(ts.URL, headers)
		user, err = client.GetMe()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "401")
		assert.Nil(t, user)
	})

	t.Run("update password", func(t *testing.T) {
		client := model.NewClient(ts.URL)

		err := client.UpdatePassword(&model.UpdatePasswordRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "401")

		user := signUp(t, client, sqlStore)

		err = client.UpdatePassword(&model.UpdatePasswordRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")

		err = client.UpdatePassword(&model.UpdatePasswordRequest{CurrentPassword: "junk", NewPassword: ""})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")

		err = client.UpdatePassword(&model.UpdatePasswordRequest{CurrentPassword: "", NewPassword: "junk"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")

		err = client.UpdatePassword(&model.UpdatePasswordRequest{CurrentPassword: testPassword, NewPassword: "notvalid"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")

		err = client.UpdatePassword(&model.UpdatePasswordRequest{CurrentPassword: testPassword, NewPassword: "test123456"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")

		// Check if session is present for that user
		sessionHeader := strings.Split(client.Headers()[model.HeaderAuthorization], " ")
		session, err := sqlStore.GetSession(sessionHeader[1])
		require.NoError(t, err)
		require.NotNil(t, session)

		err = client.UpdatePassword(&model.UpdatePasswordRequest{CurrentPassword: testPassword, NewPassword: "Test123456"})
		require.NoError(t, err)

		// CHeck that session is in place again after the login
		sessionHeader = strings.Split(client.Headers()[model.HeaderAuthorization], " ")
		session, err = sqlStore.GetSession(sessionHeader[1])
		require.NoError(t, err)
		require.NotNil(t, session)

		_, err = client.Login(&model.LoginRequest{Email: user.Email, Password: testPassword})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "401")

		_, err = client.Login(&model.LoginRequest{Email: user.Email, Password: "Test123456"})
		require.NoError(t, err)
	})

	t.Run("get me", func(t *testing.T) {
		client := model.NewClient(ts.URL)

		user := signUp(t, client, sqlStore)
		password := testPassword
		email := user.Email

		client.Logout()

		_, err := client.GetMe()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "401")

		user, err = client.Login(&model.LoginRequest{Email: email, Password: password})
		require.NoError(t, err)

		newUser, err := client.GetMe()
		assert.NoError(t, err)
		require.NotNil(t, newUser)
		assert.Equal(t, email, newUser.Email)
		assert.Equal(t, user.ID, newUser.ID)
		assert.Equal(t, "", user.Password)
	})

	t.Run("sign up", func(t *testing.T) {
		client := model.NewClient(ts.URL)

		_, err := client.SignUp(&model.SignUpRequest{})
		assert.Error(t, err)

		email := "usertest" + model.NewID() + "@example.com"

		_, err = client.SignUp(&model.SignUpRequest{Email: email, Password: "password"})
		assert.Error(t, err)

		resp, err := client.SignUp(&model.SignUpRequest{Email: email, Password: testPassword})
		assert.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)
		assert.Equal(t, email, resp.User.Email)
		assert.Equal(t, "", resp.User.Password)
		assert.Equal(t, 33, len(client.Headers()[model.HeaderAuthorization])) // 33 is length of "bearer <token>""

		email = "USERTEST" + model.NewID() + "@example.com"
		resp, err = client.SignUp(&model.SignUpRequest{Email: email, Password: testPassword})
		assert.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)
		assert.Equal(t, strings.ToLower(email), resp.User.Email)
	})
}
