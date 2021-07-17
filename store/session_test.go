package store

import (
	"testing"

	"github.com/saturninoabril/dashboard-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessions(t *testing.T) {
	th := SetupStoreTestHelper(t)
	defer th.TearDown(t)

	t.Run("get unknown session", func(t *testing.T) {
		token, err := th.SqlStore.Session().GetSession("")
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("create, get and delete session", func(t *testing.T) {
		var err error
		session := &model.Session{}

		userID := model.NewID()
		session.UserID = userID

		session, err = th.SqlStore.Session().CreateSession(session)
		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, userID, session.UserID)

		id := session.ID
		token := session.Token

		session, err = th.SqlStore.Session().GetSession(session.ID)
		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, token, session.Token)

		session, err = th.SqlStore.Session().GetSession(session.Token)
		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, id, session.ID)

		err = th.SqlStore.Session().DeleteSession(session.ID)
		assert.NoError(t, err)

		session, err = th.SqlStore.Session().GetSession(session.ID)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})

	t.Run("delete non-existent session", func(t *testing.T) {
		err := th.SqlStore.Session().DeleteSession("junk")
		assert.NoError(t, err)
	})

	t.Run("delete all user sessions", func(t *testing.T) {
		var err error
		session1 := &model.Session{}
		session2 := &model.Session{}

		userID := model.NewID()
		session1.UserID = userID
		session2.UserID = userID

		session1, err = th.SqlStore.Session().CreateSession(session1)
		require.NoError(t, err)
		require.NotNil(t, session1)
		assert.Equal(t, userID, session1.UserID)

		session2, err = th.SqlStore.Session().CreateSession(session2)
		require.NoError(t, err)
		require.NotNil(t, session2)
		assert.Equal(t, userID, session2.UserID)

		err = th.SqlStore.Session().DeleteSessionsForUser(userID)
		require.NoError(t, err)

		session1, err = th.SqlStore.Session().GetSession(session1.ID)
		assert.NoError(t, err)
		assert.Nil(t, session1)

		session2, err = th.SqlStore.Session().GetSession(session2.ID)
		assert.NoError(t, err)
		assert.Nil(t, session2)
	})
}
