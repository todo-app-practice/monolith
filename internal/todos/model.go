package todos

import (
	"gorm.io/gorm"
)

type ToDoItem struct {
	gorm.Model
	Text string `gorm:"not null" validate:"required"`
	Done bool   `gorm:"default:false"`
}

type ToDoItemUpdateInput struct {
	Text *string `json:"text"`
	Done *bool   `json:"done"`
}

type PaginationDetails struct {
	Page  int
	Limit int
	Order string
}

type PaginationMetadata struct {
	ResultCount int
	TotalCount  int
}

type PaginatedResponse struct {
	Data []ToDoItem         `json:"data"`
	Meta PaginationMetadata `json:"metadata, omitempty"`
}
