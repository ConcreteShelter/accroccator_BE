package main

import (
	"accroccator/internal/api/handlers"
	"accroccator/internal/config"
	"accroccator/internal/repository"
	"context"
	"log"
)

func main() {

	err := config.SetEnvVariablesFromFile(".env")

	if err != nil {
		log.Printf("Error setting env variables: %v", err)
	}

	client, cardRepo, containerRepo := repository.StartMongoDb()

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting: %v", err)
		}
	}()

	handlers.StartEndpoints(cardRepo, containerRepo)
}
