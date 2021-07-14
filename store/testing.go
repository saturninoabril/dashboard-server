package store

import (
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func makeUnmigratedTestStore(tb testing.TB, logger log.FieldLogger) *SqlStore {
	dsn := os.Getenv("DASHBOARD_DATABASE_TEST")
	fmt.Printf("dsn: %v\n", dsn)
	store, err := New(dsn, logger)
	fmt.Printf("err: %v\n", err)
	require.NoError(tb, err)

	return store
}

func MakeTestStore(tb testing.TB, logger log.FieldLogger) *SqlStore {
	sqlStore := makeUnmigratedTestStore(tb, logger)
	err := sqlStore.Migrate()
	require.NoError(tb, err)

	return sqlStore
}

func CloseConnection(tb testing.TB, store *SqlStore) {
	err := store.db.Close()
	require.NoError(tb, err)
}
