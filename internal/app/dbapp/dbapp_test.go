package dbapp

// 本测试文件固定应用组装根的最小可启动契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	"github.com/rei0721/go-scaffold/pkg/database"
)

// TestDemoSchemaSQLUsesSQLGenDDL 固定应用组装根的最小可启动契约，确保后续注释补全或结构调整不改变该场景。
func TestDemoSchemaSQLUsesSQLGenDDL(t *testing.T) {
	sql, err := DemoSchemaSQL(string(database.DriverSQLite))
	if err != nil {
		t.Fatalf("DemoSchemaSQL() error = %v", err)
	}
	for _, want := range []string{
		`CREATE TABLE IF NOT EXISTS "demo_todos"`,
		`"id" INTEGER PRIMARY KEY AUTOINCREMENT`,
		`"title" TEXT NOT NULL`,
		`"deleted_at"`,
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("schema SQL %q does not contain %q", sql, want)
		}
	}
}

// TestDatabaseSQLUsesSQLGenDDL 固定应用组装根的最小可启动契约，确保后续注释补全或结构调整不改变该场景。
func TestDatabaseSQLUsesSQLGenDDL(t *testing.T) {
	sql, err := DatabaseSQL(string(database.DriverMySQL), "demo_app")
	if err != nil {
		t.Fatalf("DatabaseSQL() error = %v", err)
	}
	if sql != "CREATE DATABASE IF NOT EXISTS `demo_app`;" {
		t.Fatalf("DatabaseSQL() = %q, want sqlgen CREATE DATABASE", sql)
	}
}

// TestTodoOperationsUseSQLGenCRUD 固定应用组装根的最小可启动契约，确保后续注释补全或结构调整不改变该场景。
func TestTodoOperationsUseSQLGenCRUD(t *testing.T) {
	ctx := context.Background()
	db := newSQLiteDatabase(t)

	schemaSQL, err := ApplyDemoSchema(ctx, db, string(database.DriverSQLite))
	if err != nil {
		t.Fatalf("ApplyDemoSchema() error = %v", err)
	}
	if !strings.Contains(schemaSQL, "CREATE TABLE IF NOT EXISTS") {
		t.Fatalf("schema SQL = %q, want IF NOT EXISTS", schemaSQL)
	}
	hasTable, err := db.HasTable(ctx, &model.Todo{})
	if err != nil {
		t.Fatalf("HasTable() error = %v", err)
	}
	if !hasTable {
		t.Fatal("expected demo_todos table to exist")
	}

	createSQL, err := CreateTodo(ctx, db, string(database.DriverSQLite), TodoCreateInput{
		Title:       "Write db cli",
		Description: "generated through sqlgen",
		Completed:   false,
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if !strings.Contains(createSQL, `INSERT INTO "demo_todos"`) {
		t.Fatalf("create SQL = %q, want INSERT", createSQL)
	}

	todos, listSQL, err := ListTodos(ctx, db, string(database.DriverSQLite), 10, 0)
	if err != nil {
		t.Fatalf("ListTodos() error = %v", err)
	}
	if !strings.Contains(listSQL, `SELECT * FROM "demo_todos"`) {
		t.Fatalf("list SQL = %q, want SELECT", listSQL)
	}
	if len(todos) != 1 {
		t.Fatalf("len(todos) = %d, want 1", len(todos))
	}
	if todos[0].Title != "Write db cli" {
		t.Fatalf("todo title = %q, want Write db cli", todos[0].Title)
	}

	updatedTitle := "Ship db cli"
	completed := true
	updateSQL, err := UpdateTodo(ctx, db, string(database.DriverSQLite), TodoUpdateInput{
		ID:        todos[0].ID,
		Title:     &updatedTitle,
		Completed: &completed,
	})
	if err != nil {
		t.Fatalf("UpdateTodo() error = %v", err)
	}
	if !strings.Contains(updateSQL, `UPDATE "demo_todos" SET`) {
		t.Fatalf("update SQL = %q, want UPDATE", updateSQL)
	}

	got, getSQL, err := GetTodo(ctx, db, string(database.DriverSQLite), todos[0].ID)
	if err != nil {
		t.Fatalf("GetTodo() error = %v", err)
	}
	if !strings.Contains(getSQL, "LIMIT 1") {
		t.Fatalf("get SQL = %q, want LIMIT 1", getSQL)
	}
	if got.Title != updatedTitle || !got.Completed {
		t.Fatalf("GetTodo() = %#v, want updated title and completed", got)
	}

	deleteSQL, err := DeleteTodo(ctx, db, string(database.DriverSQLite), todos[0].ID)
	if err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if !strings.Contains(deleteSQL, `UPDATE "demo_todos" SET "deleted_at"=`) {
		t.Fatalf("delete SQL = %q, want soft delete UPDATE", deleteSQL)
	}

	if _, _, err := GetTodo(ctx, db, string(database.DriverSQLite), todos[0].ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("GetTodo() after delete error = %v, want record not found", err)
	}
}

// TestTodoOperationsValidateInputs 固定应用组装根的最小可启动契约，确保后续注释补全或结构调整不改变该场景。
func TestTodoOperationsValidateInputs(t *testing.T) {
	ctx := context.Background()
	db := newSQLiteDatabase(t)

	if _, err := CreateTodo(ctx, db, string(database.DriverSQLite), TodoCreateInput{}); !errors.Is(err, ErrMissingTodoTitle) {
		t.Fatalf("CreateTodo() error = %v, want ErrMissingTodoTitle", err)
	}
	if _, err := UpdateTodo(ctx, db, string(database.DriverSQLite), TodoUpdateInput{ID: 1}); !errors.Is(err, ErrNoTodoUpdate) {
		t.Fatalf("UpdateTodo() error = %v, want ErrNoTodoUpdate", err)
	}
	if _, err := DeleteTodo(ctx, db, string(database.DriverSQLite), 0); !errors.Is(err, ErrMissingTodoID) {
		t.Fatalf("DeleteTodo() error = %v, want ErrMissingTodoID", err)
	}
}

// newSQLiteDatabase 构造当前测试场景所需的最小依赖集合，避免测试直接耦合生产装配流程。
func newSQLiteDatabase(t *testing.T) database.Database {
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
	return db
}
