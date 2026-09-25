package model

type PaginationRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

// Handler
type QuantityUpdateRequest struct {
	ContainerID string `json:"container_id,omitempty"`
	Quantity    int    `json:"quantity"`
}

type CreateContainerRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type CardAdvancedSearchRequest struct {
	PageParams PaginationRequest `json:"page_params"`
	Name       string            `json:"name"`
	Cmc        int               `json:"cmc"`
	Keywords   []string          `json:"keywords"`
	TypeLine   string            `json:"type_line"`
	IsOwned    bool              `json:"is_owned"`
}
