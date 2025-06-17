package todos

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	e "todo-app/pkg/errors"

	"go.uber.org/zap"
)

type Service interface {
	Create(ctx context.Context, item *ToDoItem) error
	GetAll(ctx context.Context) ([]ToDoItem, error)
	GetById(ctx context.Context, id uint) (ToDoItem, error)
	UpdateById(ctx context.Context, id uint, item ToDoItemUpdateInput) (ToDoItem, error)
	DeleteById(ctx context.Context, id uint) error
}

type service struct {
	logger     *zap.SugaredLogger
	repository Repository
	validator  *validator.Validate
}

func GetService(logger *zap.SugaredLogger, repo Repository, validator *validator.Validate) Service {
	return &service{
		logger:     logger,
		repository: repo,
		validator:  validator,
	}
}

func (s *service) Create(ctx context.Context, item *ToDoItem) error {
	if err := s.validator.Struct(item); err != nil {
		return e.ResponseError{Message: "invalid item", Details: err.Error()}
	}

	return s.repository.Create(ctx, item)
}

func (s *service) GetAll(ctx context.Context) ([]ToDoItem, error) {
	return s.repository.GetAll(ctx)
}

func (s *service) GetById(ctx context.Context, id uint) (ToDoItem, error) {
	return s.repository.GetById(ctx, id)
}

func (s *service) UpdateById(ctx context.Context, id uint, item ToDoItemUpdateInput) (ToDoItem, error) {
	updates := map[string]interface{}{}

	if item.Text != nil && *item.Text != "" {
		updates["text"] = *item.Text
	}
	if item.Done != nil {
		updates["done"] = *item.Done
	}

	if len(updates) == 0 {
		return ToDoItem{}, errors.New("no updates found")
	}

	err := s.repository.Update(ctx, id, updates)
	if err != nil {
		return ToDoItem{}, err
	}

	return s.repository.GetById(ctx, id)
}

func (s *service) DeleteById(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}
