package config

import (
	"fmt"
	"strings"

	"github.com/rei0721/go-scaffold/pkg/configloader"
)

type ConfigDiagnosticSeverity string

const (
	ConfigDiagnosticError ConfigDiagnosticSeverity = "error"
)

type ConfigDiagnostic struct {
	Section  string
	Path     string
	Message  string
	EnvNames []string
	Severity ConfigDiagnosticSeverity
}

func LoadDiagnostics(configPath string) (*Config, []ConfigDiagnostic, error) {
	m := &manager{
		v:     configloader.New(),
		hooks: make([]HookHandler, 0),
	}
	m.configPath = configPath

	LoadEnv()
	m.v.SetConfigFile(configPath)
	if err := m.v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := m.processEnvSubstitution(); err != nil {
		return nil, nil, fmt.Errorf("failed to process env substitution: %w", err)
	}

	cfg := &Config{}
	if err := m.v.Unmarshal(cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	OverrideWithEnvExcept(cfg, cfg.EnvOverride.DisabledPaths)
	return cfg, cfg.Diagnostics(), nil
}

func (c *Config) Diagnostics() []ConfigDiagnostic {
	if c == nil {
		return []ConfigDiagnostic{newDiagnostic("", "", "config is required")}
	}
	var diagnostics []ConfigDiagnostic
	add := func(section string, path string, message string) {
		diagnostics = append(diagnostics, newDiagnostic(section, path, message))
	}

	c.diagnoseServer(add)
	c.diagnoseDatabase(add)
	c.diagnoseRedis(add)
	c.diagnoseLogger(add)
	c.diagnoseI18n(add)
	c.diagnoseExecutor(add)
	c.diagnoseStorage(add)
	c.diagnoseCORS(add)
	c.diagnoseRPC(add)
	c.diagnoseAuth(add)
	c.diagnoseWebUI(add)
	c.diagnosePlugins(add)
	if c.Plugins.Enabled && !c.Auth.Enabled {
		add(AppPluginsName, "plugins.enabled", "auth must be enabled when plugins are enabled")
	}
	c.EnvOverride.DisabledPaths = normalizeConfigPaths(c.EnvOverride.DisabledPaths)
	return diagnostics
}

func newDiagnostic(section string, path string, message string) ConfigDiagnostic {
	return ConfigDiagnostic{
		Section:  section,
		Path:     path,
		Message:  message,
		EnvNames: EnvNamesForPath(path),
		Severity: ConfigDiagnosticError,
	}
}

func (c *Config) diagnoseServer(add func(string, string, string)) {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		add(AppServerName, "server.port", "port must be between 1 and 65535")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		add(AppServerName, "server.mode", "mode must be debug, release, or test")
	}
	if c.Server.ReadTimeout <= 0 {
		add(AppServerName, "server.read_timeout", "read_timeout must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		add(AppServerName, "server.write_timeout", "write_timeout must be positive")
	}
}

func (c *Config) diagnoseDatabase(add func(string, string, string)) {
	driver := strings.ToLower(strings.TrimSpace(c.Database.Driver))
	validDrivers := map[string]bool{"postgres": true, "mysql": true, "sqlite": true}
	if !validDrivers[driver] {
		add(AppDatabaseName, "database.driver", "driver must be postgres, mysql, or sqlite")
		return
	}
	if driver != "sqlite" {
		if strings.TrimSpace(c.Database.Host) == "" {
			add(AppDatabaseName, "database.host", "host is required when database driver is mysql or postgres")
		}
		if c.Database.Port <= 0 || c.Database.Port > 65535 {
			add(AppDatabaseName, "database.port", "port must be between 1 and 65535")
		}
		if strings.TrimSpace(c.Database.User) == "" {
			add(AppDatabaseName, "database.user", "user is required when database driver is mysql or postgres")
		}
	}
	if strings.TrimSpace(c.Database.DBName) == "" {
		add(AppDatabaseName, "database.dbname", "dbname is required")
	}
	if c.Database.MaxOpenConns < 0 {
		add(AppDatabaseName, "database.max_open_conns", "max_open_conns must be non-negative")
	}
	if c.Database.MaxIdleConns < 0 {
		add(AppDatabaseName, "database.max_idle_conns", "max_idle_conns must be non-negative")
	}
}

func (c *Config) diagnoseRedis(add func(string, string, string)) {
	if !c.Redis.Enabled {
		return
	}
	if strings.TrimSpace(c.Redis.Host) == "" {
		add(AppRedisName, "redis.host", "host is required when redis is enabled")
	}
	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		add(AppRedisName, "redis.port", "port must be between 1 and 65535")
	}
	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		add(AppRedisName, "redis.db", "db must be between 0 and 15")
	}
	if c.Redis.PoolSize < 0 {
		add(AppRedisName, "redis.pool_size", "pool_size must be non-negative")
	}
}

func (c *Config) diagnoseLogger(add func(string, string, string)) {
	if !stringInSet(c.Logger.Level, "debug", "info", "warn", "error") {
		add(AppLoggerName, "logger.level", "level must be debug, info, warn, or error")
	}
	if !stringInSet(c.Logger.Format, "json", "console") {
		add(AppLoggerName, "logger.format", "format must be json or console")
	}
	if !stringInSet(c.Logger.Output, "stdout", "file", "both") {
		add(AppLoggerName, "logger.output", "output must be stdout, file, or both")
	}
}

func (c *Config) diagnoseI18n(add func(string, string, string)) {
	if strings.TrimSpace(c.I18n.Default) == "" {
		add(AppI18nName, "i18n.default", "default locale is required")
	}
	if len(c.I18n.Supported) == 0 {
		add(AppI18nName, "i18n.supported", "at least one supported locale is required")
		return
	}
	found := false
	for _, supported := range c.I18n.Supported {
		if supported == c.I18n.Default {
			found = true
			break
		}
	}
	if !found {
		add(AppI18nName, "i18n.default", "default locale must be in supported list")
	}
}

func (c *Config) diagnoseExecutor(add func(string, string, string)) {
	if !c.Executor.Enabled {
		return
	}
	if len(c.Executor.Pools) == 0 {
		add(AppExecutorName, "executor.pools", "at least one pool is required when executor is enabled")
		return
	}
	seen := map[string]struct{}{}
	for i, pool := range c.Executor.Pools {
		basePath := fmt.Sprintf("executor.pools.%d", i)
		name := strings.TrimSpace(pool.Name)
		if name == "" {
			add(AppExecutorName, basePath+".name", "pool name is required")
		} else if _, ok := seen[name]; ok {
			add(AppExecutorName, basePath+".name", "duplicate pool name: "+name)
		} else {
			seen[name] = struct{}{}
		}
		if pool.Size <= 0 {
			add(AppExecutorName, basePath+".size", "pool size must be positive")
		}
		if pool.Size > 10000 {
			add(AppExecutorName, basePath+".size", "pool size must not exceed 10000")
		}
		if pool.Expiry < 0 {
			add(AppExecutorName, basePath+".expiry", "pool expiry must be non-negative")
		}
	}
}

func (c *Config) diagnoseStorage(add func(string, string, string)) {
	if !c.Storage.Enabled {
		return
	}
	if !stringInSet(c.Storage.FSType, "os", "memory", "readonly", "basepath") {
		add(AppStorageName, "storage.fs_type", "fs_type must be one of: os, memory, readonly, basepath")
	}
	if c.Storage.FSType == "basepath" && strings.TrimSpace(c.Storage.BasePath) == "" {
		add(AppStorageName, "storage.base_path", "base_path is required when fs_type is basepath")
	}
	if c.Storage.WatchBufferSize < 0 {
		add(AppStorageName, "storage.watch_buffer_size", "watch_buffer_size must be non-negative")
	}
}

func (c *Config) diagnoseCORS(add func(string, string, string)) {
	if !c.CORS.Enabled {
		return
	}
	if c.CORS.AllowCredentials {
		for _, origin := range c.CORS.AllowOrigins {
			if origin == "*" {
				add(AppCORSName, "cors.allow_origins", "allow_origins cannot contain wildcard \"*\" when allow_credentials is true")
				break
			}
		}
	}
	if c.CORS.MaxAge < 0 {
		add(AppCORSName, "cors.max_age", "max_age must be non-negative")
	}
}

func (c *Config) diagnoseRPC(add func(string, string, string)) {
	if !c.RPC.Enabled {
		return
	}
	c.RPC.ApplyDefaults()
	if strings.TrimSpace(c.RPC.Host) == "" {
		add(AppRPCName, "rpc.host", "host is required")
	}
	if c.RPC.Port <= 0 || c.RPC.Port > 65535 {
		add(AppRPCName, "rpc.port", "port must be between 1 and 65535")
	}
	if c.RPC.ReadTimeout <= 0 {
		add(AppRPCName, "rpc.read_timeout", "read_timeout must be positive")
	}
	if c.RPC.WriteTimeout <= 0 {
		add(AppRPCName, "rpc.write_timeout", "write_timeout must be positive")
	}
	if c.RPC.IdleTimeout < 0 {
		add(AppRPCName, "rpc.idle_timeout", "idle_timeout must be non-negative")
	}
}

func (c *Config) diagnoseAuth(add func(string, string, string)) {
	if !c.Auth.Enabled {
		return
	}
	c.Auth.ApplyDefaults()
	if strings.TrimSpace(c.Auth.Issuer) == "" {
		add(AppAuthName, "auth.issuer", "issuer is required")
	}
	if len(c.Auth.SigningKey) < 32 {
		add(AppAuthName, "auth.signing_key", "signing_key must be at least 32 bytes")
	}
	if strings.TrimSpace(c.Auth.RefreshTokenPepper) == "" {
		add(AppAuthName, "auth.refresh_token_pepper", "refresh_token_pepper is required")
	}
	if len(c.Auth.MFASecretKey) < 32 {
		add(AppAuthName, "auth.mfa_secret_key", "mfa_secret_key must be at least 32 bytes")
	}
	if c.Auth.AccessTokenTTLSeconds <= 0 {
		add(AppAuthName, "auth.access_token_ttl_seconds", "access_token_ttl_seconds must be positive")
	}
	if c.Auth.RefreshTokenTTLSeconds <= 0 {
		add(AppAuthName, "auth.refresh_token_ttl_seconds", "refresh_token_ttl_seconds must be positive")
	}
	if c.Auth.InvitationTTLSeconds <= 0 {
		add(AppAuthName, "auth.invitation_ttl_seconds", "invitation_ttl_seconds must be positive")
	}
	if c.Auth.PasswordResetTTLSeconds <= 0 {
		add(AppAuthName, "auth.password_reset_ttl_seconds", "password_reset_ttl_seconds must be positive")
	}
	if c.Auth.LoginMaxFailures <= 0 {
		add(AppAuthName, "auth.login_max_failures", "login_max_failures must be positive")
	}
	if c.Auth.LoginLockMinutes <= 0 {
		add(AppAuthName, "auth.login_lock_minutes", "login_lock_minutes must be positive")
	}
	if c.Auth.LoginCaptchaEnabled && c.Auth.CaptchaTTLSeconds <= 0 {
		add(AppAuthName, "auth.captcha_ttl_seconds", "captcha_ttl_seconds must be positive when login captcha is enabled")
	}
	if c.Auth.PasswordPolicy.MinLength <= 0 {
		add(AppAuthName, "auth.password_policy.min_length", "password policy min_length must be positive")
	}
	if strings.EqualFold(c.Auth.NotificationDriver, "smtp") {
		if strings.TrimSpace(c.Auth.SMTP.Host) == "" {
			add(AppAuthName, "auth.smtp.host", "smtp host is required when notification_driver is smtp")
		}
		if c.Auth.SMTP.Port <= 0 {
			add(AppAuthName, "auth.smtp.port", "smtp port is required when notification_driver is smtp")
		}
		if strings.TrimSpace(c.Auth.SMTP.From) == "" {
			add(AppAuthName, "auth.smtp.from", "smtp from is required when notification_driver is smtp")
		}
	}
}

func (c *Config) diagnoseWebUI(add func(string, string, string)) {
	c.WebUI.ApplyDefaults()
	if !c.WebUI.EnabledValue() {
		return
	}
	if c.WebUI.MountPath == "" || !strings.HasPrefix(c.WebUI.MountPath, "/") {
		add(AppWebUIName, "webui.mount_path", "mount_path must start with /")
	}
	if c.WebUI.MountPath == "/" {
		add(AppWebUIName, "webui.mount_path", "mount_path cannot be /")
	}
	if webUIReservedPath(c.WebUI.MountPath) {
		add(AppWebUIName, "webui.mount_path", "mount_path conflicts with reserved API or probe path")
	}
	if strings.TrimSpace(c.WebUI.DistDir) == "" {
		add(AppWebUIName, "webui.dist_dir", "dist_dir is required")
	}
	c.WebUI.PublicBaseURL = strings.TrimRight(strings.TrimSpace(c.WebUI.PublicBaseURL), "/")
}

func (c *Config) diagnosePlugins(add func(string, string, string)) {
	if !c.Plugins.Enabled {
		return
	}
	c.Plugins.ApplyDefaults()
	seen := map[string]struct{}{}
	for i := range c.Plugins.Items {
		item := &c.Plugins.Items[i]
		if err := item.Validate(); err != nil {
			add(AppPluginsName, fmt.Sprintf("plugins.items.%d", i), err.Error())
		}
		if id := strings.TrimSpace(item.ID); id != "" {
			if _, ok := seen[id]; ok {
				add(AppPluginsName, fmt.Sprintf("plugins.items.%d.id", i), fmt.Sprintf("duplicate plugin id %q", id))
			}
			seen[id] = struct{}{}
		}
	}
}

func stringInSet(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}
