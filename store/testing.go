package store

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

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
