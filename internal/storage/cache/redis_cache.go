package cache

import "time"

type redisCache struct {
	host    string
	db      int
	expires time.Duration
}

func NewRedisCache(host string, db int, expires time.Duration) *redisCache {
	return &redisCache{
		host:    host,
		db:      db,
		expires: expires,
	}
}
