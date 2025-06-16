package app_server

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"todo-app/internal/todos"
)

var (
	logger *zap.SugaredLogger
	e      *echo.Echo
	db     *gorm.DB
)

func InitializeServer() {
	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync() // flushes buffer, if any
	logger = baseLogger.Sugar()

	e = echo.New()
	e.HideBanner = true

	err := initializeDb()
	if err != nil {
		logger.Errorw("failed to initialize database", "error", err)

		os.Exit(1)
	}
	logger.Info("initialized database")

	todoService := todos.GetService(logger, db)

	todoEndpointHandler := todos.GetEndpointHandler(
		logger,
		todoService,
		e,
	)

	todoEndpointHandler.AddEndpoints()

	// adding all middlewares here
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 2 << 10,
		LogLevel:  log.ERROR,
	}))

	logger.Fatal(e.Start(":8765"))
}

func initializeDb() error {
	dbString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", "root", "root", "mysql", "todo")
	var err error
	db, err = gorm.Open(mysql.Open(dbString), &gorm.Config{})
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)

		os.Exit(1)
	}
	logger.Info("connected to database")

	err = migrateDb()
	if err != nil {
		return err
	}

	return nil
}

func migrateDb() error {
	err := db.AutoMigrate(&todos.ToDoItem{})
	if err != nil {
		return err
	}

	return nil
}
