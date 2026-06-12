package initapp

// 本测试文件固定应用分层组装和 Demo schema 策略，防止注释补全和后续重构改变外部可观察行为。

import (
	"path/filepath"
	"testing"

	"github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/logger"
)

// TestDemoSchemaPolicyFor 固定应用分层组装和 Demo schema 策略，确保后续注释补全或结构调整不改变该场景。
func TestDemoSchemaPolicyFor(t *testing.T) {
	tests := []struct {
		name    string
		trigger DemoSchemaTrigger
		apply   bool
	}{
		{name: "server start applies demo schema", trigger: DemoSchemaTriggerServerStart, apply: true},
		{name: "reload skips demo schema apply", trigger: DemoSchemaTriggerReload, apply: false},
		{name: "unknown trigger skips schema apply", trigger: DemoSchemaTrigger("unknown"), apply: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := DemoSchemaPolicyFor(tt.trigger)
			if policy.Trigger != tt.trigger {
				t.Fatalf("trigger = %q, want %q", policy.Trigger, tt.trigger)
			}
			if policy.Apply != tt.apply {
				t.Fatalf("Apply = %v, want %v", policy.Apply, tt.apply)
			}
			if policy.Reason == "" {
				t.Fatal("Reason must describe why the policy exists")
			}
		})
	}
}

// TestApplyDemoSchemaForTrigger 固定应用分层组装和 Demo schema 策略，确保后续注释补全或结构调整不改变该场景。
func TestApplyDemoSchemaForTrigger(t *testing.T) {
	tests := []struct {
		name     string
		trigger  DemoSchemaTrigger
		hasTable bool
	}{
		{name: "server start creates todo table", trigger: DemoSchemaTriggerServerStart, hasTable: true},
		{name: "reload leaves schema untouched", trigger: DemoSchemaTriggerReload, hasTable: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDatabase(t)

			policy, err := ApplyDemoSchemaForTrigger(db, string(database.DriverSQLite), testLogger{}, tt.trigger)
			if err != nil {
				t.Fatalf("ApplyDemoSchemaForTrigger() error = %v", err)
			}
			if policy.Apply != tt.hasTable {
				t.Fatalf("Apply = %v, want %v", policy.Apply, tt.hasTable)
			}
			if got := hasDemoTodoTable(t, db); got != tt.hasTable {
				t.Fatalf("HasTable(Todo) = %v, want %v", got, tt.hasTable)
			}
			if got := hasDemoCustomerTable(t, db); got != tt.hasTable {
				t.Fatalf("HasTable(Customer) = %v, want %v", got, tt.hasTable)
			}
		})
	}
}

// TestApplyDemoSchemaKeepsServerStartDefault 固定应用分层组装和 Demo schema 策略，确保后续注释补全或结构调整不改变该场景。
func TestApplyDemoSchemaKeepsServerStartDefault(t *testing.T) {
	db := newTestDatabase(t)

	if err := ApplyDemoSchema(db, string(database.DriverSQLite), testLogger{}); err != nil {
		t.Fatalf("ApplyDemoSchema() error = %v", err)
	}
	if !hasDemoTodoTable(t, db) {
		t.Fatal("ApplyDemoSchema() should create the demo todo table")
	}
	if !hasDemoCustomerTable(t, db) {
		t.Fatal("ApplyDemoSchema() should create the demo customer table")
	}
}

// TestNewModulesRespectsDemoConfig 固定应用分层组装和 Demo schema 策略，确保后续注释补全或结构调整不改变该场景。
func TestNewModulesRespectsDemoConfig(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		applySchema bool
		wantHandler bool
		wantTable   bool
	}{
		{name: "demo disabled", enabled: false, applySchema: true, wantHandler: false, wantTable: false},
		{name: "demo enabled without schema apply", enabled: true, applySchema: false, wantHandler: true, wantTable: false},
		{name: "demo enabled with schema apply", enabled: true, applySchema: true, wantHandler: true, wantTable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDatabase(t)
			enabled := tt.enabled
			applySchema := tt.applySchema
			core := Core{
				Config: &config.Config{
					Database: config.DatabaseConfig{Driver: string(database.DriverSQLite)},
					Demo: config.DemoConfig{
						Enabled:            &enabled,
						ApplySchemaOnStart: &applySchema,
					},
				},
				Logger: testLogger{},
			}

			modules, err := NewModules(core, Infrastructure{Database: db})
			if err != nil {
				t.Fatalf("NewModules() error = %v", err)
			}
			if got := modules.Demo.TodoHandler != nil; got != tt.wantHandler {
				t.Fatalf("TodoHandler present = %v, want %v", got, tt.wantHandler)
			}
			if got := modules.Demo.CustomerHandler != nil; got != tt.wantHandler {
				t.Fatalf("CustomerHandler present = %v, want %v", got, tt.wantHandler)
			}
			if got := hasDemoTodoTable(t, db); got != tt.wantTable {
				t.Fatalf("HasTable(Todo) = %v, want %v", got, tt.wantTable)
			}
			if got := hasDemoCustomerTable(t, db); got != tt.wantTable {
				t.Fatalf("HasTable(Customer) = %v, want %v", got, tt.wantTable)
			}
		})
	}
}

// TestApplyDemoSchemaForTriggerAllowsMissingDatabase 固定应用分层组装和 Demo schema 策略，确保后续注释补全或结构调整不改变该场景。
func TestApplyDemoSchemaForTriggerAllowsMissingDatabase(t *testing.T) {
	policy, err := ApplyDemoSchemaForTrigger(nil, string(database.DriverSQLite), testLogger{}, DemoSchemaTriggerServerStart)
	if err != nil {
		t.Fatalf("ApplyDemoSchemaForTrigger(nil) error = %v", err)
	}
	if !policy.Apply {
		t.Fatal("server-start policy should still describe a schema apply trigger")
	}
}

// newTestDatabase 构造当前测试场景所需的最小依赖集合，避免测试直接耦合生产装配流程。
func newTestDatabase(t *testing.T) database.Database {
	t.Helper()

	db, err := database.New(&database.Config{
		Driver: database.DriverSQLite,
		DBName: filepath.Join(t.TempDir(), "demo.db"),
	})
	if err != nil {
		t.Fatalf("open sqlite memory database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite memory database: %v", err)
		}
	})
	return db
}

func hasDemoTodoTable(t *testing.T, db database.Database) bool {
	t.Helper()
	hasTable, err := db.HasTable(nil, &model.Todo{})
	if err != nil {
		t.Fatalf("HasTable(Todo) error = %v", err)
	}
	return hasTable
}

func hasDemoCustomerTable(t *testing.T, db database.Database) bool {
	t.Helper()
	hasTable, err := db.HasTable(nil, &model.Customer{})
	if err != nil {
		t.Fatalf("HasTable(Customer) error = %v", err)
	}
	return hasTable
}

type testLogger struct{}

// Debug 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (testLogger) Debug(string, ...interface{}) {}

// Info 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (testLogger) Info(string, ...interface{}) {}

// Warn 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (testLogger) Warn(string, ...interface{}) {}

// Error 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (testLogger) Error(string, ...interface{}) {}

// Fatal 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (testLogger) Fatal(string, ...interface{}) {}

// With 实现测试日志桩的字段绑定入口，返回自身以保持 logger.Logger 链式调用契约。
func (l testLogger) With(...interface{}) logger.Logger {
	return l
}

// Sync 实现测试日志桩的刷新入口，测试环境不持有真实缓冲区。
func (testLogger) Sync() error {
	return nil
}

// Reload 实现测试桩的配置重载入口，用于验证调用路径而不触发真实资源替换。
func (testLogger) Reload(*logger.Config) error {
	return nil
}
