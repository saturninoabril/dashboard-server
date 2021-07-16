package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/saturninoabril/dashboard-server/app"
	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/store"
	"github.com/saturninoabril/dashboard-server/testlib"
	"github.com/stretchr/testify/require"
)

const (
	testPassword = "Password12"
)

type TestHelper struct {
	App      *app.App
	Router   *mux.Router
	Server   *httptest.Server
	SqlStore *store.SqlStore
}

func SetupTestHelper(t *testing.T) *TestHelper {
	logger := testlib.MakeLogger(t)
	config := app.NewConfig()
	app.SetDevConfig(&config)
	logger.Debug("Using dev configuration")

	sqlStore := store.MakeTestStore(t, logger)

	userService := app.NewUserService(logger, sqlStore)
	appService := app.NewApp(logger, sqlStore, config, userService)

	err := appService.ReloadHTMLTemplates()
	if err != nil {
		logger.WithError(err).Warn("Unable to load HTML templates")
	}

	router := mux.NewRouter()

	Register(router, &Context{
		App:    appService,
		Logger: logger,
	})

	ts := httptest.NewServer(router)

	return &TestHelper{
		App:      appService,
		Router:   router,
		Server:   ts,
		SqlStore: sqlStore,
	}
}

func (th *TestHelper) TearDown(t *testing.T) {
	th.Server.Close()
	store.CloseConnection(t, th.SqlStore)
}

func signUpWithEmail(t *testing.T, email string, client *model.Client, sqlStore *store.SqlStore) *model.User {
	resp, err := client.SignUp(&model.SignUpRequest{Email: email, Password: testPassword})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.User)

	err = sqlStore.User().VerifyEmail(resp.User.ID, resp.User.Email)
	require.NoError(t, err)

	return resp.User
}

func signUp(t *testing.T, client *model.Client, sqlStore *store.SqlStore) *model.User {
	return signUpWithEmail(t, model.NewID()+"@example.com", client, sqlStore)
}
