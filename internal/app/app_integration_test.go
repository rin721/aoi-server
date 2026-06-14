package app

// 本测试文件固定应用组装根的最小可启动契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	systemmodel "github.com/rei0721/go-scaffold/internal/modules/system/model"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
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

	snapshot, err := application.Modules.System.Service.ListConfig(context.Background())
	if err != nil {
		t.Fatalf("list system config: %v", err)
	}
	value, ok := systemConfigValue(snapshot, "server.port")
	if !ok {
		t.Fatalf("expected system config to include server.port, got %#v", snapshot)
	}
	if value != 18082 {
		t.Fatalf("expected system config port to follow manager update, got %#v", value)
	}
	corsOriginsItem, ok := systemConfigItem(snapshot, "cors.allow_origins")
	if !ok || !corsOriginsItem.Editable || corsOriginsItem.ValueType != systemmodel.ConfigValueTypeArray {
		t.Fatalf("expected cors.allow_origins to be editable string array, got %#v ok=%v", corsOriginsItem, ok)
	}
	executorPoolSizeItem, ok := systemConfigItem(snapshot, "executor.pools.0.size")
	if !ok || !executorPoolSizeItem.Editable || executorPoolSizeItem.ValueType != systemmodel.ConfigValueTypeNumber {
		t.Fatalf("expected executor.pools.0.size to be editable number, got %#v ok=%v", executorPoolSizeItem, ok)
	}

	updated, err := application.Modules.System.Service.UpdateConfig(context.Background(), systemservice.UpdateConfigInput{
		Items: []systemservice.UpdateConfigItem{
			{Key: "server.port", Value: 18083},
			{Key: "database.password", Value: "runtime-secret"},
		},
	})
	if err != nil {
		t.Fatalf("update system config through service: %v", err)
	}
	if got := application.Core.ConfigManager.Get().Server.Port; got != 18083 {
		t.Fatalf("expected manager config port to be 18083, got %d", got)
	}
	value, ok = systemConfigValue(updated, "server.port")
	if !ok || value != 18083 {
		t.Fatalf("expected updated snapshot port 18083, got %#v ok=%v", value, ok)
	}
	value, ok = systemConfigValue(updated, "database.password")
	if !ok || value != "已配置" {
		t.Fatalf("expected database.password to remain redacted, got %#v ok=%v", value, ok)
	}
	if got := application.Core.ConfigManager.Get().Database.Password; got != "runtime-secret" {
		t.Fatalf("expected manager database password to be updated, got %q", got)
	}
	fileManager := config.NewManager()
	if err := fileManager.Load(configPath); err != nil {
		t.Fatalf("reload config file after runtime update: %v", err)
	}
	if got := fileManager.Get().Server.Port; got != 18081 {
		t.Fatalf("expected non-persist update to leave config file port 18081, got %d", got)
	}

	unchangedSecret, err := application.Modules.System.Service.UpdateConfig(context.Background(), systemservice.UpdateConfigInput{
		Items: []systemservice.UpdateConfigItem{
			{Key: "database.password", Value: ""},
		},
	})
	if err != nil {
		t.Fatalf("update empty secret through service: %v", err)
	}
	if got := application.Core.ConfigManager.Get().Database.Password; got != "runtime-secret" {
		t.Fatalf("expected empty secret value to leave current password unchanged, got %q", got)
	}
	value, ok = systemConfigValue(unchangedSecret, "database.password")
	if !ok || value != "已配置" {
		t.Fatalf("expected unchanged database.password to remain redacted, got %#v ok=%v", value, ok)
	}

	persisted, err := application.Modules.System.Service.UpdateConfig(context.Background(), systemservice.UpdateConfigInput{
		Items: []systemservice.UpdateConfigItem{
			{Key: "server.port", Value: 18084},
			{Key: "database.password", Value: "persistent-secret"},
			{Key: "cors.allow_origins", Value: []any{"https://admin.example.com", "https://app.example.com"}},
			{Key: "executor.pools.0.size", Value: 24},
		},
		Persist: true,
	})
	if err != nil {
		t.Fatalf("persist system config through service: %v", err)
	}
	value, ok = systemConfigValue(persisted, "server.port")
	if !ok || value != 18084 {
		t.Fatalf("expected persisted snapshot port 18084, got %#v ok=%v", value, ok)
	}
	value, ok = systemConfigValue(persisted, "database.password")
	if !ok || value != "已配置" {
		t.Fatalf("expected persisted database.password to remain redacted, got %#v ok=%v", value, ok)
	}
	value, ok = systemConfigValue(persisted, "cors.allow_origins")
	if !ok || !reflect.DeepEqual(value, []string{"https://admin.example.com", "https://app.example.com"}) {
		t.Fatalf("expected persisted CORS origins in snapshot, got %#v ok=%v", value, ok)
	}
	value, ok = systemConfigValue(persisted, "executor.pools.0.size")
	if !ok || value != 24 {
		t.Fatalf("expected persisted executor pool size in snapshot, got %#v ok=%v", value, ok)
	}
	persistedContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read persisted config file: %v", err)
	}
	if text := string(persistedContent); !strings.Contains(text, "port: 18084") || !strings.Contains(text, `password: "persistent-secret"`) ||
		!strings.Contains(text, `- "https://admin.example.com"`) || !strings.Contains(text, `size: 24`) {
		t.Fatalf("expected persisted config file to include updated port and password, got:\n%s", text)
	}
	fileManager = config.NewManager()
	if err := fileManager.Load(configPath); err != nil {
		t.Fatalf("reload persisted config file: %v", err)
	}
	if got := fileManager.Get().Server.Port; got != 18084 {
		t.Fatalf("expected config file port 18084, got %d", got)
	}
	if got := fileManager.Get().Database.Password; got != "persistent-secret" {
		t.Fatalf("expected config file password to be persistent-secret, got %q", got)
	}
	if !reflect.DeepEqual(fileManager.Get().CORS.AllowOrigins, []string{"https://admin.example.com", "https://app.example.com"}) {
		t.Fatalf("expected config file CORS origins to persist, got %#v", fileManager.Get().CORS.AllowOrigins)
	}
	if got := fileManager.Get().Executor.Pools[0].Size; got != 24 {
		t.Fatalf("expected config file executor pool size 24, got %d", got)
	}
}

func TestWebInitialSetupRunsSharedInitialization(t *testing.T) {
	clearAppIntegrationEnv(t)
	clearInitialSetupEnv(t)

	configPath := writeInitialSetupIntegrationConfig(t, filepath.Join(t.TempDir(), "web-setup.db"))
	application, err := New(Options{ConfigPath: configPath, Mode: ModeServer})
	if err != nil {
		t.Fatalf("new server app: %v", err)
	}
	defer shutdownApp(t, application)

	statusRecorder := performAppJSONRequest(application, http.MethodGet, "/api/v1/auth/setup/status", "", "")
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("setup status HTTP status = %d, body=%s", statusRecorder.Code, statusRecorder.Body.String())
	}
	var statusBody appAPIResponse[struct {
		Required bool `json:"required"`
	}]
	if err := json.Unmarshal(statusRecorder.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode setup status: %v", err)
	}
	if !statusBody.Data.Required {
		t.Fatalf("setup status required = false, want true: %#v", statusBody)
	}

	setupBody := `{"orgCode":"acme","orgName":"Acme Corp","username":"admin","displayName":"Admin","email":"admin@example.com","password":"password123"}`
	setupRecorder := performAppJSONRequest(application, http.MethodPost, "/api/v1/auth/setup/initial-admin", setupBody, "")
	if setupRecorder.Code != http.StatusOK {
		t.Fatalf("initial setup HTTP status = %d, body=%s", setupRecorder.Code, setupRecorder.Body.String())
	}
	var tokenBody appAPIResponse[iamservice.TokenPair]
	if err := json.Unmarshal(setupRecorder.Body.Bytes(), &tokenBody); err != nil {
		t.Fatalf("decode setup token response: %v", err)
	}
	if tokenBody.Data.AccessToken == "" || tokenBody.Data.RefreshToken == "" {
		t.Fatalf("setup did not issue tokens: %#v", tokenBody.Data)
	}

	principal, err := application.Modules.IAM.Service.AuthenticateToken(context.Background(), tokenBody.Data.AccessToken)
	if err != nil {
		t.Fatalf("authenticate setup token: %v", err)
	}
	for _, permission := range []struct {
		obj string
		act string
	}{
		{obj: "audit", act: "read"},
		{obj: "session", act: "read"},
		{obj: "server", act: "read"},
	} {
		allowed, err := application.Modules.IAM.Service.Authorize(context.Background(), principal, permission.obj, permission.act)
		if err != nil || !allowed {
			t.Fatalf("owner permission %s:%s allowed=%v err=%v", permission.obj, permission.act, allowed, err)
		}
	}

	for _, endpoint := range []string{
		fmt.Sprintf("/api/v1/orgs/%d/sessions?pageSize=6", principal.OrgID),
		fmt.Sprintf("/api/v1/orgs/%d/audit-logs?limit=6", principal.OrgID),
		"/api/v1/system/menus",
	} {
		recorder := performAppJSONRequest(application, http.MethodGet, endpoint, "", tokenBody.Data.AccessToken)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s HTTP status = %d, body=%s", endpoint, recorder.Code, recorder.Body.String())
		}
	}
}

func systemConfigValue(snapshot systemmodel.ConfigSnapshot, key string) (any, bool) {
	item, ok := systemConfigItem(snapshot, key)
	if !ok {
		return nil, false
	}
	return item.Value, true
}

func systemConfigItem(snapshot systemmodel.ConfigSnapshot, key string) (systemmodel.ConfigItem, bool) {
	for _, section := range snapshot.Sections {
		for _, item := range section.Items {
			if item.Key == key {
				return item, true
			}
		}
	}
	return systemmodel.ConfigItem{}, false
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
  pools:
    - name: "default"
      size: 8
      expiry: 30
      non_blocking: true
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

func writeInitialSetupIntegrationConfig(t *testing.T, dbPath string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	messagesDir := filepath.Join(repoRoot, "configs", "locales")
	migrationsDir := filepath.Join(repoRoot, "internal", "migrations")
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	content := fmt.Sprintf(`server:
  host: "127.0.0.1"
  port: 18081
  mode: "test"
  read_timeout: 1
  write_timeout: 1
  idle_timeout: 1
webui:
  enabled: false
database:
  driver: "sqlite"
  dbname: %s
  max_open_conns: 1
  max_idle_conns: 1
redis:
  enabled: false
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
storage:
  enabled: false
  fs_type: "memory"
cors:
  enabled: true
  allow_origins:
    - "*"
  allow_methods:
    - "GET"
    - "POST"
    - "PATCH"
    - "OPTIONS"
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Authorization"
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
plugins:
  enabled: false
demo:
  enabled: false
system:
  seed_defaults_on_start: false
auth:
  enabled: true
  self_signup_enabled: false
  issuer: "go-scaffold"
  audience:
    - "go-scaffold-api"
  signing_key: "12345678901234567890123456789012"
  refresh_token_pepper: "pepper"
  mfa_secret_key: "12345678901234567890123456789012"
  notification_driver: "debug"
  password_policy:
    min_length: 8
migration:
  auto_apply: false
  dir: %s
`, yamlString(dbPath), yamlString(messagesDir), yamlString(migrationsDir))

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("write setup test config: %v", err)
	}
	return configPath
}

type appAPIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func performAppJSONRequest(application *App, method string, path string, body string, accessToken string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	request := httptest.NewRequest(method, path, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := httptest.NewRecorder()
	application.Transport.Router.ServeHTTP(recorder, request)
	return recorder
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

func clearInitialSetupEnv(t *testing.T) {
	t.Helper()

	for _, path := range []string{
		"auth.enabled",
		"auth.signing_key",
		"auth.refresh_token_pepper",
		"auth.mfa_secret_key",
		"auth.notification_driver",
		"database.driver",
		"database.dbname",
		"migration.auto_apply",
		"migration.dir",
	} {
		for _, key := range config.EnvNamesForPath(path) {
			unsetAppIntegrationEnvForTest(t, key)
		}
	}
}

func unsetAppIntegrationEnvForTest(t *testing.T, key string) {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(key, oldValue); err != nil {
				t.Errorf("restore %s: %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("restore unset %s: %v", key, err)
		}
	})
}
