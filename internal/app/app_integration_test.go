package app

// 本测试文件固定应用组装根的最小可启动契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
)

// TestNewServerModeBuildsMinimalApplication 固定应用组装根的最小可启动契约，确保后续注释补全或结构调整不改变该场景。
func TestNewServerModeBuildsMinimalApplication(t *testing.T) {
	clearAppIntegrationEnv(t)

	configPath := writeAppIntegrationConfig(t, filepath.Join(t.TempDir(), "server-mode.db"))

	application, err := New(Options{ConfigPath: configPath, Mode: ModeServer})
	if err != nil {
		t.Fatalf("new server app: %v", err)
	}
	defer shutdownApp(t, application)

	if application.Core.Config == nil {
		t.Fatal("expected core config")
	}
	if application.Core.ConfigManager == nil {
		t.Fatal("expected core config manager")
	}
	if application.Core.Logger == nil {
		t.Fatal("expected core logger")
	}
	if application.Core.I18n == nil {
		t.Fatal("expected core i18n")
	}
	if application.Core.I18nUtils == nil {
		t.Fatal("expected core i18n utils")
	}
	if application.Core.IDGenerator == nil {
		t.Fatal("expected core id generator")
	}

	if application.Infra.Database == nil {
		t.Fatal("expected database infrastructure")
	}
	if application.Infra.Cache != nil {
		t.Fatal("expected redis cache to be disabled")
	}
	if application.Infra.Executor != nil {
		t.Fatal("expected executor to be disabled")
	}
	if application.Infra.Storage != nil {
		t.Fatal("expected storage to be disabled")
	}

	if application.Modules.Demo.TodoRepository == nil {
		t.Fatal("expected demo repository")
	}
	if application.Modules.Demo.TodoService == nil {
		t.Fatal("expected demo service")
	}
	if application.Modules.Demo.TodoHandler == nil {
		t.Fatal("expected demo handler")
	}
	if application.Modules.Demo.CustomerHandler == nil {
		t.Fatal("expected demo customer handler")
	}

	if application.Transport.Router == nil {
		t.Fatal("expected HTTP router")
	}
	if application.Transport.HTTPServer == nil {
		t.Fatal("expected HTTP server wrapper")
	}
	if application.Transport.RPCServer == nil {
		t.Fatal("expected RPC server wrapper")
	}

	hasTable, err := application.Infra.Database.HasTable(context.Background(), &model.Todo{})
	if err != nil {
		t.Fatalf("check demo todo schema: %v", err)
	}
	if !hasTable {
		t.Fatal("expected demo todo schema to be created in server mode")
	}
	hasTable, err = application.Infra.Database.HasTable(context.Background(), &model.Customer{})
	if err != nil {
		t.Fatalf("check demo customer schema: %v", err)
	}
	if !hasTable {
		t.Fatal("expected demo customer schema to be created in server mode")
	}

	if err := application.Core.ConfigManager.Update(func(cfg *config.Config) {
		cfg.Server.Port = 18082
		cfg.Server.ReadTimeout = 2
	}); err != nil {
		t.Fatalf("update config through manager: %v", err)
	}
	if got := application.Core.Config.Server.Port; got != 18082 {
		t.Fatalf("expected app hook to update core config port to 18082, got %d", got)
	}
	if got := application.Core.ConfigManager.Get().Server.Port; got != 18082 {
		t.Fatalf("expected manager config port to be 18082, got %d", got)
	}
}

// writeAppIntegrationConfig 写入测试夹具文件，并把文件系统准备细节限制在测试辅助层。
func writeAppIntegrationConfig(t *testing.T, dbPath string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	messagesDir := filepath.Join(repoRoot, "configs", "locales")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	storagePath := filepath.Join(t.TempDir(), "storage")

	content := fmt.Sprintf(`server:
  host: "127.0.0.1"
  port: 18081
  mode: "test"
  read_timeout: 1
  write_timeout: 1
  idle_timeout: 1
database:
  driver: "sqlite"
  host: ""
  port: 0
  user: ""
  password: ""
  dbname: %s
  max_open_conns: 1
  max_idle_conns: 1
redis:
  enabled: false
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 1
  min_idle_conns: 0
  max_retries: 0
  dial_timeout: 1
  read_timeout: 1
  write_timeout: 1
logger:
  level: "error"
  format: "console"
  output: "stdout"
  file_path: ""
  max_size: 1
  max_backups: 1
  max_age: 1
i18n:
  default: "zh-CN"
  supported:
    - "zh-CN"
    - "en-US"
  messages_dir: %s
executor:
  enabled: false
  pools: []
storage:
  enabled: false
  fs_type: "memory"
  base_path: %s
  enable_watch: false
  watch_buffer_size: 1
cors:
  enabled: true
  allow_origins:
    - "*"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "PATCH"
    - "OPTIONS"
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "X-Request-ID"
  expose_headers:
    - "X-Request-ID"
  allow_credentials: false
  max_age: 60
rpc:
  enabled: false
  host: "127.0.0.1"
  port: 10099
  read_timeout: 1
  write_timeout: 1
  idle_timeout: 1
`, yamlString(dbPath), yamlString(messagesDir), yamlString(storagePath))

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	return configPath
}

// shutdownApp 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func shutdownApp(t *testing.T, application *App) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := application.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown app: %v", err)
	}
}

// yamlString 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func yamlString(value string) string {
	return strconv.Quote(filepath.ToSlash(value))
}

// clearAppIntegrationEnv 清理测试期间设置的环境变量或全局状态，避免用例之间互相污染。
func clearAppIntegrationEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"DB_DRIVER",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_MAX_OPEN_CONNS",
		"DB_MAX_IDLE_CONNS",
		"REI_APP_DB_DRIVER",
		"REI_APP_DB_HOST",
		"REI_APP_DB_PORT",
		"REI_APP_DB_USER",
		"REI_APP_DB_PASSWORD",
		"REI_APP_DB_NAME",
		"REI_APP_DB_MAX_OPEN_CONNS",
		"REI_APP_DB_MAX_IDLE_CONNS",
		"REDIS_ENABLED",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"REDIS_POOL_SIZE",
		"REDIS_MIN_IDLE_CONNS",
		"REDIS_MAX_RETRIES",
		"REDIS_DIAL_TIMEOUT",
		"REDIS_READ_TIMEOUT",
		"REDIS_WRITE_TIMEOUT",
		"SERVER_PORT",
		"SERVER_MODE",
		"SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"LOG_OUTPUT",
		"I18N_DEFAULT",
		"I18N_SUPPORTED",
		"STORAGE_ENABLED",
		"STORAGE_FS_TYPE",
		"STORAGE_BASE_PATH",
		"STORAGE_ENABLE_WATCH",
		"STORAGE_WATCH_BUFFER_SIZE",
		"CORS_ENABLED",
		"CORS_ALLOW_ORIGINS",
		"CORS_ALLOW_METHODS",
		"CORS_ALLOW_HEADERS",
		"CORS_EXPOSE_HEADERS",
		"CORS_ALLOW_CREDENTIALS",
		"CORS_MAX_AGE",
		"RPC_ENABLED",
		"RPC_HOST",
		"RPC_PORT",
		"RPC_READ_TIMEOUT",
		"RPC_WRITE_TIMEOUT",
		"RPC_IDLE_TIMEOUT",
	}
	for _, key := range keys {
		t.Setenv(key, "")
		t.Setenv(config.EnvPrefixJoin(key), "")
	}
}
