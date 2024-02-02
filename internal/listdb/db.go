package listdb

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"members-platform/internal/listdb/queries"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrate/*.sql
var migrations embed.FS

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.24.0 generate

var DB *queries.Queries

func ConnectPG(doMigrate bool) error {
	log.Println("connecting to postgres")
	url := os.Getenv("LIST_DATABASE_URL")
	if url == "" {
		return fmt.Errorf("missing LIST_DATABASE_URL in environment")
	}
	pg, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}
	if doMigrate {
		source, err := iofs.New(migrations, "migrate")
		if err != nil {
			return fmt.Errorf("set up migration source: %s", err)
		}
		database, err := postgres.WithInstance(pg, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("set up migration db connection: %w", err)
		}
		m, err := migrate.NewWithInstance(
			"file://migrate",
			source,
			url,
			database,
		)
		if err != nil {
			return fmt.Errorf("set up migration instance: %w", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("run migrations: %w", err)
		}
	}
	DB = queries.New(pg)
	return nil
}
