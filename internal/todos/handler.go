package todos

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type EndpointHandler interface {
	AddEndpoints()
}

type endpointHandler struct {
	logger  *zap.SugaredLogger
	service Service
	e       *echo.Echo
}

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
		Handler: getAll,
	},
	{
		Method:  http.MethodPost,
		Path:    "/todos",
		Handler: create,
	},
	{
		Method:  http.MethodPut,
		Path:    "/todos/:id",
		Handler: updateById,
	},
	{
		Method:  http.MethodDelete,
		Path:    "/todos/:id",
		Handler: deleteById,
	},
}

var h *endpointHandler

func GetEndpointHandler(logger *zap.SugaredLogger, service Service, e *echo.Echo) EndpointHandler {
	h = &endpointHandler{
		logger:  logger,
		service: service,
		e:       e,
	}

	return h
}

func (handler *endpointHandler) AddEndpoints() {
	for _, endpoint := range endpoints {
		h.logger.Infow("adding endpoint", "method", endpoint.Method, "path", endpoint.Path)
		method(handler.e, endpoint.Method, endpoint.Path, endpoint.Handler)
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
		h.logger.Warn("could not parse id", "error", err.Error())

		return 0, err
	}

	return uint(id), nil
}

func hello(ctx echo.Context) error {
	h.logger.Infow("testing zappy...",
		"attempt", 3,
		"backoff", time.Second,
	)

	return ctx.String(http.StatusOK, "Hello, World!")
}

func getAll(ctx echo.Context) error {
	h.logger.Infow("reading todo item...")

	items, err := h.service.GetAll()
	if err != nil {
		h.logger.Warn("could not read todo items", "error", err.Error())

		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, items)
}

func create(ctx echo.Context) error {
	h.logger.Infow("creating todo item...")

	item := ToDoItem{}
	err := ctx.Bind(&item)
	if err != nil {
		h.logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not read todo-item")
	}

	err = h.service.Create(&item)
	if err != nil {
		h.logger.Warn("could not create todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not create todo-item")
	}
	h.logger.Infow("created todo item successfully")

	return ctx.JSON(http.StatusOK, item)
}

func updateById(ctx echo.Context) error {
	h.logger.Infow("updating todo item...")

	item := ToDoItemUpdateInput{}
	err := ctx.Bind(&item)
	if err != nil {
		h.logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not read todo-item")
	}

	id, err := getUrlId(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "invalid id")
	}

	err = h.service.UpdateById(id, item)
	if err != nil {
		h.logger.Warn("could not update todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not update todo-item")
	}

	return ctx.String(http.StatusOK, "updated todo item")
}

func deleteById(ctx echo.Context) error {
	h.logger.Infow("deleting todo item...")

	id, err := getUrlId(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "invalid id")
	}

	err = h.service.DeleteById(id)
	if err != nil {
		h.logger.Warn("could not delete todo-item", "error", err.Error())

		return ctx.String(http.StatusBadRequest, "could not delete todo-item")
	}

	return ctx.String(http.StatusOK, "deleted todo item")
}
