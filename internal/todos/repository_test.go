package todos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (Repository, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&ToDoItem{})
	assert.NoError(t, err)

	logger := zap.NewNop().Sugar()
	repo := GetRepository(logger, db)

	cleanup := func() {
		sqlDB, err := db.DB()
		assert.NoError(t, err)
		sqlDB.Close()
	}

	return repo, cleanup
}

func TestRepository_Create(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	item := &ToDoItem{Text: "Test todo"}

	err := repo.Create(ctx, item)
	assert.NoError(t, err)
	assert.NotZero(t, item.ID)

	found, err := repo.GetById(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test todo", found.Text)
}

func TestRepository_GetAll(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create some items
	repo.Create(ctx, &ToDoItem{Text: "Todo 1"})
	repo.Create(ctx, &ToDoItem{Text: "Todo 2"})

	items, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestRepository_GetById(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	item := &ToDoItem{Text: "Test todo for get by id"}
	repo.Create(ctx, item)

	found, err := repo.GetById(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, item.Text, found.Text)
}

func TestRepository_GetById_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repo.GetById(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestRepository_Update(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	item := &ToDoItem{Text: "Initial Text"}
	repo.Create(ctx, item)

	updates := map[string]interface{}{"text": "Updated Text", "done": true}
	err := repo.Update(ctx, item.ID, updates)
	assert.NoError(t, err)

	updatedItem, err := repo.GetById(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Text", updatedItem.Text)
	assert.True(t, updatedItem.Done)
}

func TestRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	item := &ToDoItem{Text: "To be deleted"}
	repo.Create(ctx, item)

	err := repo.Delete(ctx, item.ID)
	assert.NoError(t, err)

	_, err = repo.GetById(ctx, item.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
