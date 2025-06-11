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
	e.GET("/", func(c echo.Context) error {
		logger.Infow("testing zappy...",
			"attempt", 3,
			"backoff", time.Second,
		)

		return c.String(http.StatusOK, "Hello, World!")
	})
}
