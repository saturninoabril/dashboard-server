package store

import (
	"github.com/blang/semver"
	"github.com/pkg/errors"
)

const systemDatabaseVersionKey = "DatabaseVersion"

// GetCurrentVersion queries the System table for the current database version.
func (store *SqlStore) GetCurrentVersion() (semver.Version, error) {
	return store.getCurrentVersion(store.db)
}

// getCurrentVersion queries the System table for the current database version against the given
// queryer.
func (store *SqlStore) getCurrentVersion(q queryer) (semver.Version, error) {
	currentVersionStr, err := store.getSystemValue(q, systemDatabaseVersionKey)
	if currentVersionStr == "" {
		return semver.Version{}, nil
	}

	currentVersion, err := semver.Parse(currentVersionStr)
	if err != nil {
		return semver.Version{}, errors.Wrapf(err, "failed to parse current version %s", currentVersionStr)
	}

	return currentVersion, nil
}

// setCurrentVersion updates the System table with the given database version.
func (store *SqlStore) setCurrentVersion(e execer, version string) error {
	return store.setSystemValue(e, systemDatabaseVersionKey, version)
}
