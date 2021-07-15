package store

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/store/migrations"
)

type PrefixedMigration struct {
	*bindata.Bindata
	prefix   string
	postgres bool
}

func init() {
	source.Register("prefixed-migrations", &PrefixedMigration{})
}

func (pm *PrefixedMigration) executeTemplate(r io.ReadCloser, identifier string) (io.ReadCloser, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	tmpl, err := template.New("sql").Parse(string(data))
	if err != nil {
		return nil, "", err
	}

	buffer := bytes.NewBufferString("")
	err = tmpl.Execute(buffer, map[string]interface{}{"prefix": pm.prefix})
	if err != nil {
		return nil, "", err
	}

	return io.NopCloser(bytes.NewReader(buffer.Bytes())), identifier, nil
}

func (pm *PrefixedMigration) ReadUp(version uint) (io.ReadCloser, string, error) {
	r, identifier, err := pm.Bindata.ReadUp(version)
	if err != nil {
		return nil, "", err
	}

	return pm.executeTemplate(r, identifier)
}

func (pm *PrefixedMigration) ReadDown(version uint) (io.ReadCloser, string, error) {
	r, identifier, err := pm.Bindata.ReadDown(version)
	if err != nil {
		return nil, "", err
	}

	return pm.executeTemplate(r, identifier)
}

func (s *SqlStore) Migrate() error {
	migrationsTable := fmt.Sprintf("%sschema_migrations", s.tablePrefix)

	driver, err := postgres.WithInstance(s.db.DB, &postgres.Config{MigrationsTable: migrationsTable})
	if err != nil {
		return err
	}

	bresource := bindata.Resource(migrations.AssetNames(), migrations.Asset)

	d, err := bindata.WithInstance(bresource)
	if err != nil {
		return err
	}
	prefixedData := &PrefixedMigration{
		Bindata:  d.(*bindata.Bindata),
		prefix:   s.tablePrefix,
		postgres: true,
	}

	m, err := migrate.NewWithInstance("prefixed-migration", prefixedData, "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
