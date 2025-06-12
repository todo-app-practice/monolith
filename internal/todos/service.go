package todos

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

type Service interface {
	CreateToDoItem(item *ToDoItem) error
	GetToDoItems() ([]ToDoItem, error)
	UpdateToDoItem(id uint, item ToDoItemUpdateInput) error
	DeleteToDoItem(id uint) error
}

type service struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func GetService(logger *zap.SugaredLogger) Service {
	dbString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", "root", "root", "mysql", "todo")
	db, err := gorm.Open(mysql.Open(dbString), &gorm.Config{})
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)

		os.Exit(1)
	}
	logger.Info("connected to database")

	service := service{
		logger: logger,
		db:     db,
	}

	err = service.initializeDb()
	if err != nil {
		logger.Errorw("failed to initialize database", "error", err)

		os.Exit(1)
	}
	logger.Info("initialized database")

	return &service
}

func (s *service) initializeDb() error {
	err := s.db.AutoMigrate(&ToDoItem{})

	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateToDoItem(item *ToDoItem) error {
	result := s.db.Create(item)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *service) GetToDoItems() ([]ToDoItem, error) {
	var items []ToDoItem
	result := s.db.Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

func (s *service) UpdateToDoItem(id uint, item ToDoItemUpdateInput) error {
	updates := map[string]interface{}{}

	if item.Text != nil {
		updates["text"] = item.Text
	}
	if item.Done != nil {
		updates["done"] = item.Done
	}

	if len(updates) == 0 {
		return errors.New("no updates found")
	}

	result := s.db.Model(&ToDoItem{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *service) DeleteToDoItem(id uint) error {
	result := s.db.Delete(&ToDoItem{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
