package todos

import (
	"context"
	"testing"
	e "todo-app/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockRepository is a mock for the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, item *ToDoItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockRepository) GetAll(ctx context.Context) ([]ToDoItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]ToDoItem), args.Error(1)
}

func (m *MockRepository) GetById(ctx context.Context, id uint) (ToDoItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(ToDoItem), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockRepository)
	v := validator.New()
	service := GetService(nil, mockRepo, v)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		todo := &ToDoItem{Text: "buy milk"}
		mockRepo.On("Create", ctx, todo).Return(nil).Once()

		err := service.Create(ctx, todo)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		todo := &ToDoItem{Text: ""} // Empty text should fail validation

		err := service.Create(ctx, todo)
		assert.Error(t, err)
		responseErr, ok := err.(e.ResponseError)
		assert.True(t, ok)
		assert.Equal(t, "invalid item", responseErr.Message)
		mockRepo.AssertNotCalled(t, "Create", ctx, todo)
	})
}

func TestService_GetAll(t *testing.T) {
	mockRepo := new(MockRepository)
	v := validator.New()
	service := GetService(nil, mockRepo, v)
	ctx := context.Background()

	expectedTodos := []ToDoItem{
		{Text: "Todo 1", Done: false},
		{Text: "Todo 2", Done: true},
	}

	mockRepo.On("GetAll", ctx).Return(expectedTodos, nil).Once()

	todos, err := service.GetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos, todos)
	mockRepo.AssertExpectations(t)
}

func TestService_GetById(t *testing.T) {
	mockRepo := new(MockRepository)
	v := validator.New()
	service := GetService(nil, mockRepo, v)
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		expectedTodo := ToDoItem{Text: "found me"}
		expectedTodo.ID = 1
		mockRepo.On("GetById", ctx, uint(1)).Return(expectedTodo, nil).Once()

		todo, err := service.GetById(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedTodo, todo)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetById", ctx, uint(99)).Return(ToDoItem{}, gorm.ErrRecordNotFound).Once()

		_, err := service.GetById(ctx, 99)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_UpdateById(t *testing.T) {
	mockRepo := new(MockRepository)
	v := validator.New()
	service := GetService(nil, mockRepo, v)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		updateInput := ToDoItemUpdateInput{Text: stringPtr("updated text")}
		updates := map[string]interface{}{"text": "updated text"}
		updatedTodo := ToDoItem{Text: "updated text"}
		updatedTodo.ID = 1

		mockRepo.On("Update", ctx, uint(1), updates).Return(nil).Once()
		mockRepo.On("GetById", ctx, uint(1)).Return(updatedTodo, nil).Once()

		todo, err := service.UpdateById(ctx, 1, updateInput)
		assert.NoError(t, err)
		assert.Equal(t, updatedTodo, todo)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no updates provided", func(t *testing.T) {
		updateInput := ToDoItemUpdateInput{}

		_, err := service.UpdateById(ctx, 1, updateInput)
		assert.Error(t, err)
		assert.Equal(t, "no updates found", err.Error())
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("update fails", func(t *testing.T) {
		updateInput := ToDoItemUpdateInput{Done: boolPtr(true)}
		updates := map[string]interface{}{"done": true}
		mockRepo.On("Update", ctx, uint(1), updates).Return(gorm.ErrInvalidDB).Once()

		_, err := service.UpdateById(ctx, 1, updateInput)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_DeleteById(t *testing.T) {
	mockRepo := new(MockRepository)
	v := validator.New()
	service := GetService(nil, mockRepo, v)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, uint(1)).Return(nil).Once()

		err := service.DeleteById(ctx, 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete fails", func(t *testing.T) {
		mockRepo.On("Delete", ctx, uint(1)).Return(gorm.ErrInvalidDB).Once()

		err := service.DeleteById(ctx, 1)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)
		mockRepo.AssertExpectations(t)
	})
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
