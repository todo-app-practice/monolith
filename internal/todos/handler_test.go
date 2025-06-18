package todos

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	todoItemData := `{"text":"go for a run", "done": false}`
	todoItem := ToDoItem{
		Text: "go for a run",
		Done: false,
	}

	req := httptest.NewRequest(http.MethodPost, "/todos/", strings.NewReader(todoItemData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	mockService.
		EXPECT().
		Create(ctx.Request().Context(), &todoItem).
		Return(nil).
		Times(1)

	if assert.NoError(t, h.create(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// unmarshalling response to check if response data is ok
		var responseItem ToDoItem
		err := json.Unmarshal(rec.Body.Bytes(), &responseItem)
		assert.NoError(t, err)

		assert.Equal(t, todoItem.Text, responseItem.Text)
		assert.Equal(t, todoItem.Done, responseItem.Done)
	}
}
