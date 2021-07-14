package api

import (
	"testing"

	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/store"
	"github.com/stretchr/testify/require"
)

const (
	testPassword = "Password12"
)

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
