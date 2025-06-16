package todos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	testDBHost     = "localhost"
	testDBPort     = "3307" // Using the test DB port
	testDBUser     = "root"
	testDBPassword = "root"
	testDBName     = "todo_test"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Connect to the test database container
	dsn := testDBUser + ":" + testDBPassword + "@tcp(" + testDBHost + ":" + testDBPort + ")/" + testDBName + "?charset=utf8mb4&parseTime=True&loc=Local"

	// Try to connect with retries (container might need time to start)
	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	require.NoError(t, err, "Failed to connect to test database after retries")

	// Auto migrate the schema
	err = db.AutoMigrate(&ToDoItem{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// Clean up the database by dropping all tables
	err := db.Migrator().DropTable(&ToDoItem{})
	require.NoError(t, err, "Failed to clean up test database")
}

func setupTestRepository(t *testing.T) (Repository, func()) {
	// Create a test logger
	logger := zap.NewNop().Sugar()

	// Setup test database
	db := setupTestDB(t)

	// Create repository instance
	repo := GetRepository(logger, db)

	// Return cleanup function
	cleanup := func() {
		cleanupTestDB(t, db)
	}

	return repo, cleanup
}

func TestRepository_Create(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		item    *ToDoItem
		wantErr bool
	}{
		{
			name: "create valid todo",
			item: &ToDoItem{
				Text: "Test todo",
				Done: false,
			},
			wantErr: false,
		},
		{
			name: "create todo with empty text",
			item: &ToDoItem{
				Text: "",
				Done: false,
			},
			wantErr: false, // Repository does not validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.item)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.item.ID)
				assert.NotZero(t, tt.item.CreatedAt)
				assert.NotZero(t, tt.item.UpdatedAt)
			}
		})
	}
}

func TestRepository_GetAll(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create some test todos
	todos := []ToDoItem{
		{Text: "Todo 1", Done: false},
		{Text: "Todo 2", Done: true},
		{Text: "Todo 3", Done: false},
	}

	for i := range todos {
		err := repo.Create(ctx, &todos[i])
		require.NoError(t, err)
	}

	// Test getting all todos
	got, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, got, len(todos))

	// Verify todos are returned in the correct order
	for i, todo := range got {
		assert.Equal(t, todos[i].Text, todo.Text)
		assert.Equal(t, todos[i].Done, todo.Done)
	}
}

func TestRepository_Update(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test todo
	todo := &ToDoItem{
		Text: "Original text",
		Done: false,
	}
	err := repo.Create(ctx, todo)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uint
		updates map[string]interface{}
		want    ToDoItem
		wantErr bool
	}{
		{
			name: "update text only",
			id:   todo.ID,
			updates: map[string]interface{}{
				"text": "Updated text",
			},
			want: ToDoItem{
				Text: "Updated text",
				Done: false,
			},
			wantErr: false,
		},
		{
			name: "update done status only",
			id:   todo.ID,
			updates: map[string]interface{}{
				"done": true,
			},
			want: ToDoItem{
				Text: "Updated text", // Should keep previous text
				Done: true,
			},
			wantErr: false,
		},
		{
			name: "update both fields",
			id:   todo.ID,
			updates: map[string]interface{}{
				"text": "New text",
				"done": false,
			},
			want: ToDoItem{
				Text: "New text",
				Done: false,
			},
			wantErr: false,
		},
		{
			name:    "update non-existent todo",
			id:      999,
			updates: map[string]interface{}{"text": "Won't update"},
			wantErr: false, // GORM doesn't error on update with no rows affected.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.id, tt.updates)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.id != 999 {
					got, err := repo.GetById(ctx, tt.id)
					require.NoError(t, err)
					assert.Equal(t, tt.want.Text, got.Text)
					assert.Equal(t, tt.want.Done, got.Done)
				}
			}
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test todo
	todo := &ToDoItem{
		Text: "To be deleted",
		Done: false,
	}
	err := repo.Create(ctx, todo)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing todo",
			id:      todo.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent todo",
			id:      999,
			wantErr: false, // GORM doesn't return error for non-existent deletes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify todo is deleted
				_, err := repo.GetById(ctx, tt.id)
				assert.Error(t, err)
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			}
		})
	}
}
