package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaHost    string
	MigrationURL string
}

var Conf Config

func Load() {
	godotenv.Load()
	Conf.KafkaHost = os.Getenv("KAFKA_HOST")
	Conf.MigrationURL = os.Getenv("MIGRATION_URL")
}
