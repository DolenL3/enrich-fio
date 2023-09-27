package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"enrich-fio/internal/config"
	"enrich-fio/internal/models"
)

// Storage is storage implementation via postgresql.
type Storage struct {
	db     *pgxpool.Pool
	config *config.DBConfig
}

// New returns storage implemented with postgresql.
func New(db *pgxpool.Pool, config *config.DBConfig) *Storage {
	return &Storage{
		db:     db,
		config: config,
	}
}

func (s *Storage) connect(ctx context.Context) error {
	return nil
}

func (s *Storage) ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *Storage) close() {
	s.db.Close()
}

func (s *Storage) MigrateUp(ctx context.Context) error {
	logger := zap.L()
	p := &migratePgx.Postgres{}
	driver, err := p.Open(fmt.Sprintf("postgresql://%s:%s@%s/%s", s.config.User, s.config.Password, s.config.Host, s.config.DBName))
	if err != nil {
		return errors.Wrap(err, "opening connection")
	}
	m, err := migrate.NewWithDatabaseInstance(
		s.config.MigrationURL,
		"pgx", driver)
	if err != nil {
		return errors.Wrap(err, "get migrate instance")
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return errors.Wrap(err, "migrate up")
		}
		logger.Info("no change during migration")
		return nil
	}
	logger.Info("database migrated successfully")
	return nil
}

func (s *Storage) Save(ctx context.Context, person models.Person) error {
	query := `
	INSERT INTO person (id, name, surname, patronymic, gender, nationality, age)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.Exec(ctx, query, person.ID, person.Name, person.Surname, person.Patronymic,
		person.Gender, person.Nationality, person.Age)
	if err != nil {
		return errors.Wrap(err, "exec insert query")
	}
	return nil
}

var _resultsPerPage = 5

func (s *Storage) GetWithFilter(ctx context.Context, filter models.FilterConfig, page int) ([]models.Person, error) {
	query := `
	SELECT *
	FROM person

	`
	filters := []string{}
	if filter.ID != uuid.Nil {
		filters = append(filters, "ID = @ID")
	}
	if filter.Name != "" {
		filters = append(filters, "name = @name")
	}
	if filter.Surname != "" {
		filters = append(filters, "surname = @surname")
	}
	if filter.Patronymic != "" {
		filters = append(filters, "patronymic = @patronymic")
	}
	if filter.Age != (models.FilterAge{}) {
		filters = append(filters, "age BETWEEN @ageMin and @ageMax")
	}
	if filter.Gender != "" {
		filters = append(filters, "gender = @gender")
	}
	if filter.Nationality != "" {
		filters = append(filters, "nationality = @nationality")
	}

	args := pgx.NamedArgs{
		"ID":             filter.ID,
		"name":           filter.Name,
		"surname":        filter.Surname,
		"patronymic":     filter.Patronymic,
		"ageMin":         filter.Age.Min,
		"ageMax":         filter.Age.Max,
		"gender":         filter.Gender,
		"nationality":    filter.Nationality,
		"offset":         page * _resultsPerPage,
		"resultsPerPage": _resultsPerPage,
	}

	if len(filters) != 0 {
		query += `WHERE `
		query += strings.Join(filters, ` AND `)
	}

	query += `
	ORDER BY name
	LIMIT @resultsPerPage
	OFFSET @offset
	`

	rows, err := s.db.Query(ctx, query, args)
	if err != nil {
		return nil, errors.Wrap(err, "query people")
	}
	defer rows.Close()
	people, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Person])
	if err != nil {
		return []models.Person{}, errors.Wrap(err, "collect rows")
	}
	return people, nil
}

func (s *Storage) GetByID(ctx context.Context, id uuid.UUID) (models.Person, error) {
	query := `
	SELECT * FROM person
	WHERE ID = $1
	`
	rows, err := s.db.Query(ctx, query, id)
	if err != nil {
		return models.Person{}, err
	}
	person, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Person])
	return person, nil
}

func (s *Storage) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := `
	DELETE FROM person
	WHERE ID = $1
	`
	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "exec delete query")
	}
	return nil
}

func (s *Storage) ChangeByID(ctx context.Context, id uuid.UUID, change models.ChangeConfig) error {
	query := `
	UPDATE person
	SET %s
	WHERE id = @currentID
	`
	changes := []string{}
	if change.ID != uuid.Nil {
		changes = append(changes, "ID = @ID")
	}
	if change.Name != "" {
		changes = append(changes, "name = @name")
	}
	if change.Surname != "" {
		changes = append(changes, "surname = @surname")
	}
	if change.Patronymic != "" {
		changes = append(changes, "patronymic = @patronymic")
	}
	if change.Age != 0 {
		changes = append(changes, "age = @age")
	}
	if change.Gender != "" {
		changes = append(changes, "gender = @gender")
	}
	if change.Nationality != "" {
		changes = append(changes, "nationality = @nationality")
	}

	if len(changes) == 0 {
		return models.ErrNoChangesMade
	}
	query = fmt.Sprintf(query, strings.Join(changes, ", "))
	args := pgx.NamedArgs{
		"ID":          change.ID,
		"name":        change.Name,
		"surname":     change.Surname,
		"patronymic":  change.Patronymic,
		"age":         change.Age,
		"gender":      change.Gender,
		"nationality": change.Nationality,
		"currentID":   id,
	}
	_, err := s.db.Exec(ctx, query, args)
	if err != nil {
		return errors.Wrap(err, "exec update query")
	}
	return nil
}
