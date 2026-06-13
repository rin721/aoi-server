package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/pkg/configloader"
)

func TestDocumentedExampleConfigYAMLFilesAreValid(t *testing.T) {
	root := configExamplesRepoRoot(t)
	files := documentedExampleConfigFiles(t, root)

	for _, file := range files {
		file := file
		t.Run(filepath.ToSlash(strings.TrimPrefix(file, root+string(os.PathSeparator))), func(t *testing.T) {
			loader := configloader.New()
			loader.SetConfigFile(file)
			if err := loader.ReadInConfig(); err != nil {
				t.Fatalf("parse example config: %v", err)
			}
			if len(loader.AllSettings()) == 0 {
				t.Fatal("example config must contain a mapping document")
			}
		})
	}
}

func TestDocumentedExampleConfigsLoadWithControlledEnvironment(t *testing.T) {
	root := configExamplesRepoRoot(t)
	files := documentedExampleConfigFiles(t, root)

	for _, file := range files {
		file := file
		t.Run(filepath.ToSlash(strings.TrimPrefix(file, root+string(os.PathSeparator))), func(t *testing.T) {
			setControlledExampleEnv(t, filepath.Base(file))
			mgr := NewManager()
			if err := mgr.Load(file); err != nil {
				t.Fatalf("load example config: %v", err)
			}
			if cfg := mgr.Get(); cfg == nil {
				t.Fatal("loaded config is nil")
			}
		})
	}
}

func documentedExampleConfigFiles(t *testing.T, root string) []string {
	t.Helper()

	files := []string{
		filepath.Join(root, "configs", "config.example.yaml"),
		filepath.Join(root, "deploy", "config.production.example.yaml"),
	}
	matches, err := filepath.Glob(filepath.Join(root, "configs", "examples", "*.example.yaml"))
	if err != nil {
		t.Fatalf("glob scenario examples: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no scenario example configs found")
	}
	files = append(files, matches...)
	return files
}

func configExamplesRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func setControlledExampleEnv(t *testing.T, fileName string) {
	t.Helper()

	env := map[string]string{
		"RIN_APP_SERVER_MODE":                         "debug",
		"RIN_APP_DB_DRIVER":                           "sqlite",
		"RIN_APP_DB_HOST":                             "127.0.0.1",
		"RIN_APP_DB_PORT":                             "5432",
		"RIN_APP_DB_USER":                             "go_scaffold",
		"RIN_APP_DB_PASSWORD":                         "example-db-password",
		"RIN_APP_DB_NAME":                             "go_scaffold",
		"RIN_APP_DB_MAX_OPEN_CONNS":                   "10",
		"RIN_APP_DB_MAX_IDLE_CONNS":                   "5",
		"RIN_APP_REDIS_ENABLED":                       "false",
		"RIN_APP_REDIS_HOST":                          "127.0.0.1",
		"RIN_APP_REDIS_PORT":                          "6379",
		"RIN_APP_REDIS_DB":                            "0",
		"RIN_APP_REDIS_POOL_SIZE":                     "10",
		"RIN_APP_REDIS_MIN_IDLE_CONNS":                "2",
		"RIN_APP_REDIS_MAX_RETRIES":                   "2",
		"RIN_APP_REDIS_DIAL_TIMEOUT":                  "5",
		"RIN_APP_REDIS_READ_TIMEOUT":                  "3",
		"RIN_APP_REDIS_WRITE_TIMEOUT":                 "3",
		"RIN_APP_LOG_LEVEL":                           "info",
		"RIN_APP_LOG_FORMAT":                          "console",
		"RIN_APP_LOG_CONSOLE_FORMAT":                  "console",
		"RIN_APP_LOG_FILE_FORMAT":                     "json",
		"RIN_APP_LOG_OUTPUT":                          "stdout",
		"RIN_APP_LOG_FILE_PATH":                       "./logs/app.log",
		"RIN_APP_LOG_MAX_SIZE":                        "100",
		"RIN_APP_LOG_MAX_BACKUPS":                     "7",
		"RIN_APP_LOG_MAX_AGE":                         "30",
		"RIN_APP_I18N_DEFAULT":                        "zh-CN",
		"RIN_APP_I18N_SUPPORTED":                      "zh-CN,en-US",
		"RIN_APP_I18N_MESSAGES_DIR":                   "./configs/locales",
		"RIN_APP_EXECUTOR_ENABLED":                    "true",
		"RIN_APP_STORAGE_ENABLED":                     "false",
		"RIN_APP_STORAGE_FS_TYPE":                     "basepath",
		"RIN_APP_STORAGE_BASE_PATH":                   "./data",
		"RIN_APP_STORAGE_ENABLE_WATCH":                "false",
		"RIN_APP_STORAGE_WATCH_BUFFER_SIZE":           "100",
		"RIN_APP_DEMO_ENABLED":                        "true",
		"RIN_APP_DEMO_APPLY_SCHEMA_ON_START":          "true",
		"RIN_APP_SYSTEM_SEED_DEFAULTS_ON_START":       "true",
		"RIN_APP_WEBUI_ENABLED":                       "true",
		"RIN_APP_WEBUI_MOUNT_PATH":                    "/admin",
		"RIN_APP_WEBUI_DIST_DIR":                      "./web/admin/.output/public",
		"RIN_APP_WEBUI_PUBLIC_BASE_URL":               "/admin",
		"RIN_APP_RPC_ENABLED":                         "false",
		"RIN_APP_RPC_HOST":                            "127.0.0.1",
		"RIN_APP_RPC_PORT":                            "10099",
		"RIN_APP_RPC_READ_TIMEOUT":                    "10",
		"RIN_APP_RPC_WRITE_TIMEOUT":                   "10",
		"RIN_APP_RPC_IDLE_TIMEOUT":                    "30",
		"RIN_APP_PLUGINS_ENABLED":                     "false",
		"RIN_APP_PLUGINS_HEALTH_TIMEOUT_SECONDS":      "3",
		"RIN_APP_PLUGINS_PROXY_TIMEOUT_SECONDS":       "30",
		"RIN_APP_AUTH_ENABLED":                        "true",
		"RIN_APP_AUTH_SELF_SIGNUP_ENABLED":            "true",
		"RIN_APP_AUTH_ISSUER":                         "go-scaffold",
		"RIN_APP_AUTH_AUDIENCE":                       "go-scaffold-api",
		"RIN_APP_AUTH_SIGNING_KEY":                    "example-signing-key-at-least-32-bytes",
		"RIN_APP_AUTH_ACCESS_TOKEN_TTL_SECONDS":       "900",
		"RIN_APP_AUTH_REFRESH_TOKEN_TTL_SECONDS":      "604800",
		"RIN_APP_AUTH_REFRESH_TOKEN_PEPPER":           "example-refresh-pepper-at-least-32",
		"RIN_APP_AUTH_MFA_ISSUER":                     "go-scaffold",
		"RIN_APP_AUTH_MFA_SECRET_KEY":                 "example-mfa-secret-key-at-least-32",
		"RIN_APP_AUTH_LOGIN_MAX_FAILURES":             "5",
		"RIN_APP_AUTH_LOGIN_LOCK_MINUTES":             "15",
		"RIN_APP_AUTH_LOGIN_CAPTCHA_ENABLED":          "false",
		"RIN_APP_AUTH_CAPTCHA_TTL_SECONDS":            "120",
		"RIN_APP_AUTH_INVITATION_TTL_SECONDS":         "86400",
		"RIN_APP_AUTH_PASSWORD_RESET_TTL_SECONDS":     "1800",
		"RIN_APP_AUTH_NOTIFICATION_DRIVER":            "debug",
		"RIN_APP_AUTH_SMTP_HOST":                      "127.0.0.1",
		"RIN_APP_AUTH_SMTP_PORT":                      "1025",
		"RIN_APP_AUTH_SMTP_USERNAME":                  "mailer",
		"RIN_APP_AUTH_SMTP_PASSWORD":                  "example-smtp-password",
		"RIN_APP_AUTH_SMTP_FROM":                      "no-reply@example.invalid",
		"RIN_APP_AUTH_SMTP_FROM_NAME":                 "Aoi Admin",
		"RIN_APP_AUTH_SMTP_STARTTLS":                  "false",
		"RIN_APP_AUTH_PASSWORD_MIN_LENGTH":            "8",
		"RIN_APP_AUTH_PASSWORD_REQUIRE_LOWER":         "false",
		"RIN_APP_AUTH_PASSWORD_REQUIRE_UPPER":         "false",
		"RIN_APP_AUTH_PASSWORD_REQUIRE_NUMBER":        "false",
		"RIN_APP_AUTH_PASSWORD_REQUIRE_SYMBOL":        "false",
		"RIN_APP_AUTH_CASBIN_RELOAD_INTERVAL_SECONDS": "300",
		"RIN_APP_MIGRATION_AUTO_APPLY":                "true",
		"RIN_APP_MIGRATION_DIR":                       "./internal/migrations",
		"RIN_APP_CORS_ENABLED":                        "true",
		"RIN_APP_CORS_ALLOW_ORIGINS":                  "*",
		"RIN_APP_CORS_ALLOW_METHODS":                  "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		"RIN_APP_CORS_ALLOW_HEADERS":                  "Origin,Content-Type,X-Request-ID,Authorization",
		"RIN_APP_CORS_EXPOSE_HEADERS":                 "X-Request-ID",
		"RIN_APP_CORS_ALLOW_CREDENTIALS":              "false",
		"RIN_APP_CORS_MAX_AGE":                        "3600",
		"AUTH_SIGNING_KEY":                            "example-signing-key-at-least-32-bytes",
		"AUTH_REFRESH_TOKEN_PEPPER":                   "example-refresh-pepper-at-least-32",
		"AUTH_MFA_SECRET_KEY":                         "example-mfa-secret-key-at-least-32",
		"AUTH_SMTP_HOST":                              "127.0.0.1",
		"AUTH_SMTP_PORT":                              "1025",
		"AUTH_SMTP_USERNAME":                          "mailer",
		"AUTH_SMTP_PASSWORD":                          "example-smtp-password",
		"AUTH_SMTP_FROM":                              "no-reply@example.invalid",
		"AUTH_SMTP_FROM_NAME":                         "Aoi Admin",
		"AUTH_SMTP_STARTTLS":                          "false",
		"AOI_DEMO1_PLUGIN_SECRET":                     "example-demo-plugin-secret",
	}

	switch fileName {
	case "config.production.example.yaml", "postgres-production.example.yaml":
		env["RIN_APP_SERVER_MODE"] = "release"
		env["RIN_APP_DB_DRIVER"] = "postgres"
		env["RIN_APP_DEMO_ENABLED"] = "false"
		env["RIN_APP_DEMO_APPLY_SCHEMA_ON_START"] = "false"
		env["RIN_APP_AUTH_SELF_SIGNUP_ENABLED"] = "false"
		env["RIN_APP_AUTH_NOTIFICATION_DRIVER"] = "smtp"
		env["RIN_APP_MIGRATION_AUTO_APPLY"] = "false"
		env["RIN_APP_CORS_ALLOW_ORIGINS"] = "https://admin.example.invalid"
	case "mysql-redis.example.yaml":
		env["RIN_APP_DB_DRIVER"] = "mysql"
		env["RIN_APP_DB_PORT"] = "3306"
		env["RIN_APP_REDIS_ENABLED"] = "true"
	case "smtp-auth.example.yaml":
		env["RIN_APP_AUTH_SELF_SIGNUP_ENABLED"] = "false"
		env["RIN_APP_AUTH_NOTIFICATION_DRIVER"] = "smtp"
	case "storage-media.example.yaml":
		env["RIN_APP_STORAGE_ENABLED"] = "true"
	case "plugins-sidecar.example.yaml":
		env["RIN_APP_PLUGINS_ENABLED"] = "true"
	}

	for key, value := range env {
		t.Setenv(key, value)
	}
}
