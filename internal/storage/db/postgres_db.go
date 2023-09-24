package storage

import (
	"context"
	"database/sql"
	"enrich-fio/internal/config"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// TODO move to config
const (
	host     = "postgres"
	port     = 5432
	user     = ""
	password = ""
	dbname   = ""
)

// Storage is storage implementation via postgresql.
type Storage struct {
	db *sql.DB
}

// New returns storage implemented with postgresql.
func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// connect connects to Storage. TODO context should pass around logger
func (s *Storage) connect(ctx context.Context) (err error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	s.db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return errors.Wrap(err, "open db connection")
	}

	err = s.db.Ping()
	if err != nil {
		return errors.Wrap(err, "ping db")
	}
	fmt.Println("database connected successfully")
	return nil
}
func (s *Storage) close() {
	s.db.Close()
}

func (s *Storage) MigrateUp(ctx context.Context) error {
	err := s.connect(ctx)
	if err != nil {
		return errors.Wrap(err, "connect db")
	}
	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "get postgres driver")
	}
	m, err := migrate.NewWithDatabaseInstance(
		config.Conf.MigrationURL,
		dbname, driver)
	if err != nil {
		return errors.Wrap(err, "get migrate instance")
	}
	err = m.Up()
	if err != nil {
		return errors.Wrap(err, "migrate up")
	}
	return nil
}

func (s *Storage) AddPerson(ctx context.Context) {
}
