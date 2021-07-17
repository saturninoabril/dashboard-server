package store

import (
	"math/rand"
	"testing"
	"time"

	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/testlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	th := SetupStoreTestHelper(t)
	defer th.TearDown(t)

	t.Run("get unknown token", func(t *testing.T) {
		token, err := th.SqlStore.Token().GetToken("")
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("create invalid token", func(t *testing.T) {
		var err error
		token := model.NewToken("invalid_type", "extra")
		require.Error(t, token.IsValid())

		token, err = th.SqlStore.Token().CreateToken(token)
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("create and get token", func(t *testing.T) {
		var err error
		originalToken := model.NewToken(model.TokenTypeVerifyEmail, "extra")
		require.NoError(t, originalToken.IsValid())

		token, err := th.SqlStore.Token().CreateToken(originalToken)
		require.NoError(t, err)
		require.Equal(t, originalToken, token)

		token, err = th.SqlStore.Token().GetToken(token.Token)
		require.NoError(t, err)
		require.Equal(t, originalToken, token)
	})

	t.Run("delete token", func(t *testing.T) {
		var err error
		originalToken := model.NewToken(model.TokenTypeVerifyEmail, "extra")
		require.NoError(t, originalToken.IsValid())

		token, err := th.SqlStore.Token().CreateToken(originalToken)
		require.NoError(t, err)
		require.Equal(t, originalToken, token)

		token, err = th.SqlStore.Token().GetToken(token.Token)
		require.NoError(t, err)
		require.NotNil(t, token)

		err = th.SqlStore.Token().DeleteToken(token.Token)
		require.NoError(t, err)

		token, err = th.SqlStore.Token().GetToken(token.Token)
		require.NoError(t, err)
		require.Nil(t, token)
	})

	t.Run("delete tokens by email and type", func(t *testing.T) {
		email := testlib.GetTestEmail()

		extra, err := model.CreateTokenTypeResetPasswordExtra(email)
		require.NoError(t, err)
		originalTokenOne := model.NewToken(model.TokenTypeVerifyEmail, extra)
		require.NoError(t, originalTokenOne.IsValid())
		originalTokenTwo := model.NewToken(model.TokenTypeResetPassword, extra)
		require.NoError(t, originalTokenTwo.IsValid())

		tokenOne, err := th.SqlStore.Token().CreateToken(originalTokenOne)
		require.NoError(t, err)
		require.Equal(t, originalTokenOne, tokenOne)
		tokenTwo, err := th.SqlStore.Token().CreateToken(originalTokenTwo)
		require.NoError(t, err)
		require.Equal(t, originalTokenTwo, tokenTwo)

		tokenOne, err = th.SqlStore.Token().GetToken(tokenOne.Token)
		require.NoError(t, err)
		require.NotNil(t, tokenOne)
		tokenTwo, err = th.SqlStore.Token().GetToken(tokenTwo.Token)
		require.NoError(t, err)
		require.NotNil(t, tokenTwo)

		err = th.SqlStore.Token().DeleteTokensByEmail(email, model.TokenTypeVerifyEmail)
		require.NoError(t, err)

		tokenOne, err = th.SqlStore.Token().GetToken(tokenOne.Token)
		require.NoError(t, err)
		require.Nil(t, tokenOne)
		tokenTwo, err = th.SqlStore.Token().GetToken(tokenTwo.Token)
		require.NoError(t, err)
		require.NotNil(t, tokenTwo)
	})
}
