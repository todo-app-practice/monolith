package handlers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
	todo_items2 "todo-app/internal/todos"
)

var (
	logger *zap.SugaredLogger
	e      *echo.Echo
	s      todo_items2.Service
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
		Path:    "/todos",
		Handler: getToDoItems,
	},
	{
		Method:  http.MethodPost,
		Path:    "/todos",
		Handler: createToDoItem,
	},
	{
		Method:  http.MethodPut,
		Path:    "/todos/:id",
		Handler: updateToDoItem,
	},
	{
		Method:  http.MethodDelete,
		Path:    "/todos/:id",
		Handler: deleteToDoItem,
	},
}

func InitializeServer() {
	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync() // flushes buffer, if any
	logger = baseLogger.Sugar()

	s = todo_items2.GetService(logger)

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

func getUrlId(ctx echo.Context) (uint, error) {
	idString := ctx.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		logger.Warn("could not parse id", "error", err.Error())

		return 0, err
	}

	return uint(id), nil
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

	items, err := s.GetToDoItems()
	if err != nil {
		logger.Warn("could not read todo items", "error", err.Error())

		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, items)
}

func createToDoItem(ctx echo.Context) error {
	logger.Infow("creating todo item...")

	item := todo_items2.ToDoItem{}
	err := ctx.Bind(&item)
	if err != nil {
		logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not read todo-item")
	}

	err = s.CreateToDoItem(&item)
	if err != nil {
		logger.Warn("could not create todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not create todo-item")
	}
	logger.Infow("created todo item successfully")

	return ctx.JSON(http.StatusOK, item)
}

func updateToDoItem(ctx echo.Context) error {
	logger.Infow("updating todo item...")

	item := todo_items2.ToDoItemUpdateInput{}
	err := ctx.Bind(&item)
	if err != nil {
		logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not read todo-item")
	}

	id, err := getUrlId(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "invalid id")
	}

	err = s.UpdateToDoItem(id, item)
	if err != nil {
		logger.Warn("could not update todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not update todo-item")
	}

	return ctx.String(http.StatusOK, "updated todo item")
}

func deleteToDoItem(ctx echo.Context) error {
	logger.Infow("deleting todo item...")

	id, err := getUrlId(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "invalid id")
	}

	err = s.DeleteToDoItem(id)
	if err != nil {
		logger.Warn("could not delete todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not delete todo-item")
	}

	return ctx.String(http.StatusOK, "deleted todo item")
}
