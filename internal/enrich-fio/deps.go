package enrichfio

import (
	"context"

	"github.com/google/uuid"

	"enrich-fio/internal/models"
)

// Storage is interface to interact with storage.
type Storage interface {
	// Save saves given person in storage.
	Save(ctx context.Context, person models.Person) error
	// GetWithFilter returns []models.Person that match the given filter.
	// Returns models.ErrPersonNotFound if no such people found in the storage.
	GetWithFilter(ctx context.Context, filter models.FilterConfig, page int) ([]models.Person, error)
	// GetByID returns one models.Person by given ID.
	// Returns models.ErrPersonNotFound if no such people found in the storage.
	GetByID(ctx context.Context, id uuid.UUID) (models.Person, error)
	// DeleteByID deletes person from storage by given ID.
	// Returns models.ErrPersonNotFound if no such people found in the storage.
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// ChangeByID applies given changes person from storage by given ID.
	// Returns models.ErrPersonNotFound if no such people found in the storage.
	ChangeByID(ctx context.Context, id uuid.UUID, changes models.ChangeConfig) error
	// MigrateUp performs a database migration to the last available version.
	MigrateUp(ctx context.Context) error
}

// ProbableGender is interface to get the most likely gender for a given person.
type ProbableGender interface {
	// Get returns the most likely gender for a given person.
	Get(ctx context.Context, name string, surname string, patronymic string) (models.Gender, error)
}

// ProbableAge is inetrace to get the most likely age for a given person.
type ProbableAge interface {
	// Get returns the most likely age for a given person.
	Get(ctx context.Context, name string, surname string, patronymic string) (int, error)
}

// ProbableNationality is interface to get the most likely nationality for a given person.
type ProbableNationality interface {
	// Get returns the most likely nationality for a given person
	Get(ctx context.Context, name string, surname string, patronymic string) (string, error)
}
