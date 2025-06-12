package todos

import (
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	Create(item *ToDoItem) error
	GetAll() ([]ToDoItem, error)
	UpdateById(id uint, item ToDoItemUpdateInput) (ToDoItem, error)
	DeleteById(id uint) error
}

type service struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func GetService(logger *zap.SugaredLogger, db *gorm.DB) Service {
	return &service{
		logger: logger,
		db:     db,
	}
}

func (s *service) Create(item *ToDoItem) error {
	result := s.db.Create(item)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *service) GetAll() ([]ToDoItem, error) {
	var items []ToDoItem
	result := s.db.Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

func (s *service) UpdateById(id uint, item ToDoItemUpdateInput) (ToDoItem, error) {
	updates := map[string]interface{}{}

	if item.Text != nil {
		updates["text"] = item.Text
	}
	if item.Done != nil {
		updates["done"] = item.Done
	}

	if len(updates) == 0 {
		return ToDoItem{}, errors.New("no updates found")
	}

	result := s.db.Model(&ToDoItem{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return ToDoItem{}, result.Error
	}

	var updatedItem ToDoItem

	s.db.First(&updatedItem, id)

	return updatedItem, nil
}

func (s *service) DeleteById(id uint) error {
	result := s.db.Delete(&ToDoItem{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
