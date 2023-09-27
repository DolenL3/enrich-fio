package config

import (
	"os"
)

// Cached is config with sensitive data, needed for working with cache.
type CacheConfig struct {
	Host string
}

// NewCacheConfig return CacheConfig with sensitive data, needed for working with cache.
func NewCacheConfig() *CacheConfig {
	return &CacheConfig{
		Host: os.Getenv("REDIS_HOST"),
	}
}

// DBConfig is config with sensitive data, needed for working with db.
type DBConfig struct {
	User         string
	Host         string
	DBName       string
	Password     string
	MigrationURL string
}

// NewDBConfig return DBConfig with sensitive data, needed for working with db.
func NewDBConfig() *DBConfig {
	return &DBConfig{
		User:         os.Getenv("POSTGRES_USER"),
		Host:         os.Getenv("POSTGRES_HOST"),
		DBName:       os.Getenv("POSTGRES_DB"),
		Password:     os.Getenv("POSTGRES_PASSWORD"),
		MigrationURL: os.Getenv("MIGRATION_URL"),
	}
}

// GraphQLConfig is config with sensitive data, needed for graphQL handler.
type GraphQLConfig struct {
	Host string
}

// NewGraphQLConfig returns GraphQLConfig with sensitive data, needed for graphQL handler.
func NewGraphQLConfig() *GraphQLConfig {
	return &GraphQLConfig{
		Host: os.Getenv("GRAPHQL_HOST"),
	}
}

// RestConfig is config with sensitive data, needed for rest API.
type RestConfig struct {
	Host string
}

// NewRestConfig returns RestConfig with sensitive data, needed for rest API.
func NewRestConfig() *RestConfig {
	return &RestConfig{
		Host: os.Getenv("REST_API_HOST"),
	}
}

// KafkaConfig is config with sensitive data, needed for kafka handler.
type KafkaConfig struct {
	Host string
}

// NewKafkaConfig returns KafkaConfig with sensitive data, needed for kafka handler.
func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Host: os.Getenv("KAFKA_HOST"),
	}
}
