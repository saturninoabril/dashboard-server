package store

import (
	"strings"

	"github.com/lib/pq"
	"gopkg.in/pkg/errors.v0"
)

// tableExists determines if the given table name exists in the database.
func (store *DashboardStore) tableExists(tableName string) (bool, error) {
	var tableExists bool

	switch store.db.DriverName() {
	case "postgres":
		err := store.get(store.db, &tableExists,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'system')",
		)
		if err != nil {
			return false, errors.Wrapf(err, "failed to check if %s table exists", tableName)
		}

	default:
		return false, errors.Errorf("unsupported driver %s", store.db.DriverName())
	}

	return tableExists, nil
}

func isUniqueConstraintError(err error, indexName []string) bool {
	unique := false
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		unique = true
	}

	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		unique = true
	}

	field := false
	for _, contain := range indexName {
		if strings.Contains(err.Error(), contain) {
			field = true
			break
		}
	}

	return unique && field
}
