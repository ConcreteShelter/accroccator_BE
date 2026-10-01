package routes

import (
	"accroccator/internal/model"
	"accroccator/internal/repository"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cardRepo      *repository.CardRepository
	containerRepo *repository.ContainerRepository
}

func NewHandler(
	cardRepo *repository.CardRepository,
	containerRepo *repository.ContainerRepository) *Handler {
	return &Handler{
		cardRepo:      cardRepo,
		containerRepo: containerRepo,
	}
}

func (h *Handler) CreateCardRoutes(router *gin.Engine) {
	router.GET("/cards/all", h.findAllPaginated)
	router.GET("/cards/search", h.findByName) // Changed!
	router.GET("/cards/:id", h.findByID)
	router.GET("/cards/names", h.searchCardNames)
	router.POST("/cards/:id/update_ownership", h.setOwnedQuantity)
	router.DELETE("/cards/:id", h.removeCard)
}

func (h *Handler) CreateContainerRoutes(router *gin.Engine) {
	router.POST("/containers/create", h.createContainer)
	router.GET("/containers/all", h.findAllContainers)
}

func (h *Handler) createContainer(c *gin.Context) {
	var create model.CreateContainerRequest
	if err := c.BindJSON(&create); err != nil {
		c.JSON(400, gin.H{"error": "Invalid body"})
		return
	}

	newContainer, err := h.containerRepo.CreateContainer(create)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError,
			gin.H{"error": "Could not create container"})
		return
	}

	c.IndentedJSON(http.StatusOK, newContainer)
}

func (h *Handler) findAllContainers(c *gin.Context) {
	containers, err := h.containerRepo.GetAllContainers(model.PlaceholderShopID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving containers"})
	}

	c.IndentedJSON(http.StatusOK, containers)
}

func (h *Handler) removeCard(c *gin.Context) {
	cardId := c.Param("id")
	err := h.cardRepo.RemoveOwnedCard(cardId, model.PlaceholderShopID)

	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No cards found with that name"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "generic error"})
		return

	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Card successfully deleted"})
}

func (h *Handler) setOwnedQuantity(c *gin.Context) {
	cardId := c.Param("id")

	var containerQuantityUpdates []model.QuantityUpdateRequest
	if err := c.BindJSON(&containerQuantityUpdates); err != nil {
		c.JSON(400, gin.H{"error": "Invalid body"})
		return
	}

	ownedCard, err := h.cardRepo.SetOwnedCard(cardId, containerQuantityUpdates, model.PlaceholderShopID)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error updating card quantity"})
		return
	}

	c.IndentedJSON(http.StatusOK, ownedCard)
}

func (h *Handler) findByName(c *gin.Context) {
	name := c.Query("name")

	cards, err := h.cardRepo.FindCardsByName(name, model.PlaceholderShopID)

	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No cards found with that name"})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving cards"})
		return
	}

	c.IndentedJSON(http.StatusOK, cards)
}

func (h *Handler) findByID(c *gin.Context) {
	id := c.Param("id")
	card, err := h.cardRepo.FindCardByID(id, model.PlaceholderShopID)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, card)
}

func (h *Handler) findAllPaginated(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	pageSizeStr := c.DefaultQuery("pageSize", "25")
	pageSize, _ := strconv.Atoi(pageSizeStr)

	fmt.Printf("pageStr is %v and pageSizeStr is %v", pageStr, pageSizeStr)

	cards, err := h.cardRepo.GetAllCards(page, pageSize, model.PlaceholderShopID)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving cards"})
		return
	}

	c.IndentedJSON(http.StatusOK, cards)
}

func (h *Handler) searchCardNames(c *gin.Context) {
	queryString := c.Query("q")

	if utf8.RuneCountInString(queryString) < 3 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "The query needs more than 3 chars"})
		return
	}

	results, err := h.cardRepo.SearchCardNames(queryString)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving names"})
		return
	}

	c.IndentedJSON(http.StatusOK, results)
}
