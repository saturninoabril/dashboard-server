package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

// getSystemValue queries the System table for the given key
func (s *SqlStore) getSystemValue(q queryer, key string) (string, error) {
	var value string

	err := s.getBuilder(q, &value,
		sq.Select("Value").From(s.tablePrefix+"System").Where("Key = ?", key),
	)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", errors.Wrapf(err, "failed to query system key %s", key)
	}

	return value, nil
}

// setSystemValue updates the System table for the given key.
func (s *SqlStore) setSystemValue(e execer, key, value string) error {
	result, err := s.execBuilder(e,
		sq.Update(s.tablePrefix+"System").Set("Value", value).Where("Key = ?", key),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update system key %s", key)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return nil
	}

	result, err = s.execBuilder(e,
		sq.Insert(s.tablePrefix+"System").Columns("Key", "Value").Values(key, value),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert system key %s", key)
	}

	return nil
}
