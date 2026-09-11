package handlers

import (
	"accroccator/internal/api/routes"
	"accroccator/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartEndpoints(cardrepo *repository.CardRepository, containerRepo *repository.ContainerRepository) {
	router := gin.Default()
	setCorsConfig(router)

	handler := routes.NewHandler(cardrepo, containerRepo)

	handler.CreateCardRoutes(router)
	handler.CreateContainerRoutes(router)
	router.Run("localhost:8080")
}

func setCorsConfig(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))
}
