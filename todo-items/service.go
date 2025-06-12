package todo_items

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

type Service interface {
	CreateToDoItem(ctx *echo.Context, toDoItemText string) error
}

type service struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func GetService(logger *zap.SugaredLogger) Service {
	dbString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", "root", "root", "mysql", "todo")
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

func (s *service) CreateToDoItem(ctx *echo.Context, toDoItemText string) error {
	// TODO

	return nil
}

func (s *service) initializeDb() error {
	err := s.db.AutoMigrate(&ToDoItem{})

	if err != nil {
		return err
	}

	return nil
}
