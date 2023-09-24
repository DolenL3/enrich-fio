package main

import (
	"context"
	"enrich-fio/internal/config"
	"enrich-fio/internal/kafka"
	"log"
	"os"
)

func main() {
	err := run()
	if err != nil {
		log.Printf("run app: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	config.Load()
	ctx := context.Background()
	kafkaHandler := kafka.NewHandler()
	err := kafkaHandler.Start(ctx)
	if err != nil {
		return err
	}
	return nil
}
