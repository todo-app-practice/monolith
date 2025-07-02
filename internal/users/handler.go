package users

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	e "todo-app/pkg/errors"
	"todo-app/pkg/handlers"
	"todo-app/pkg/locale"
)

type endpointHandler struct {
	logger  *zap.SugaredLogger
	service Service
	e       *echo.Echo
}

func GetEndpointHandler(
	logger *zap.SugaredLogger,
	service Service,
	e *echo.Echo,
) handlers.EndpointHandler {
	return &endpointHandler{
		logger:  logger,
		service: service,
		e:       e,
	}
}

func (h *endpointHandler) AddEndpoints() {
	var endpoints = []handlers.Endpoint{
		{
			Method:  http.MethodPost,
			Path:    "/user",
			Handler: h.create,
		},
		{
			Method:  http.MethodPut,
			Path:    "/user/:id",
			Handler: h.update,
		},
	}

	for _, endpoint := range endpoints {
		handlers.Method(h.e, endpoint.Method, endpoint.Path, endpoint.Handler)
	}
}

func (h *endpointHandler) create(ctx echo.Context) error {
	h.logger.Infow("creating user...")

	user := User{}
	err := ctx.Bind(&user)
	if err != nil {
		h.logger.Warn("could not bind body to user struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadUser})
	}

	err = h.service.Create(ctx.Request().Context(), &user)
	if err != nil {
		h.logger.Warn("could not create user", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidUser, Details: err.Error()})
	}
	h.logger.Infow("created user successfully")

	user.Password = ""
	return ctx.JSON(http.StatusOK, user)
}

func (h *endpointHandler) update(ctx echo.Context) error {
	h.logger.Infow("updating user...")

	user := User{}
	err := ctx.Bind(&user)
	if err != nil {
		h.logger.Warn("could not bind body to user struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadUser})
	}

	id, err := handlers.GetUrlId(ctx, h.logger)
	if err != nil {
		h.logger.Warn("could not get id from url", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidID})
	}
	user.ID = id

	user, err = h.service.Update(ctx.Request().Context(), &user)
	if err != nil {
		h.logger.Warn("could not update user", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: err.Error()})
	}
	h.logger.Infow("updated user successfully")

	return ctx.JSON(http.StatusOK, user)
}
