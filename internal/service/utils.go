package utils

import (
	"accroccator/internal/model"
	"math"
)

func NewPaginatedResponse[T any](
	data []T,
	page int,
	pageSize int,
	totalItems int64,
) model.PaginatedResponse[T] {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	return model.PaginatedResponse[T]{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
