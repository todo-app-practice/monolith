package handlers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
	todo_items "todo-app/todo-items"
)

var (
	logger *zap.SugaredLogger
	e      *echo.Echo
	s      todo_items.Service
)

type endpoint struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

var endpoints = []endpoint{
	{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: hello,
	},
	{
		Method:  http.MethodGet,
		Path:    "/todo-items",
		Handler: getToDoItems,
	},
	{
		Method:  http.MethodPost,
		Path:    "/todo-items",
		Handler: createToDoItem,
	},
	{
		Method:  http.MethodPut,
		Path:    "/todo-items/:id",
		Handler: updateToDoItem,
	},
	{
		Method:  http.MethodDelete,
		Path:    "/todo-items/:id",
		Handler: deleteToDoItem,
	},
}

func InitializeServer() {
	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync() // flushes buffer, if any
	logger = baseLogger.Sugar()

	s = todo_items.GetService(logger)

	e = echo.New()
	e.HideBanner = true

	addEndpoints()

	logger.Fatal(e.Start(":8765"))
}

func addEndpoints() {
	for _, endpoint := range endpoints {
		method(e, endpoint.Method, endpoint.Path, endpoint.Handler)
	}
}

func method(e *echo.Echo, method string, path string, handler echo.HandlerFunc) {
	switch method {
	case "GET":
		e.GET(path, handler)
	case "POST":
		e.POST(path, handler)
	case "PUT":
		e.PUT(path, handler)
	case "DELETE":
		e.DELETE(path, handler)
	default:
		panic("unsupported method: " + method)
	}
}

func hello(ctx echo.Context) error {
	logger.Infow("testing zappy...",
		"attempt", 3,
		"backoff", time.Second,
	)

	return ctx.String(http.StatusOK, "Hello, World!")
}

func getToDoItems(ctx echo.Context) error {
	logger.Infow("reading todo item...")

	return ctx.String(http.StatusOK, "read todo item")
}

func createToDoItem(ctx echo.Context) error {
	logger.Infow("creating todo item...")

	item := todo_items.ToDoItem{}
	err := ctx.Bind(&item)
	if err != nil {
		logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not read todo-item")
	}

	err = s.CreateToDoItem(&ctx, item)
	if err != nil {
		logger.Warn("could not create todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not create todo-item")
	}
	logger.Infow("created todo item successfully")

	return ctx.JSON(http.StatusOK, item)
}

func updateToDoItem(ctx echo.Context) error {
	logger.Infow("updating todo item...")

	return ctx.String(http.StatusOK, "updated todo item")
}

func deleteToDoItem(ctx echo.Context) error {
	logger.Infow("deleting todo item...")

	return ctx.String(http.StatusOK, "deleted todo item")
}
