package store

import (
	"github.com/blang/semver"
	"github.com/saturninoabril/dashboard-server/model"
)

type migration struct {
	fromVersion   semver.Version
	toVersion     semver.Version
	migrationFunc func(execer) error
}

// migrations defines the set of migrations necessary to advance the database to the latest
// expected version.
//
// Note that the canonical schema is currently obtained by applying all migrations to an empty
// database.
var migrations = []migration{
	{semver.MustParse("0.0.0"), semver.MustParse("0.1.0"), func(e execer) error {
		_, err := e.Exec(`
			CREATE TABLE System (
				Key    VARCHAR(64) PRIMARY KEY,
				Value  VARCHAR(1024) NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE Users (
				ID         CHAR(26) PRIMARY KEY,
				CreateAt   BIGINT NOT NULL,
				UpdateAt   BIGINT NOT NULL,
				Email      VARCHAR(64) NOT NULL UNIQUE,
				EmailVerified  BOOLEAN NOT NULL,
				Password   VARCHAR(128) NOT NULL,
				FirstName  VARCHAR(64) NOT NULL,
				LastName   VARCHAR(64) NOT NULL,
				State TEXT DEFAULT 'active'
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE Session (
				ID         CHAR(26) PRIMARY KEY,
				Token      CHAR(26) NOT NULL,
				CreateAt   BIGINT NOT NULL,
				ExpiresAt  BIGINT NOT NULL,
				UserID     VARCHAR(64) NOT NULL,
				CSRFToken  CHAR(26) NOT NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE Tokens (
				Token     VARCHAR(64) PRIMARY KEY,
				CreateAt  BIGINT DEFAULT NULL,
				Type      VARCHAR(64) DEFAULT NULL,
				Extra     VARCHAR(256) DEFAULT NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE Roles (
				ID             CHAR(26) PRIMARY KEY,
				Name		   TEXT NOT NULL,
				CreateAt       BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
				UpdateAt       BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000)
		);`)
		if err != nil {
			return err
		}
		_, err = e.Exec(`
			CREATE TABLE UserRoles (
				UserID      CHAR(26) NOT NULL,
				RoleID		CHAR(64) NOT NULL,
				CreateAt    BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
				UpdateAt    BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
				PRIMARY     KEY (UserID, RoleID)
		);`)
		if err != nil {
			return err
		}
		userRoleID := model.NewID()
		if _, err = e.Exec(`INSERT INTO Roles VALUES ($1, $2);`, userRoleID, model.UserRoleName); err != nil {
			return err
		}
		if _, err = e.Exec(`INSERT INTO Roles VALUES ($1, $2);`, model.NewID(), model.AdminRoleName); err != nil {
			return err
		}

		return nil
	}},
}
