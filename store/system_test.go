package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemValue(t *testing.T) {
	th := SetupStoreTestHelper(t)
	defer th.TearDown(t)

	t.Run("unknown value", func(t *testing.T) {
		value, err := th.SqlStore.getSystemValue(th.SqlStore.db, "unknown")
		require.NoError(t, err)
		require.Empty(t, value)
	})

	t.Run("known value", func(t *testing.T) {

		key1 := "key1"
		value1 := "value1"
		key2 := "key2"
		value2 := "value2"

		err := th.SqlStore.setSystemValue(th.SqlStore.db, key1, value1)
		require.NoError(t, err)

		err = th.SqlStore.setSystemValue(th.SqlStore.db, key2, value2)
		require.NoError(t, err)

		actualValue1, err := th.SqlStore.getSystemValue(th.SqlStore.db, key1)
		require.NoError(t, err)
		require.Equal(t, value1, actualValue1)

		actualValue2, err := th.SqlStore.getSystemValue(th.SqlStore.db, key2)
		require.NoError(t, err)
		require.Equal(t, value2, actualValue2)
	})
}
