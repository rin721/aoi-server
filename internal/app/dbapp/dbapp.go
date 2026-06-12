// Package dbapp 提供 Demo Todo 对数据库的 SQLGen 桥接能力。
//
// 本包用于脚手架示例和本地验证：它返回生成 SQL 以便测试和审计，但不是生产迁移框架。
package dbapp

// 本文件连接 Demo 业务与数据库基础设施，用 SQL 生成器为命令行和启动钩子提供受控的数据操作。

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/sqlgen"
)

const (
	// DefaultTodoLimit 是列表查询未传入有效 limit 时的默认条数。
	DefaultTodoLimit = 100
	// MaxTodoLimit 限制 Demo 列表查询单次返回规模，避免示例接口被超大分页拖垮。
	MaxTodoLimit = 1000
)

var (
	// ErrUnsupportedDriver 表示当前 database driver 无法映射到 sqlgen 方言。
	ErrUnsupportedDriver = errors.New("unsupported database driver")
	// ErrMissingDatabase 表示调用方未注入数据库实例。
	ErrMissingDatabase = errors.New("database is nil")
	// ErrMissingTodoTitle 表示创建 Todo 时缺少必填标题。
	ErrMissingTodoTitle = errors.New("todo title is required")
	// ErrMissingTodoID 表示需要定位单条 Todo 的操作缺少有效 ID。
	ErrMissingTodoID = errors.New("todo id is required")
	// ErrNoTodoUpdate 表示更新请求没有提供任何可变更字段。
	ErrNoTodoUpdate = errors.New("no todo fields to update")
)

// TodoCreateInput 描述创建 Demo Todo 所需的业务字段。
type TodoCreateInput struct {
	Title       string
	Description string
	Completed   bool
}

// TodoUpdateInput 描述更新 Demo Todo 的局部字段。
//
// 指针字段用于区分“未传入”和“传入零值”，例如 Completed=false 仍然是一次有效更新。
type TodoUpdateInput struct {
	ID          uint
	Title       *string
	Description *string
	Completed   *bool
}

// DialectForDriver 将应用数据库 driver 映射为 sqlgen 方言。
func DialectForDriver(driver string) (sqlgen.Dialect, error) {
	switch database.Driver(driver) {
	case database.DriverSQLite:
		return sqlgen.SQLite, nil
	case database.DriverMySQL:
		return sqlgen.MySQL, nil
	case database.DriverPostgres:
		return sqlgen.PostgreSQL, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedDriver, driver)
	}
}

// NewGenerator 创建 Demo 数据操作使用的 sqlgen 生成器。
//
// Demo Todo 使用软删除，并跳过零值字段，以贴近常见业务 CRUD 语义。
func NewGenerator(driver string) (*sqlgen.Generator, error) {
	dialect, err := DialectForDriver(driver)
	if err != nil {
		return nil, err
	}
	return sqlgen.New(&sqlgen.Config{
		Dialect:       dialect,
		SoftDelete:    true,
		SkipZeroValue: true,
	}), nil
}

// DemoSchemaSQL 生成 Demo 示例表结构 SQL。
func DemoSchemaSQL(driver string) (string, error) {
	statements, err := demoSchemaStatements(driver)
	if err != nil {
		return "", err
	}
	return strings.Join(statements, "\n"), nil
}

func demoSchemaStatements(driver string) ([]string, error) {
	gen, err := NewGenerator(driver)
	if err != nil {
		return nil, err
	}
	models := []any{&model.Todo{}, &model.Customer{}}
	statements := make([]string, 0, len(models))
	for _, item := range models {
		sql, err := gen.TableIfNotExists(item)
		if err != nil {
			return nil, err
		}
		statements = append(statements, sql)
	}
	return statements, nil
}

// DatabaseSQL 生成创建数据库的 SQL。
func DatabaseSQL(driver, name string) (string, error) {
	gen, err := NewGenerator(driver)
	if err != nil {
		return "", err
	}
	return gen.DatabaseIfNotExists(name)
}

// ApplyDatabase 执行创建数据库 SQL，并返回实际生成的语句。
//
// 返回 SQL 是为了让测试和运维日志能确认 sqlgen 生成结果，而不是隐藏在副作用里。
func ApplyDatabase(ctx context.Context, db database.Database, driver, name string) (string, error) {
	if db == nil {
		return "", ErrMissingDatabase
	}
	sql, err := DatabaseSQL(driver, name)
	if err != nil {
		return "", err
	}
	if _, err := db.Exec(ctx, sql); err != nil {
		return sql, fmt.Errorf("apply database create: %w", err)
	}
	return sql, nil
}

// ApplyDemoSchema 执行 Demo 示例表结构 SQL，并返回实际生成的语句。
func ApplyDemoSchema(ctx context.Context, db database.Database, driver string) (string, error) {
	if db == nil {
		return "", ErrMissingDatabase
	}
	statements, err := demoSchemaStatements(driver)
	if err != nil {
		return "", err
	}
	for _, sql := range statements {
		if _, err := db.Exec(ctx, sql); err != nil {
			return strings.Join(statements, "\n"), fmt.Errorf("apply demo schema: %w", err)
		}
	}
	return strings.Join(statements, "\n"), nil
}

// CreateTodo 通过 sqlgen 生成并执行 Demo Todo 创建语句。
func CreateTodo(ctx context.Context, db database.Database, driver string, input TodoCreateInput) (string, error) {
	if db == nil {
		return "", ErrMissingDatabase
	}
	if input.Title == "" {
		return "", ErrMissingTodoTitle
	}

	now := time.Now().UTC()
	todo := model.Todo{
		Title:       input.Title,
		Description: input.Description,
		Completed:   input.Completed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gen, err := NewGenerator(driver)
	if err != nil {
		return "", err
	}
	sql, err := gen.Omit("ID", "DeletedAt").Create(&todo)
	if err != nil {
		return "", err
	}
	if _, err := db.Exec(ctx, sql); err != nil {
		return sql, fmt.Errorf("create demo todo: %w", err)
	}
	return sql, nil
}

// ListTodos 通过 sqlgen 生成并执行 Demo Todo 分页查询。
//
// limit 会被约束在默认值和 MaxTodoLimit 之间，offset 小于 0 时按 0 处理。
func ListTodos(ctx context.Context, db database.Database, driver string, limit, offset int) ([]model.Todo, string, error) {
	if db == nil {
		return nil, "", ErrMissingDatabase
	}
	limit = normalizeLimit(limit)
	if offset < 0 {
		offset = 0
	}

	gen, err := NewGenerator(driver)
	if err != nil {
		return nil, "", err
	}
	sql, err := gen.Order("id DESC").Limit(limit).Offset(offset).Find(&model.Todo{})
	if err != nil {
		return nil, "", err
	}

	var todos []model.Todo
	if _, err := db.Raw(ctx, &todos, sql); err != nil {
		return nil, sql, fmt.Errorf("list demo todos: %w", err)
	}
	return todos, sql, nil
}

// GetTodo 通过 sqlgen 生成并执行单条 Demo Todo 查询。
//
// 查询无结果时统一返回 database.ErrNotFound，便于 handler 层映射为 404。
func GetTodo(ctx context.Context, db database.Database, driver string, id uint) (*model.Todo, string, error) {
	if db == nil {
		return nil, "", ErrMissingDatabase
	}
	if id == 0 {
		return nil, "", ErrMissingTodoID
	}

	gen, err := NewGenerator(driver)
	if err != nil {
		return nil, "", err
	}
	sql, err := gen.First(&model.Todo{}, id)
	if err != nil {
		return nil, "", err
	}

	var todo model.Todo
	result, err := db.Raw(ctx, &todo, sql)
	if err != nil {
		return nil, sql, fmt.Errorf("get demo todo: %w", err)
	}
	if result.RowsAffected == 0 {
		return nil, sql, database.ErrNotFound
	}
	return &todo, sql, nil
}

// UpdateTodo 通过 sqlgen 生成并执行 Demo Todo 局部更新。
//
// 只会更新指针字段中非 nil 的值，并始终刷新 updated_at。
func UpdateTodo(ctx context.Context, db database.Database, driver string, input TodoUpdateInput) (string, error) {
	if db == nil {
		return "", ErrMissingDatabase
	}
	if input.ID == 0 {
		return "", ErrMissingTodoID
	}

	values := map[string]interface{}{}
	if input.Title != nil {
		values["title"] = *input.Title
	}
	if input.Description != nil {
		values["description"] = *input.Description
	}
	if input.Completed != nil {
		values["completed"] = *input.Completed
	}
	if len(values) == 0 {
		return "", ErrNoTodoUpdate
	}
	values["updated_at"] = time.Now().UTC()

	gen, err := NewGenerator(driver)
	if err != nil {
		return "", err
	}
	sql, err := gen.Model(&model.Todo{}).Where("id = ?", input.ID).Updates(values)
	if err != nil {
		return "", err
	}
	if _, err := db.Exec(ctx, sql); err != nil {
		return sql, fmt.Errorf("update demo todo: %w", err)
	}
	return sql, nil
}

// DeleteTodo 通过 sqlgen 生成并执行 Demo Todo 删除语句。
//
// 由于生成器启用了 SoftDelete，这里实际写入 deleted_at，而不是物理删除记录。
func DeleteTodo(ctx context.Context, db database.Database, driver string, id uint) (string, error) {
	if db == nil {
		return "", ErrMissingDatabase
	}
	if id == 0 {
		return "", ErrMissingTodoID
	}

	gen, err := NewGenerator(driver)
	if err != nil {
		return "", err
	}
	sql, err := gen.Delete(&model.Todo{}, id)
	if err != nil {
		return "", err
	}
	if _, err := db.Exec(ctx, sql); err != nil {
		return sql, fmt.Errorf("delete demo todo: %w", err)
	}
	return sql, nil
}

// normalizeLimit 约束 Demo Todo 查询的分页上限，避免命令行入口发起无界读取。
func normalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultTodoLimit
	}
	if limit > MaxTodoLimit {
		return MaxTodoLimit
	}
	return limit
}
