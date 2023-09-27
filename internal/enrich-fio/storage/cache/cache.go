package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	enrichfio "enrich-fio/internal/enrich-fio"
	"enrich-fio/internal/models"
)

// CacheStorage is a wrapper for storage, that implements caching.
type CacheStorage struct {
	Storage enrichfio.Storage
	client  *redis.Client
	ttl     time.Duration
}

// NewCacheStorage returns CacheStorage, which implements caching.
func NewCacheStorage(storage enrichfio.Storage, client *redis.Client, ttl time.Duration) *CacheStorage {
	return &CacheStorage{
		Storage: storage,
		client:  client,
		ttl:     ttl,
	}
}

// Save saves given person in storage.
func (c *CacheStorage) Save(ctx context.Context, person models.Person) error {
	c.setPerson(ctx, person)
	return c.Storage.Save(ctx, person)
}

// GetWithFilter returns models.People that match the given filter.
// Returns models.ErrPersonNotFound if no such people found in the storage.
func (c *CacheStorage) GetWithFilter(ctx context.Context, filter models.FilterConfig, page int) ([]models.Person, error) {
	// Not implemented.
	return c.Storage.GetWithFilter(ctx, filter, page)
}

// GetByID returns one models.Person by given ID.
// Returns models.ErrPersonNotFound if no such people found in the storage.
func (c *CacheStorage) GetByID(ctx context.Context, id uuid.UUID) (models.Person, error) {
	logger := zap.L()
	result := c.client.Get(ctx, idKey(id))
	if result.Err() != nil {
		if errors.Is(result.Err(), redis.Nil) {
			logger.Info("no cache found")
		}
	}
	if result.Err() == nil {
		data, err := result.Bytes()
		if err != nil {
			logger.Warn("can't get cache data")
		}
		person := models.Person{}
		err = json.Unmarshal(data, &person)
		if err != nil {
			logger.Warn("can't unmarshal cached person")
		}
		return person, nil
	}
	logger.Info("no record matched in cache")
	person, err := c.Storage.GetByID(ctx, id)
	if err != nil {
		logger.Info("no record matched in storage")
		return models.Person{}, errors.Wrap(err, "get by id")
	}
	if person != (models.Person{}) {
		c.setPerson(ctx, person)
	}
	return person, nil
}

// DeleteByID deletes person from storage by given ID.
// Returns models.ErrPersonNotFound if no such people found in the storage.
func (c *CacheStorage) DeleteByID(ctx context.Context, id uuid.UUID) error {
	logger := zap.L()
	result := c.client.Del(ctx, idKey(id))
	if result.Err() != nil {
		if errors.Is(result.Err(), redis.Nil) {
			logger.Info("can't get cache")
		}
	}
	return c.Storage.DeleteByID(ctx, id)
}

// ChangeByID applies given changes person from storage by given ID.
// Returns models.ErrPersonNotFound if no such people found in the storage.
func (c *CacheStorage) ChangeByID(ctx context.Context, id uuid.UUID, changes models.ChangeConfig) error {
	// Not implemented. (Deletes from cache, not changes)
	logger := zap.L()
	result := c.client.Del(ctx, idKey(id))
	if result.Err() != nil {
		if errors.Is(result.Err(), redis.Nil) {
			logger.Info("can't get cache")
		}
	}
	return c.Storage.ChangeByID(ctx, id, changes)
}

// MigrateUp performs a database migration to the last available version.
func (c *CacheStorage) MigrateUp(ctx context.Context) error {
	return c.Storage.MigrateUp(ctx)
}

func (c *CacheStorage) setPerson(ctx context.Context, person models.Person) {
	logger := zap.L()
	personData, err := json.Marshal(person)
	if err != nil {
		logger.Warn(fmt.Sprintf("could marshal person to save in cache. Err: %v", err))
	}
	err = c.client.Set(ctx, idKey(person.ID), personData, c.ttl).Err()
	if err != nil {
		logger.Warn(fmt.Sprintf("could not save to cache. Err: %v", err))
	}
}

func idKey(id uuid.UUID) string {
	return id.String()
}
