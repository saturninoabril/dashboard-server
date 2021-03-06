package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

func (s *SqlStore) getSystemTable() string {
	return s.tablePrefix + "system"
}

// getSystemValue queries the System table for the given key
func (s *SqlStore) getSystemValue(q queryer, key string) (string, error) {
	var value string

	err := s.getBuilder(
		q,
		&value,
		sq.Select("value").From(s.getSystemTable()).Where("key = ?", key),
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
	result, err := s.execBuilder(
		e,
		sq.Update(s.getSystemTable()).
			Where("key = ?", key).
			Set("value", value),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update system key %s", key)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return nil
	}

	_, err = s.execBuilder(
		e,
		sq.Insert(s.getSystemTable()).
			Columns("key", "value").
			Values(key, value),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert system key %s", key)
	}

	return nil
}
