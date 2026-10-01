package main

import (
	"accroccator/internal/api/handlers"
	"accroccator/internal/config"
	"accroccator/internal/repository"
	"accroccator/internal/service/refresh"
	"context"
	"log"
	"time"
)

func main() {

	err := config.SetEnvVariablesFromFile(".env")

	if err != nil {
		log.Printf("Error setting env variables: %v", err)
	}

	client, cardRepo, containerRepo, bulkRepo := repository.StartMongoDb()

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting: %v", err)
		}
	}()

	bulkDataUrl := "https://api.scryfall.com/bulk-data/all-cards"

	go refresh.StartScheduler(bulkDataUrl, bulkRepo, 12*time.Hour)

	handlers.StartEndpoints(cardRepo, containerRepo)
}
