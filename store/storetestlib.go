package store

import (
	"os"
	"testing"

	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/testlib"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testPassword = "Password12"
)

type StoreTestHelper struct {
	SqlStore *SqlStore
}

func SetupStoreTestHelper(t *testing.T) *StoreTestHelper {
	logger := testlib.MakeLogger(t)
	sqlStore := MakeTestStore(t, logger)

	return &StoreTestHelper{
		SqlStore: sqlStore,
	}
}

func (th *StoreTestHelper) TearDown(t *testing.T) {
	CloseConnection(t, th.SqlStore)
}

func MakeTestStore(tb testing.TB, logger log.FieldLogger) *SqlStore {
	dsn := os.Getenv("DASHBOARD_DATABASE_TEST")
	tablePrefix := os.Getenv("DASHBOARD_TABLE_PREFIX")
	sqlStore, err := New(dsn, tablePrefix, logger)
	require.NoError(tb, err)

	return sqlStore
}

func CloseConnection(tb testing.TB, store *SqlStore) {
	err := store.db.Close()
	require.NoError(tb, err)
}

func createTestUser(t *testing.T, store *SqlStore) *model.User {
	user := &model.User{
		Email:    testlib.GetTestEmail(),
		Password: testPassword,
	}
	user, err := store.User().CreateUser(user)
	require.NoError(t, err)
	require.NotNil(t, user)

	return user
}
