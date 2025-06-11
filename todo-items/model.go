package todo_items

import (
	"gorm.io/gorm"
)

type ToDoItem struct {
	gorm.Model
	Done bool   `gorm:"default:false"`
	Text string `gorm:"not null"`
}
