package app_server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "todo-app/docs"
	"todo-app/internal/auth"
	"todo-app/internal/todos"
	"todo-app/internal/users"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	echoSwagger "github.com/swaggo/echo-swagger"
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

	// Initialize repositories
	todoRepository := todos.GetRepository(logger, db)
	userRepository := users.GetRepository(logger, db)
	authRepository := auth.GetRepository(logger, db)

	v := validator.New()

	// Initialize services
	todoService := todos.GetService(logger, todoRepository, v)
	userService := users.GetService(logger, userRepository, v)
	authService := auth.GetService(logger, userRepository, authRepository, v)

	// Initialize handlers
	todoEndpointHandler := todos.GetEndpointHandler(logger, todoService, e)
	userEndpointHandler := users.GetEndpointHandler(logger, userService, e)
	authEndpointHandler := auth.GetEndpointHandler(logger, authService, e)

	jwtMiddleware := auth.JWTMiddleware(authService, logger)

	// Apply JWT middleware to protected routes
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip JWT validation for auth endpoints
			path := c.Request().URL.Path
			if path == "/login" || path == "/logout" || path == "/refresh" || path == "/user" || strings.Contains(path, "/swagger") {
				return next(c)
			}
			// Apply JWT middleware for all other routes
			return jwtMiddleware(next)(c)
		}
	})

	todoEndpointHandler.AddEndpoints()
	userEndpointHandler.AddEndpoints()
	authEndpointHandler.AddEndpoints()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Adding all middlewares here
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 2 << 10,
		LogLevel:  log.ERROR,
	}))

	// Start server in a goroutine
	go func() {
		if err := e.Start(":8765"); err != nil && err != http.ErrServerClosed {
			logger.Fatalw("shutting down the server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Fatalw("server shutdown failed", "error", err)
	}
	logger.Info("server shut down")

	logger.Info("closing database connection")
	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorw("failed to get underlying sql.DB", "error", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			logger.Errorw("failed to close database connection", "error", err)
		} else {
			logger.Info("database connection closed")
		}
	}
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

	err = db.AutoMigrate(&auth.RefreshToken{})
	if err != nil {
		return err
	}

	return nil
}
