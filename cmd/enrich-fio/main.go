package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"enrich-fio/internal/config"
	graphql "enrich-fio/internal/controllers/graphQL"
	"enrich-fio/internal/controllers/kafka"
	"enrich-fio/internal/controllers/rest"
	enrichfio "enrich-fio/internal/enrich-fio"
	probableage "enrich-fio/internal/enrich-fio/api/probable-age"
	probablegender "enrich-fio/internal/enrich-fio/api/probable-gender"
	probablenationality "enrich-fio/internal/enrich-fio/api/probable-nationality"
	"enrich-fio/internal/enrich-fio/storage"
	"enrich-fio/internal/enrich-fio/storage/cache"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	err := run()
	if err != nil {
		log.Printf("run app: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// load .env file
	godotenv.Load()
	// Collecting prerequisites.
	gin.SetMode(gin.ReleaseMode)
	ctx := context.Background()
	client := &http.Client{}
	logger, err := zap.NewDevelopment()
	if err != nil {
		return errors.Wrap(err, "initialising logger")
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Creating cached storage with cache realisation via redis and storage realisation via postgres/pgx.
	cacheConfig := config.NewCacheConfig()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cacheConfig.Host,
		Password: "",
		DB:       0,
	})
	dbConfig := config.NewDBConfig()
	pgxConfig, err := pgxpool.ParseConfig(fmt.Sprintf("postgresql://%s:%s@%s/%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.DBName))
	if err != nil {
		return errors.Wrap(err, "parsing pgx config")
	}
	pgxConfig.MaxConnIdleTime = time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return errors.Wrap(err, "creating new pgx pool")
	}
	var cacheTTL time.Duration = time.Hour
	s := cache.NewCacheStorage(storage.New(pool, dbConfig), redisClient, cacheTTL)

	// Applying the last version of storage schema.
	err = s.MigrateUp(ctx)
	if err != nil {
		return errors.Wrap(err, "migrating storage up")
	}

	// Creating probable age/gender/nationality realisations.
	pa := probableage.New(client)
	pg := probablegender.New(client)
	pn := probablenationality.New(client)

	// Creating enrich-fio service from collected dependencies.
	service := enrichfio.New(s, pa, pg, pn)

	// Creating controllers.
	graphQLHandler := graphql.NewGraphQLHandler(service, config.NewGraphQLConfig())

	router := gin.Default()
	httpHandler := rest.NewHTTPHandler(router, service, config.NewRestConfig())

	kafkaHandler := kafka.NewHandler(service, config.NewKafkaConfig())

	// Starting all the controllers.
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		err = graphQLHandler.Start(ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("GraphQL handler died. err %v", err))
		}
	}()

	go func() {
		defer wg.Done()
		err = httpHandler.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("http handler died. err: %v", err))
		}
	}()

	go func() {
		defer wg.Done()
		err := kafkaHandler.Start(ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("kafka handler died. err: %v", err))
		}
	}()

	wg.Wait()
	if err != nil {
		return errors.Wrap(err, "all handlers died")
	}
	return nil
}
