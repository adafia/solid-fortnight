package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// DSN returns a data source name for the postgres database.
func DSN(host, user, password, dbname string, port int) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

// NewDB creates a new database connection.
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs the database migrations.
func Migrate(db *sql.DB, migrationsPath string) error {
	log.Printf("Running migrations from path: %s", migrationsPath)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("Failed to get migration version: %v", err)
		return err
	}
	log.Printf("Current migration version: %v, dirty: %v", version, dirty)

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("Migration failed: %v", err)
		return err
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply")
	} else {
		newVersion, _, _ := m.Version()
		log.Printf("Migrations applied successfully. New version: %v", newVersion)
	}
	return nil
}
