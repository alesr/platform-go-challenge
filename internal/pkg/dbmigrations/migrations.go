package dbmigrations

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(db *sql.DB, dbName, path string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not get driver: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", path)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, dbName, driver)
	if err != nil {
		return fmt.Errorf("could not crerate  migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not apply migrations: %w", err)
	}
	return nil
}
