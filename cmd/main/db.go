package main

// 本文件定义 db 子命令，承接 SQL 预览、Demo 表结构初始化和 Todo 示例数据操作的命令行入口。

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	"github.com/rei0721/go-scaffold/internal/app/initapp"
	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/pkg/migrator"
	"github.com/rei0721/go-scaffold/types/constants"
)

const (
	dbCommandName       = "db"
	dbOperationDatabase = "database"
	dbOperationSchema   = "schema"
	dbOperationCreate   = "todo-create"
	dbOperationList     = "todo-list"
	dbOperationGet      = "todo-get"
	dbOperationUpdate   = "todo-update"
	dbOperationDelete   = "todo-delete"
	defaultDBOperation  = dbOperationSchema
	dbCompletedFlagName = "completed"
)

// DBCommand exposes database DDL, demo schema, and demo Todo CRUD operations
// through sqlgen-backed application services.
type DBCommand struct {
	runner func(*cli.Context, dbOptions) error
}

// dbOptions is the parsed flag state for one db command invocation.
type dbOptions struct {
	ConfigPath  string
	Operation   string
	Apply       bool
	ID          int
	Title       string
	Description string
	Completed   string
	Limit       int
	Offset      int
	PrintSQL    bool
}

// NewDBCommand builds the cmd/main db command. Keep command wiring here and
// SQL generation/execution in internal/app/dbapp so the CLI remains thin.
func NewDBCommand() *DBCommand {
	return &DBCommand{}
}

// Name 返回 db 子命令在 CLI 注册表中的唯一名称。
func (c *DBCommand) Name() string {
	return dbCommandName
}

// Description 返回 db 子命令的短描述，提示该命令面向 SQL 预览和 Demo 数据操作。
func (c *DBCommand) Description() string {
	return "Run sqlgen-powered schema, migrations, and demo CRUD operations"
}

// Usage 返回 db 子命令的调用格式，强调 operation 参数决定底层执行路径。
func (c *DBCommand) Usage() string {
	return fmt.Sprintf("%s [--operation=<database|schema|todo-create|todo-list|todo-get|todo-update|todo-delete>] [flags]", dbCommandName)
}

// Flags 声明 db 子命令支持的配置、执行开关、Todo 字段和分页参数。
func (c *DBCommand) Flags() []cli.FlagSpec {
	return []cli.FlagSpec{
		{
			Name:        "config",
			Type:        cli.FlagTypeString,
			Default:     constants.AppDefaultConfigPath,
			Description: "Config file path",
			EnvVar:      appconfig.EnvConfigPathName(),
		},
		{
			Name:        "operation",
			Type:        cli.FlagTypeString,
			Default:     defaultDBOperation,
			Description: "Database operation",
		},
		{
			Name:        "apply",
			Type:        cli.FlagTypeBool,
			Default:     false,
			Description: "Apply generated schema SQL instead of printing it",
		},
		{
			Name:        "id",
			Type:        cli.FlagTypeInt,
			Default:     0,
			Description: "Demo todo ID",
		},
		{
			Name:        "title",
			Type:        cli.FlagTypeString,
			Default:     "",
			Description: "Demo todo title",
		},
		{
			Name:        "description",
			Type:        cli.FlagTypeString,
			Default:     "",
			Description: "Demo todo description",
		},
		{
			Name:        dbCompletedFlagName,
			Type:        cli.FlagTypeString,
			Default:     "",
			Description: "Demo todo completed value: true or false",
		},
		{
			Name:        "limit",
			Type:        cli.FlagTypeInt,
			Default:     dbapp.DefaultTodoLimit,
			Description: "Demo todo list limit",
		},
		{
			Name:        "offset",
			Type:        cli.FlagTypeInt,
			Default:     0,
			Description: "Demo todo list offset",
		},
		{
			Name:        "print-sql",
			Type:        cli.FlagTypeBool,
			Default:     false,
			Description: "Print generated SQL after executing the operation",
		},
	}
}

// Spec 将 db 命令转换为 pkg/cli 的新声明式命令规格。
func (c *DBCommand) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        c.Name(),
		Use:         c.Usage(),
		Description: c.Description(),
		Flags:       c.Flags(),
		Run:         c.Execute,
		Commands: []cli.CommandSpec{
			c.MigrateSpec(),
		},
	}
}

func (c *DBCommand) MigrateSpec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        "migrate",
		Use:         "migrate <up|down|status>",
		Description: "Run database migrations",
		Flags: []cli.FlagSpec{
			{
				Name:        "config",
				Type:        cli.FlagTypeString,
				Default:     constants.AppDefaultConfigPath,
				Description: "Config file path",
				EnvVar:      appconfig.EnvConfigPathName(),
			},
		},
		Args: func(ctx *cli.Context) error {
			if len(ctx.Args) != 1 {
				return &cli.UsageError{Command: ctx.CommandPath, Message: "expected migrate operation: up, down, or status"}
			}
			switch ctx.Args[0] {
			case "up", "down", "status":
				return nil
			default:
				return &cli.UsageError{Command: ctx.CommandPath, Message: "unsupported migrate operation: " + ctx.Args[0]}
			}
		},
		Run: c.ExecuteMigrate,
	}
}

// Execute 将 CLI flag 归一化为 DBCommand 状态，并把实际数据库流程委托给可替换 runner。
func (c *DBCommand) Execute(ctx *cli.Context) error {
	opts := dbOptions{
		ConfigPath:  ctx.GetString("config"),
		Operation:   ctx.GetString("operation"),
		Apply:       ctx.GetBool("apply"),
		ID:          ctx.GetInt("id"),
		Title:       ctx.GetString("title"),
		Description: ctx.GetString("description"),
		Completed:   ctx.GetString(dbCompletedFlagName),
		Limit:       ctx.GetInt("limit"),
		Offset:      ctx.GetInt("offset"),
		PrintSQL:    ctx.GetBool("print-sql"),
	}
	if opts.Operation == "" {
		opts.Operation = defaultDBOperation
	}
	if c.runner != nil {
		return c.runner(ctx, opts)
	}
	return runDBCommand(ctx, opts)
}

// runDBCommand 根据 operation 分派 SQL 预览或数据库执行路径，所有外部错误都会返回给 CLI 层统一打印。
func runDBCommand(ctx *cli.Context, opts dbOptions) error {
	core, err := initapp.NewCore(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("initialize core: %w", err)
	}
	defer func() {
		if core.Logger != nil {
			_ = core.Logger.Sync()
		}
	}()

	// Read-only DDL preview must not open a database connection.
	if (opts.Operation == dbOperationDatabase || opts.Operation == dbOperationSchema) && !opts.Apply {
		sql, err := dbSQLForPrint(opts, core.Config.Database)
		if err != nil {
			return err
		}
		fmt.Fprintln(ctx.Stdout, sql)
		return nil
	}

	db, err := initapp.NewDatabase(core.Config)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	switch opts.Operation {
	case dbOperationDatabase:
		sql, err := dbapp.ApplyDatabase(ctxWithFallback(ctx), db, core.Config.Database.Driver, core.Config.Database.DBName)
		if err != nil {
			return err
		}
		writeOperationResult(ctx.Stdout, "database create applied", sql, opts.PrintSQL)
	case dbOperationSchema:
		sql, err := dbapp.ApplyDemoSchema(ctxWithFallback(ctx), db, core.Config.Database.Driver)
		if err != nil {
			return err
		}
		writeOperationResult(ctx.Stdout, "schema applied", sql, opts.PrintSQL)
	case dbOperationCreate:
		completed, err := parseCompleted(opts.Completed, false)
		if err != nil {
			return err
		}
		sql, err := dbapp.CreateTodo(ctxWithFallback(ctx), db, core.Config.Database.Driver, dbapp.TodoCreateInput{
			Title:       opts.Title,
			Description: opts.Description,
			Completed:   completed,
		})
		if err != nil {
			return err
		}
		writeOperationResult(ctx.Stdout, "todo created", sql, opts.PrintSQL)
	case dbOperationList:
		todos, sql, err := dbapp.ListTodos(ctxWithFallback(ctx), db, core.Config.Database.Driver, opts.Limit, opts.Offset)
		if err != nil {
			return err
		}
		if opts.PrintSQL {
			fmt.Fprintln(ctx.Stdout, sql)
		}
		return json.NewEncoder(ctx.Stdout).Encode(todos)
	case dbOperationGet:
		todo, sql, err := dbapp.GetTodo(ctxWithFallback(ctx), db, core.Config.Database.Driver, uint(opts.ID))
		if err != nil {
			return err
		}
		if opts.PrintSQL {
			fmt.Fprintln(ctx.Stdout, sql)
		}
		return json.NewEncoder(ctx.Stdout).Encode(todo)
	case dbOperationUpdate:
		update, err := buildTodoUpdate(opts)
		if err != nil {
			return err
		}
		sql, err := dbapp.UpdateTodo(ctxWithFallback(ctx), db, core.Config.Database.Driver, update)
		if err != nil {
			return err
		}
		writeOperationResult(ctx.Stdout, "todo updated", sql, opts.PrintSQL)
	case dbOperationDelete:
		sql, err := dbapp.DeleteTodo(ctxWithFallback(ctx), db, core.Config.Database.Driver, uint(opts.ID))
		if err != nil {
			return err
		}
		writeOperationResult(ctx.Stdout, "todo deleted", sql, opts.PrintSQL)
	default:
		return fmt.Errorf("unsupported db operation: %s", opts.Operation)
	}

	return nil
}

func (c *DBCommand) ExecuteMigrate(ctx *cli.Context) error {
	core, err := initapp.NewCore(ctx.GetString("config"))
	if err != nil {
		return fmt.Errorf("initialize core: %w", err)
	}
	defer func() {
		if core.Logger != nil {
			_ = core.Logger.Sync()
		}
	}()
	db, err := initapp.NewDatabase(core.Config)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()
	core.Config.Migration.ApplyDefaults()
	runner, err := migrator.New(db, migrator.Config{Driver: string(core.Config.Database.Driver), Dir: core.Config.Migration.Dir})
	if err != nil {
		return err
	}
	switch ctx.Args[0] {
	case "up":
		if err := runner.Up(ctx.Context); err != nil {
			return err
		}
		fmt.Fprintln(ctx.Stdout, "migrations applied")
	case "down":
		if err := runner.Down(ctx.Context); err != nil {
			return err
		}
		fmt.Fprintln(ctx.Stdout, "migration rolled back")
	case "status":
		return runner.Status(ctx.Context, ctx.Stdout)
	}
	return nil
}

// dbSQLForPrint centralizes operations that can render generated SQL without
// side effects. New printable operations should be added here and backed by
// sqlgen APIs in internal/app/dbapp.
func dbSQLForPrint(opts dbOptions, cfg appconfig.DatabaseConfig) (string, error) {
	switch opts.Operation {
	case dbOperationDatabase:
		return dbapp.DatabaseSQL(cfg.Driver, cfg.DBName)
	case dbOperationSchema:
		return dbapp.DemoSchemaSQL(cfg.Driver)
	default:
		return "", fmt.Errorf("unsupported db operation: %s", opts.Operation)
	}
}

// ctxWithFallback 为数据库命令提供上下文入口；测试桩未填充时回退到后台上下文。
func ctxWithFallback(ctx *cli.Context) context.Context {
	if ctx != nil && ctx.Context != nil {
		return ctx.Context
	}
	return context.Background()
}

// writeOperationResult 将 db 子命令执行结果写入命令输出流，保持业务执行与展示格式解耦。
func writeOperationResult(w io.Writer, message, sql string, printSQL bool) {
	fmt.Fprintln(w, message)
	if printSQL {
		fmt.Fprintln(w, sql)
	}
}

// parseCompleted 解析可选布尔字符串，空值表示调用方未请求更新 completed 字段。
func parseCompleted(value string, defaultValue bool) (bool, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid --%s value: %w", dbCompletedFlagName, err)
	}
	return parsed, nil
}

// buildTodoUpdate 把 CLI flag 转换为服务层 UpdateTodoInput，利用指针字段保留未传入与零值的区别。
func buildTodoUpdate(opts dbOptions) (dbapp.TodoUpdateInput, error) {
	update := dbapp.TodoUpdateInput{ID: uint(opts.ID)}
	if opts.Title != "" {
		update.Title = &opts.Title
	}
	if opts.Description != "" {
		update.Description = &opts.Description
	}
	if opts.Completed != "" {
		completed, err := parseCompleted(opts.Completed, false)
		if err != nil {
			return update, err
		}
		update.Completed = &completed
	}
	return update, nil
}
