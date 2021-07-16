package store

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	// enable the pq driver
	_ "github.com/lib/pq"
)

type SqlStoreStores struct {
	role    RoleStore
	session SessionStore
	token   TokenStore
	user    UserStore
}

type SqlStore struct {
	db          *sqlx.DB
	tablePrefix string
	logger      logrus.FieldLogger
	stores      SqlStoreStores
}

// New constructs a new instance of SqlStore.
func New(dsn string, tablePrefix string, logger logrus.FieldLogger) (*SqlStore, error) {
	url, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dsn as url")
	}
	url.Host = strings.Replace(url.Host, "fileColonPlaceholder", "file:", 1)

	var db *sqlx.DB

	switch strings.ToLower(url.Scheme) {
	case "postgres", "postgresql":
		url.Scheme = "postgres"

		usePgTemp := false
		query := url.Query()
		if _, ok := query["pg_temp"]; ok {
			usePgTemp = true
			query.Del("pg_temp")
			url.RawQuery = query.Encode()
		}

		db, err = sqlx.Connect("postgres", url.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to connect to postgres database")
		}

		if usePgTemp {
			// Force the use of the current session's temporary-table schema,
			// simplifying cleanup for unit tests configured to use same.
			db.Exec("SET search_path TO pg_temp")
		}

		// Leave the default mapper as strings.ToLower.

	default:
		return nil, errors.Errorf("unsupported dsn scheme %s", url.Scheme)
	}

	var stores SqlStoreStores
	store := &SqlStore{
		db,
		tablePrefix,
		logger,
		stores,
	}
	store.stores.role = newSqlRoleStore(store)
	store.stores.session = newSqlSessionStore(store)
	store.stores.token = newSqlTokenStore(store)
	store.stores.user = newSqlUserStore(store)

	err = store.Migrate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tables")
	}

	err = store.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize store")
	}

	return store, nil
}

// queryer is an interface describing a resource that can query.
//
// It exactly matches sqlx.Queryer, existing simply to constrain sqlx usage to this file.
type queryer interface {
	sqlx.Queryer
}

// get queries for a single row, writing the result into dest.
//
// Use this to simplify querying for a single row or column. Dest may be a pointer to a simple
// type, or a struct with fields to be populated from the returned columns.
func (s *SqlStore) get(q sqlx.Queryer, dest interface{}, query string, args ...interface{}) error {
	query = s.db.Rebind(query)

	return sqlx.Get(q, dest, query, args...)
}

// builder is an interface describing a resource that can construct SQL and arguments.
//
// It exists to allow consuming any squirrel.*Builder type.
type builder interface {
	ToSql() (string, []interface{}, error)
}

// getBuilder queries for a single row, building the sql, and writing the result into dest.
//
// Use this to simplify querying for a single row or column. Dest may be a pointer to a simple
// type, or a struct with fields to be populated from the returned columns.
func (s *SqlStore) getBuilder(q sqlx.Queryer, dest interface{}, b builder) error {
	sql, args, err := b.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build sql")
	}

	sql = s.db.Rebind(sql)

	err = sqlx.Get(q, dest, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// selectBuilder queries for one or more rows, building the sql, and writing the result into dest.
//
// Use this to simplify querying for multiple rows (and possibly columns). Dest may be a slice of
// a simple, or a slice of a struct with fields to be populated from the returned columns.
func (s *SqlStore) selectBuilder(q sqlx.Queryer, dest interface{}, b builder) error {
	sql, args, err := b.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build sql")
	}

	sql = s.db.Rebind(sql)

	err = sqlx.Select(q, dest, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// execer is an interface describing a resource that can execute write queries.
//
// It allows the use of *sqlx.Db and *sqlx.Tx.
type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	DriverName() string
}

// exec executes the given query using positional arguments, automatically rebinding for the db.
func (s *SqlStore) exec(e execer, sql string, args ...interface{}) (sql.Result, error) {
	sql = s.db.Rebind(sql)
	return e.Exec(sql, args...)
}

// exec executes the given query, building the necessary sql.
func (s *SqlStore) execBuilder(e execer, b builder) (sql.Result, error) {
	sql, args, err := b.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql")
	}

	return s.exec(e, sql, args...)
}
