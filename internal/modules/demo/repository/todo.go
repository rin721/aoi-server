package repository

import (
	"context"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	database "github.com/rei0721/go-scaffold/internal/ports"
)

// TodoRepository defines the persistence port for Demo Todo.
type TodoRepository interface {
	WithExecutor(db database.Executor) TodoRepository
	Create(ctx context.Context, todo *model.Todo) error
	List(ctx context.Context) ([]model.Todo, error)
	FindByID(ctx context.Context, id uint) (*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id uint) error
}

type todoRepository struct {
	db database.Executor
}

// NewTodoRepository creates a Todo repository backed by the project database facade.
func NewTodoRepository(db database.Executor) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) WithExecutor(db database.Executor) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *model.Todo) error {
	return r.db.Create(ctx, todo)
}

func (r *todoRepository) List(ctx context.Context) ([]model.Todo, error) {
	var todos []model.Todo
	err := r.db.Find(ctx, &todos, aliveTodos(), database.Order("id DESC"))
	return todos, err
}

func (r *todoRepository) FindByID(ctx context.Context, id uint) (*model.Todo, error) {
	var todo model.Todo
	if err := r.db.First(ctx, &todo, database.Where("id = ?", id), aliveTodos()); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *model.Todo) error {
	return r.db.Save(ctx, todo)
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	now := time.Now().UTC()
	_, err := r.db.Update(ctx, &model.Todo{}, map[string]any{
		"deleted_at": now,
	}, database.Where("id = ?", id), aliveTodos())
	return err
}

func aliveTodos() database.QueryOption {
	return database.Where("deleted_at IS NULL")
}
