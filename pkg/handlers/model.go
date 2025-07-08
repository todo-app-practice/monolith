package handlers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"strconv"
)

type EndpointHandler interface {
	AddEndpoints()
}

type Endpoint struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

func Method(e *echo.Echo, method string, path string, handler echo.HandlerFunc) {
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

func GetUrlId(ctx echo.Context, logger *zap.SugaredLogger) (uint, error) {
	idString := ctx.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		logger.Warn("could not parse id", "error", err.Error())

		return 0, err
	}

	return uint(id), nil
}
