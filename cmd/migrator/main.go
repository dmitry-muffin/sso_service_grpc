package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	// Флаги командной строки
	host := flag.String("host", "localhost", "Database host")
	port := flag.Int("port", 5432, "Database port")
	user := flag.String("user", "postgres", "Database user")
	password := flag.String("password", "admin", "Database password")
	dbname := flag.String("dbname", "test_auth", "Database name")
	sslMode := flag.String("ssl", "disable", "SSL mode")
	migrationsPath := flag.String("migrations", "migrations", "Path to migrations")
	migrationsTable := flag.String("migrations-table", "schema_migrations", "Migrations table name")

	flag.Parse()

	// Формирование строки подключения
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		*host, *port, *user, *password, *dbname, *sslMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Connection test failed:", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: *migrationsTable,
	})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+*migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No new migrations to apply")
			return
		}
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Migrations applied successfully")
}
