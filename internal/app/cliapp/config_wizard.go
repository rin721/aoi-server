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
	paths, normalized := normalizePrivacyUpdates(updates)
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

func applyPrivacyForceFileUpdates(configPath string, updates map[string]string) error {
	paths, normalized := normalizePrivacyUpdates(updates)
	if len(paths) == 0 {
		return nil
	}

	yamlUpdates := make([]configloader.YAMLScalarUpdate, 0, len(paths)+1)
	for _, path := range paths {
		yamlUpdates = append(yamlUpdates, configloader.YAMLScalarUpdate{
			Kind:  configloader.YAMLScalarString,
			Path:  path,
			Value: normalized[path],
		})
	}
	disabledPaths, err := configloader.YAMLStringSlice(configPath, "env_override.disabled_paths")
	if err != nil {
		return err
	}
	disabledPaths = append(disabledPaths, paths...)
	yamlUpdates = append(yamlUpdates, configloader.YAMLScalarUpdate{
		Kind:          configloader.YAMLScalarStringSlice,
		Path:          "env_override.disabled_paths",
		Values:        disabledPaths,
		CreateMissing: true,
	})
	return configloader.UpdateYAMLScalars(configPath, yamlUpdates, configloader.WithEnvPlaceholderOverwrite())
}

// ApplyPrivacyRuntimeEnvOnly 校验并应用真实环境变量中的隐私配置，不改写配置文件。
func ApplyPrivacyRuntimeEnvOnly(configPath string, paths []string) error {
	normalized := normalizePrivacyPaths(paths)
	if len(normalized) == 0 {
		return nil
	}

	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		return err
	}
	return manager.Update(func(*appconfig.Config) {}, appconfig.WithPersistedPaths(normalized...), appconfig.WithEnvManagedPersistMode(appconfig.EnvManagedPersistRuntimeEnvOnly))
}

func applyPrivacyRuntimeEnvOnlyDirect(configPath string, paths []string) error {
	normalized := normalizePrivacyPaths(paths)
	if len(normalized) == 0 {
		return nil
	}
	for _, path := range normalized {
		if _, _, err := requirePrivacyRuntimeEnv(path); err != nil {
			return err
		}
	}
	disabledPaths, err := configloader.YAMLStringSlice(configPath, "env_override.disabled_paths")
	if err != nil {
		return err
	}
	remove := make(map[string]struct{}, len(normalized))
	for _, path := range normalized {
		remove[path] = struct{}{}
	}
	kept := make([]string, 0, len(disabledPaths))
	changed := false
	for _, path := range disabledPaths {
		if _, ok := remove[path]; ok {
			changed = true
			continue
		}
		kept = append(kept, path)
	}
	if !changed {
		return nil
	}
	return configloader.UpdateYAMLScalars(configPath, []configloader.YAMLScalarUpdate{
		{
			Kind:          configloader.YAMLScalarStringSlice,
			Path:          "env_override.disabled_paths",
			Values:        kept,
			CreateMissing: true,
		},
	})
}

func normalizePrivacyUpdates(updates map[string]string) ([]string, map[string]string) {
	paths := make([]string, 0, len(updates))
	normalized := make(map[string]string, len(updates))
	for path, value := range updates {
		path = strings.TrimSpace(path)
		value = strings.TrimSpace(value)
		if value == "" || !supportedPrivacyPath(path) {
			continue
		}
		normalized[path] = value
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, normalized
}

func normalizePrivacyPaths(paths []string) []string {
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
	sort.Strings(normalized)
	return normalized
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
	cfg, err := loadConfig(configPath)
	if err == nil {
		for _, disabledPath := range cfg.EnvOverride.DisabledPaths {
			if disabledPath == path {
				return true, nil
			}
		}
	}
	for _, envName := range appconfig.EnvNamesForPath(path) {
		if value, ok := os.LookupEnv(envName); ok && value != "" {
			return true, nil
		}
	}
	return configloader.YAMLPathContainsEnvPlaceholder(configPath, path)
}

func requirePrivacyRuntimeEnv(path string) (string, string, error) {
	for _, envName := range appconfig.EnvNamesForPath(path) {
		if value, ok := os.LookupEnv(envName); ok && strings.TrimSpace(value) != "" {
			value = strings.TrimSpace(value)
			if err := validatePrivacyRuntimeEnvValue(path, value); err != nil {
				return envName, value, err
			}
			return envName, value, nil
		}
	}
	names := appconfig.EnvNamesForPath(path)
	if len(names) == 0 {
		return "", "", fmt.Errorf("%s is managed by environment placeholder but has no environment variable mapping", path)
	}
	return "", "", fmt.Errorf("%s is managed by environment placeholder; set one of %s or choose to write generated values to the config file", path, strings.Join(names, ", "))
}

func validatePrivacyRuntimeEnvValue(path string, value string) error {
	switch path {
	case "auth.signing_key":
		if len(value) < 32 {
			return fmt.Errorf("environment value for %s must be at least 32 bytes", path)
		}
	case "auth.refresh_token_pepper":
		if value == "" {
			return fmt.Errorf("environment value for %s is required", path)
		}
	case "auth.mfa_secret_key":
		if len(value) < 32 {
			return fmt.Errorf("environment value for %s must be at least 32 bytes", path)
		}
	}
	return nil
}

func isCoreSecretConfigError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	if !strings.Contains(message, "auth config:") {
		return false
	}
	for _, needle := range []string{
		"signing_key must be at least 32 bytes",
		"refresh_token_pepper is required",
		"mfa_secret_key must be at least 32 bytes",
	} {
		if strings.Contains(message, needle) {
			return true
		}
	}
	return false
}

func coreSecretConfigError(configPath string, err error) error {
	if !isCoreSecretConfigError(err) {
		return err
	}
	return fmt.Errorf("%w; IAM core secrets are missing or too short in %s. Set %s, or run the interactive `run` flow and choose generated privacy config", err, configPath, coreSecretEnvHelp())
}

func coreSecretEnvHelp() string {
	parts := make([]string, 0, len(coreSecretPaths))
	for _, path := range coreSecretPaths {
		names := appconfig.EnvNamesForPath(path)
		if len(names) == 0 {
			parts = append(parts, path)
			continue
		}
		parts = append(parts, path+" ("+strings.Join(names, " or ")+")")
	}
	return strings.Join(parts, ", ")
}
