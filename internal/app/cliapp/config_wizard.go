package cliapp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/configloader"
	"github.com/rei0721/go-scaffold/types/constants"
)

var coreSecretPaths = []string{
	"auth.signing_key",
	"auth.refresh_token_pepper",
	"auth.mfa_secret_key",
}

const (
	privacyActionForceFile      = "force_file"
	privacyActionRuntimeEnvOnly = "runtime_env_only"
	privacyActionSkip           = "skip"
)

type privacyPersistPlan struct {
	fileUpdates         map[string]string
	forceFileUpdates    map[string]string
	runtimeEnvOnlyPaths []string
}

func newPrivacyPersistPlan() privacyPersistPlan {
	return privacyPersistPlan{
		fileUpdates:      map[string]string{},
		forceFileUpdates: map[string]string{},
	}
}

func (plan privacyPersistPlan) hasChanges() bool {
	return len(plan.fileUpdates) > 0 || len(plan.forceFileUpdates) > 0 || len(plan.runtimeEnvOnlyPaths) > 0
}

// DiscoverConfigFiles 返回启动向导可选配置，默认配置优先。
func DiscoverConfigFiles() []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	add := func(path string) {
		path = filepath.Clean(strings.TrimSpace(path))
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		if _, err := os.Stat(path); err != nil {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	add(constants.AppDefaultConfigPath)
	matches, _ := filepath.Glob(filepath.Join("configs", "*.yaml"))
	sort.Strings(matches)
	for _, match := range matches {
		add(match)
	}
	return out
}

func PrintConfigSummary(stdout io.Writer, configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "配置文件：%s\n", configPath)
	fmt.Fprintf(stdout, "HTTP：%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Fprintf(stdout, "数据库：%s %s\n", cfg.Database.Driver, cfg.Database.DBName)
	fmt.Fprintf(stdout, "Redis：%v\n", cfg.Redis.Enabled)
	fmt.Fprintf(stdout, "Storage：%v\n", cfg.Storage.Enabled)
	fmt.Fprintf(stdout, "IAM：%v\n", cfg.Auth.Enabled)
	if cfg.Logger.FilePath != "" {
		fmt.Fprintf(stdout, "应用日志：%s\n", cfg.Logger.FilePath)
	}
	return nil
}

// ApplyPrivacyUpdates 使用配置管理器持久化隐私配置。
func ApplyPrivacyUpdates(configPath string, updates map[string]string, options ...appconfig.UpdateOption) error {
	paths := make([]string, 0, len(updates))
	normalized := make(map[string]string, len(updates))
	for path, value := range updates {
		value = strings.TrimSpace(value)
		if value == "" || !supportedPrivacyPath(path) {
			continue
		}
		normalized[path] = value
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return nil
	}
	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		return err
	}
	updateOptions := []appconfig.UpdateOption{appconfig.WithPersistedPaths(paths...)}
	updateOptions = append(updateOptions, options...)
	err := manager.Update(func(cfg *appconfig.Config) {
		for path, value := range normalized {
			applyPrivacyValue(cfg, path, value)
		}
	}, updateOptions...)
	if err != nil {
		return err
	}
	return nil
}

// ApplyPrivacyRuntimeEnvOnly 校验并应用真实环境变量中的隐私配置，不改写配置文件。
func ApplyPrivacyRuntimeEnvOnly(configPath string, paths []string) error {
	normalized := make([]string, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" || !supportedPrivacyPath(path) {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}
	if len(normalized) == 0 {
		return nil
	}

	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		return err
	}
	return manager.Update(func(*appconfig.Config) {}, appconfig.WithPersistedPaths(normalized...), appconfig.WithEnvManagedPersistMode(appconfig.EnvManagedPersistRuntimeEnvOnly))
}

func applyPrivacyValue(cfg *appconfig.Config, path string, value string) bool {
	switch path {
	case "auth.signing_key":
		cfg.Auth.SigningKey = value
	case "auth.refresh_token_pepper":
		cfg.Auth.RefreshTokenPepper = value
	case "auth.mfa_secret_key":
		cfg.Auth.MFASecretKey = value
	case "database.password":
		cfg.Database.Password = value
	case "redis.password":
		cfg.Redis.Password = value
	case "auth.smtp.password":
		cfg.Auth.SMTP.Password = value
	default:
		return false
	}
	return true
}

func supportedPrivacyPath(path string) bool {
	switch path {
	case "auth.signing_key", "auth.refresh_token_pepper", "auth.mfa_secret_key", "database.password", "redis.password", "auth.smtp.password":
		return true
	default:
		return false
	}
}

func isGeneratedSecretPath(path string) bool {
	for _, candidate := range coreSecretPaths {
		if path == candidate {
			return true
		}
	}
	return false
}

func randomSecret() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "change-me-generated-secret-32-bytes"
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func isExampleConfig(path string) bool {
	return strings.Contains(strings.ToLower(filepath.Base(path)), ".example.")
}

func privacyPathIsEnvManaged(configPath string, path string) (bool, error) {
	for _, envName := range appconfig.EnvNamesForPath(path) {
		if value, ok := os.LookupEnv(envName); ok && value != "" {
			return true, nil
		}
	}
	return configloader.YAMLPathContainsEnvPlaceholder(configPath, path)
}
