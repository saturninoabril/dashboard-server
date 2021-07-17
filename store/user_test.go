package store

import (
	"testing"

	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/testlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	th := SetupStoreTestHelper(t)
	defer th.TearDown(t)

	t.Run("get unknown user", func(t *testing.T) {
		user, err := th.SqlStore.User().GetUser("")
		assert.NoError(t, err)
		assert.Nil(t, user)

		user, err = th.SqlStore.User().GetUserByEmail("")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("create and get user", func(t *testing.T) {
		var err error
		user := &model.User{}

		user, err = th.SqlStore.User().CreateUser(user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email")

		user = &model.User{}
		email := testlib.GetTestEmail()
		user.Email = email
		user.Password = testPassword

		user, err = th.SqlStore.User().CreateUser(user)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, email, user.Email)

		_, err = th.SqlStore.User().CreateUser(user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint \""+th.SqlStore.tablePrefix+"users_pkey\"")

		id := user.ID

		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, email, user.Email)

		user, err = th.SqlStore.User().GetUserByEmail(user.Email)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, id, user.ID)
	})

	t.Run("verify email", func(t *testing.T) {
		var err error
		user := &model.User{
			Email:    testlib.GetTestEmail(),
			Password: testPassword,
		}

		user, err = th.SqlStore.User().CreateUser(user)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.False(t, user.EmailVerified)
		require.Equal(t, user.CreateAt, user.UpdateAt)

		err = th.SqlStore.User().VerifyEmail(user.ID, user.Email)
		require.NoError(t, err)

		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.True(t, user.EmailVerified)
		require.NotEqual(t, user.CreateAt, user.UpdateAt)
	})

	t.Run("unverify email", func(t *testing.T) {
		var err error
		user := &model.User{
			Email:    testlib.GetTestEmail(),
			Password: testPassword,
		}

		user, err = th.SqlStore.User().CreateUser(user)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.False(t, user.EmailVerified)
		require.Equal(t, user.CreateAt, user.UpdateAt)

		err = th.SqlStore.User().VerifyEmail(user.ID, user.Email)
		require.NoError(t, err)

		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.True(t, user.EmailVerified)
		require.NotEqual(t, user.CreateAt, user.UpdateAt)

		err = th.SqlStore.User().UnverifyEmail(user.ID, user.Email)
		require.NoError(t, err)

		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.False(t, user.EmailVerified)
		require.NotEqual(t, user.CreateAt, user.UpdateAt)
	})

	t.Run("update password", func(t *testing.T) {
		var err error
		password := "Password1"
		user := &model.User{
			Email:    testlib.GetTestEmail(),
			Password: password,
		}

		user, err = th.SqlStore.User().CreateUser(user)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, user.CreateAt, user.UpdateAt)
		require.True(t, model.ComparePassword(user.Password, password))

		newPassword := "Password2"
		err = th.SqlStore.User().UpdatePassword(user.ID, newPassword)
		require.NoError(t, err)

		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.NotEqual(t, user.CreateAt, user.UpdateAt)
		require.True(t, model.ComparePassword(user.Password, newPassword))
	})

	t.Run("update user", func(t *testing.T) {
		var err error
		user := &model.User{}
		email := testlib.GetTestEmail()
		user.Email = email
		user.Password = testPassword

		user, err = th.SqlStore.User().CreateUser(user)
		require.NoError(t, err)
		require.NotNil(t, user)

		newEmail := testlib.GetTestEmail()
		newFirstname := "NewFirstname"
		newLastname := "NewLastname"

		user.Email = newEmail
		user.FirstName = newFirstname
		user.LastName = newLastname

		err = th.SqlStore.User().UpdateUser(user)
		require.NoError(t, err)

		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, newEmail, user.Email)
		assert.Equal(t, newFirstname, user.FirstName)
		assert.Equal(t, newLastname, user.LastName)

		var newCreateAt int64 = 7
		user.CreateAt = newCreateAt
		err = th.SqlStore.User().UpdateUser(user)
		require.NoError(t, err)
		user, err = th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.NotEqual(t, newCreateAt, user.CreateAt)

		user.Email = ""
		err = th.SqlStore.User().UpdateUser(user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email")

		user2, err := th.SqlStore.User().CreateUser(&model.User{Email: testlib.GetTestEmail(), Password: testPassword})
		require.NoError(t, err)
		require.NotNil(t, user2)
		user.Email = user2.Email
		err = th.SqlStore.User().UpdateUser(user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "account exists")

		user.ID = "junk"
		user.Email = testlib.GetTestEmail()
		err = th.SqlStore.User().UpdateUser(user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid id")

		user.ID = model.NewID()
		err = th.SqlStore.User().UpdateUser(user)
		require.NoError(t, err)
	})

	t.Run("lock user", func(t *testing.T) {
		var err error
		user := &model.User{}
		email := testlib.GetTestEmail()
		user.Email = email
		user.Password = testPassword

		user, err = th.SqlStore.User().CreateUser(user)
		require.NoError(t, err)
		require.NotNil(t, user)

		err = th.SqlStore.User().UpdateUserState(user.ID, model.UserStateLocked)
		require.NoError(t, err)
		require.Nil(t, err)

		user2, err := th.SqlStore.User().GetUser(user.ID)
		require.NoError(t, err)
		require.NotNil(t, user2)
		assert.Equal(t, user2.State, model.UserStateLocked)
	})
}
