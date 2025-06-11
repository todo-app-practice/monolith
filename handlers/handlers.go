package handlers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var (
	logger *zap.SugaredLogger
	e      *echo.Echo
)

func InitializeServer() {
	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync() // flushes buffer, if any
	logger = baseLogger.Sugar()

	e = echo.New()

	addEndpoints()

	logger.Fatal(e.Start(":8765"))
}

func addEndpoints() {
	e.GET("/", func(ctx echo.Context) error {
		logger.Infow("testing zappy...",
			"attempt", 3,
			"backoff", time.Second,
		)

		return ctx.String(http.StatusOK, "Hello, World!.")
	})

	e.GET("/todo-items", func(ctx echo.Context) error {
		logger.Infow("reading todo item...")

		return ctx.String(http.StatusOK, "read todo item.")
	})

	e.POST("/todo-items", func(ctx echo.Context) error {
		logger.Infow("creating todo item...")

		return ctx.String(http.StatusOK, "created todo item.")
	})

	e.PUT("/todo-items/:id", func(ctx echo.Context) error {
		logger.Infow("updating todo item...")

		return ctx.String(http.StatusOK, "updated todo item.")
	})

	e.DELETE("/todo-items/:id", func(ctx echo.Context) error {
		logger.Infow("deleting todo item...")

		return ctx.String(http.StatusOK, "deleted todo item.")
	})
}
