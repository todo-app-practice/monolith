package app_server

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/swaggo/echo-swagger"
	"os"
	_ "todo-app/docs"
	"todo-app/internal/todos"
	"todo-app/internal/users"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	todoRepository := todos.GetRepository(logger, db)
	userRepository := users.GetRepository(logger, db)

	v := validator.New()

	todoService := todos.GetService(logger, todoRepository, v)
	userService := users.GetService(logger, userRepository, v)

	todoEndpointHandler := todos.GetEndpointHandler(
		logger,
		todoService,
		e,
	)
	userEndpointHandler := users.GetEndpointHandler(
		logger,
		userService,
		e,
	)

	todoEndpointHandler.AddEndpoints()
	userEndpointHandler.AddEndpoints()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// adding all middlewares here
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 2 << 10,
		LogLevel:  log.ERROR,
	}))

	logger.Fatal(e.Start(":8765"))
}

func initializeDb() error {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbName)

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
	
	err = db.AutoMigrate(&users.User{})
	if err != nil {
		return err
	}

	return nil
}
