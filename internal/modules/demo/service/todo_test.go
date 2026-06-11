package service_test

// 本测试文件固定 Demo Todo 领域服务的事务和校验行为，防止注释补全和后续重构改变外部可观察行为。

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	"github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	"github.com/rei0721/go-scaffold/internal/modules/demo/service"
	"github.com/rei0721/go-scaffold/pkg/database"
)

// TestTodoServiceCRUD 固定 Demo Todo 领域服务的事务和校验行为，确保后续注释补全或结构调整不改变该场景。
func TestTodoServiceCRUD(t *testing.T) {
	todoService := newTodoService(t)
	ctx := context.Background()

	created, err := todoService.Create(ctx, service.CreateTodoInput{
		Title:       "  Write CRUD tests  ",
		Description: "  cover create list get update delete  ",
	})
	if err != nil {
		t.Fatalf("create todo: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created todo to have an id")
	}
	if created.Title != "Write CRUD tests" {
		t.Fatalf("expected trimmed title, got %q", created.Title)
	}
	if created.Description != "cover create list get update delete" {
		t.Fatalf("expected trimmed description, got %q", created.Description)
	}
	if created.Completed {
		t.Fatal("expected created todo to default to not completed")
	}

	listed, err := todoService.List(ctx)
	if err != nil {
		t.Fatalf("list todos: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 listed todo, got %d", len(listed))
	}
	if listed[0].ID != created.ID {
		t.Fatalf("expected listed todo id %d, got %d", created.ID, listed[0].ID)
	}

	fetched, err := todoService.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get todo: %v", err)
	}
	if fetched.Title != created.Title {
		t.Fatalf("expected fetched title %q, got %q", created.Title, fetched.Title)
	}

	updatedTitle := "  Ship CRUD tests  "
	updatedDescription := "  done through service  "
	completed := true
	updated, err := todoService.Update(ctx, created.ID, service.UpdateTodoInput{
		Title:       &updatedTitle,
		Description: &updatedDescription,
		Completed:   &completed,
	})
	if err != nil {
		t.Fatalf("update todo: %v", err)
	}
	if updated.Title != "Ship CRUD tests" {
		t.Fatalf("expected updated title to be trimmed, got %q", updated.Title)
	}
	if updated.Description != "done through service" {
		t.Fatalf("expected updated description to be trimmed, got %q", updated.Description)
	}
	if !updated.Completed {
		t.Fatal("expected updated todo to be completed")
	}

	if err := todoService.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete todo: %v", err)
	}
	if _, err := todoService.Get(ctx, created.ID); !errors.Is(err, service.ErrTodoNotFound) {
		t.Fatalf("expected deleted todo to be hidden, got %v", err)
	}
	listed, err = todoService.List(ctx)
	if err != nil {
		t.Fatalf("list todos after delete: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected deleted todo to be absent from list, got %d todos", len(listed))
	}
}

// TestTodoServiceValidationAndNotFound 固定 Demo Todo 领域服务的事务和校验行为，确保后续注释补全或结构调整不改变该场景。
func TestTodoServiceValidationAndNotFound(t *testing.T) {
	todoService := newTodoService(t)
	ctx := context.Background()

	if _, err := todoService.Create(ctx, service.CreateTodoInput{Title: "   "}); !errors.Is(err, service.ErrTodoTitleRequired) {
		t.Fatalf("expected blank create title to fail, got %v", err)
	}

	created, err := todoService.Create(ctx, service.CreateTodoInput{Title: "Keep me"})
	if err != nil {
		t.Fatalf("create todo: %v", err)
	}

	blankTitle := "  "
	if _, err := todoService.Update(ctx, created.ID, service.UpdateTodoInput{Title: &blankTitle}); !errors.Is(err, service.ErrTodoTitleRequired) {
		t.Fatalf("expected blank update title to fail, got %v", err)
	}

	if _, err := todoService.Get(ctx, created.ID+100); !errors.Is(err, service.ErrTodoNotFound) {
		t.Fatalf("expected missing get to fail with not found, got %v", err)
	}
	if _, err := todoService.Update(ctx, created.ID+100, service.UpdateTodoInput{}); !errors.Is(err, service.ErrTodoNotFound) {
		t.Fatalf("expected missing update to fail with not found, got %v", err)
	}
	if err := todoService.Delete(ctx, created.ID+100); !errors.Is(err, service.ErrTodoNotFound) {
		t.Fatalf("expected missing delete to fail with not found, got %v", err)
	}
}

// newTodoService 构造当前测试场景所需的最小依赖集合，避免测试直接耦合生产装配流程。
func newTodoService(t *testing.T) service.TodoService {
	t.Helper()

	db, err := database.New(&database.Config{
		Driver: database.DriverSQLite,
		DBName: filepath.Join(t.TempDir(), "demo.db"),
	})
	if err != nil {
		t.Fatalf("create sqlite database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite database: %v", err)
		}
	})

	if _, err := dbapp.ApplyDemoSchema(context.Background(), db, string(database.DriverSQLite)); err != nil {
		t.Fatalf("apply todo schema: %v", err)
	}

	return service.NewTodoService(db, repository.NewTodoRepository(db))
}
