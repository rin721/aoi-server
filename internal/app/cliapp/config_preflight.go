package cliapp

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/pkg/configloader"
)

const (
	preflightActionSkip           = "skip"
	preflightActionRuntimeEnvOnly = "runtime_env_only"

	preflightDatabaseActionFile   = "file"
	preflightDatabaseActionSQLite = "sqlite"

	preflightSMTPActionFile  = "file"
	preflightSMTPActionDebug = "debug"
)

var databaseRequiredPaths = []string{
	"database.host",
	"database.port",
	"database.user",
	"database.dbname",
}

var smtpRequiredPaths = []string{
	"auth.smtp.host",
	"auth.smtp.port",
	"auth.smtp.from",
}

func preflightConfigForStart(ctx *cli.Context, configPath string) (*appconfig.Config, bool, error) {
	repaired := false
	for attempt := 0; attempt < 4; attempt++ {
		cfg, diagnostics, err := appconfig.LoadDiagnostics(configPath)
		if err != nil {
			return nil, repaired, err
		}
		if len(diagnostics) == 0 {
			return cfg, repaired, nil
		}
		printConfigDiagnostics(ctx.Stdout, configPath, diagnostics)
		if !canPromptPreflightRepair(ctx) {
			return nil, repaired, newConfigDiagnosticsError(configPath, diagnostics)
		}
		changed, err := promptPreflightRepairs(ctx, configPath, cfg, diagnostics)
		if err != nil {
			return nil, repaired, err
		}
		if !changed {
			return nil, repaired, newConfigDiagnosticsError(configPath, diagnostics)
		}
		repaired = true
	}
	_, diagnostics, err := appconfig.LoadDiagnostics(configPath)
	if err != nil {
		return nil, repaired, err
	}
	return nil, repaired, newConfigDiagnosticsError(configPath, diagnostics)
}

func actionableConfigLoadError(configPath string, loadErr error) error {
	if loadErr == nil {
		return nil
	}
	_, diagnostics, diagErr := appconfig.LoadDiagnostics(configPath)
	if diagErr == nil && len(diagnostics) > 0 {
		return newConfigDiagnosticsError(configPath, diagnostics)
	}
	return coreSecretConfigError(configPath, loadErr)
}

func canPromptPreflightRepair(ctx *cli.Context) bool {
	if ctx == nil || ctx.UI == nil || ctx.GetBool("yes") {
		return false
	}
	if value, ok := cli.PromptAnswer(ctx.UI, "privacy"); ok {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "false", "f", "no", "n", "0":
			return false
		}
	}
	return true
}

func promptPreflightRepairs(ctx *cli.Context, configPath string, cfg *appconfig.Config, diagnostics []appconfig.ConfigDiagnostic) (bool, error) {
	changed := false
	if hasSectionDiagnostics(diagnostics, appconfig.AppDatabaseName) &&
		(strings.EqualFold(cfg.Database.Driver, "postgres") || strings.EqualFold(cfg.Database.Driver, "mysql")) {
		ok, err := promptDatabasePreflightRepair(ctx, configPath, cfg)
		if err != nil {
			return false, err
		}
		changed = changed || ok
	}
	if hasAuthCoreDiagnostics(diagnostics) {
		ok, err := promptCoreSecretRecovery(ctx, configPath)
		if err != nil {
			return false, err
		}
		changed = changed || ok
	}
	if hasSMTPDiagnostics(diagnostics) {
		ok, err := promptSMTPPreflightRepair(ctx, configPath, cfg)
		if err != nil {
			return false, err
		}
		changed = changed || ok
	}
	return changed, nil
}

func promptDatabasePreflightRepair(ctx *cli.Context, configPath string, cfg *appconfig.Config) (bool, error) {
	action, err := cli.SelectKey(ctx.Context, ctx.UI, "preflight.database.action", "Database config is incomplete; choose a repair action", []cli.SelectOption{
		{Value: preflightDatabaseActionSQLite, Label: "switch to sqlite", Description: "Use a local SQLite database file for this deployment"},
		{Value: preflightDatabaseActionFile, Label: "write config file", Description: "Write host, port, user, and dbname into the current config file"},
		{Value: preflightActionRuntimeEnvOnly, Label: "use environment variables", Description: "Keep config placeholders and require real DB environment variables"},
		{Value: preflightActionSkip, Label: "skip", Description: "Leave database config unchanged and show the blocking error"},
	})
	if err != nil {
		return false, err
	}
	switch action {
	case preflightDatabaseActionSQLite:
		if isExampleConfig(configPath) {
			return false, exampleConfigWriteError(configPath)
		}
		updates := []configloader.YAMLScalarUpdate{
			{Kind: configloader.YAMLScalarString, Path: "database.driver", Value: "sqlite"},
			{Kind: configloader.YAMLScalarString, Path: "database.host", Value: ""},
			{Kind: configloader.YAMLScalarInt, Path: "database.port", Value: "0"},
			{Kind: configloader.YAMLScalarString, Path: "database.user", Value: ""},
			{Kind: configloader.YAMLScalarString, Path: "database.dbname", Value: "./data/app.db"},
			{Kind: configloader.YAMLScalarInt, Path: "database.max_open_conns", Value: "1"},
			{Kind: configloader.YAMLScalarInt, Path: "database.max_idle_conns", Value: "1"},
		}
		paths := []string{
			"database.driver",
			"database.host",
			"database.port",
			"database.user",
			"database.dbname",
			"database.max_open_conns",
			"database.max_idle_conns",
		}
		if err := applyConfigForceFileScalarUpdates(configPath, updates, paths); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("Database config switched to sqlite.")
		return true, nil
	case preflightDatabaseActionFile:
		if isExampleConfig(configPath) {
			return false, exampleConfigWriteError(configPath)
		}
		host, err := promptConfigStringValue(ctx, "database.host", cfg.Database.Host)
		if err != nil {
			return false, err
		}
		port, err := promptConfigIntValue(ctx, "database.port", defaultDatabasePort(cfg.Database.Driver, cfg.Database.Port))
		if err != nil {
			return false, err
		}
		user, err := promptConfigStringValue(ctx, "database.user", cfg.Database.User)
		if err != nil {
			return false, err
		}
		dbname, err := promptConfigStringValue(ctx, "database.dbname", cfg.Database.DBName)
		if err != nil {
			return false, err
		}
		updates := []configloader.YAMLScalarUpdate{
			{Kind: configloader.YAMLScalarString, Path: "database.host", Value: host},
			{Kind: configloader.YAMLScalarInt, Path: "database.port", Value: strconv.Itoa(port)},
			{Kind: configloader.YAMLScalarString, Path: "database.user", Value: user},
			{Kind: configloader.YAMLScalarString, Path: "database.dbname", Value: dbname},
		}
		if err := applyConfigForceFileScalarUpdates(configPath, updates, databaseRequiredPaths); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("Database config written to config file.")
		return true, nil
	case preflightActionRuntimeEnvOnly:
		if err := applyRuntimeEnvOnlyConfigPathsDirect(configPath, databaseRequiredPaths, validatePreflightRuntimeEnvValue); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("Database config will be read from environment variables.")
		return true, nil
	case preflightActionSkip, "":
		return false, nil
	default:
		return false, fmt.Errorf("unknown database repair action %q", action)
	}
}

func promptSMTPPreflightRepair(ctx *cli.Context, configPath string, cfg *appconfig.Config) (bool, error) {
	action, err := cli.SelectKey(ctx.Context, ctx.UI, "preflight.smtp.action", "SMTP notification config is incomplete; choose a repair action", []cli.SelectOption{
		{Value: preflightSMTPActionDebug, Label: "switch to debug", Description: "Use debug notification output instead of SMTP"},
		{Value: preflightSMTPActionFile, Label: "write config file", Description: "Write SMTP host, port, and from into the current config file"},
		{Value: preflightActionRuntimeEnvOnly, Label: "use environment variables", Description: "Keep SMTP placeholders and require real SMTP environment variables"},
		{Value: preflightActionSkip, Label: "skip", Description: "Leave SMTP config unchanged and show the blocking error"},
	})
	if err != nil {
		return false, err
	}
	switch action {
	case preflightSMTPActionDebug:
		if isExampleConfig(configPath) {
			return false, exampleConfigWriteError(configPath)
		}
		updates := []configloader.YAMLScalarUpdate{
			{Kind: configloader.YAMLScalarString, Path: "auth.notification_driver", Value: "debug"},
		}
		if err := applyConfigForceFileScalarUpdates(configPath, updates, []string{"auth.notification_driver"}); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("Auth notification driver switched to debug.")
		return true, nil
	case preflightSMTPActionFile:
		if isExampleConfig(configPath) {
			return false, exampleConfigWriteError(configPath)
		}
		host, err := promptConfigStringValue(ctx, "auth.smtp.host", cfg.Auth.SMTP.Host)
		if err != nil {
			return false, err
		}
		port, err := promptConfigIntValue(ctx, "auth.smtp.port", defaultInt(cfg.Auth.SMTP.Port, 587))
		if err != nil {
			return false, err
		}
		from, err := promptConfigStringValue(ctx, "auth.smtp.from", cfg.Auth.SMTP.From)
		if err != nil {
			return false, err
		}
		updates := []configloader.YAMLScalarUpdate{
			{Kind: configloader.YAMLScalarString, Path: "auth.smtp.host", Value: host},
			{Kind: configloader.YAMLScalarInt, Path: "auth.smtp.port", Value: strconv.Itoa(port)},
			{Kind: configloader.YAMLScalarString, Path: "auth.smtp.from", Value: from},
		}
		if err := applyConfigForceFileScalarUpdates(configPath, updates, smtpRequiredPaths); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("SMTP config written to config file.")
		return true, nil
	case preflightActionRuntimeEnvOnly:
		if err := applyRuntimeEnvOnlyConfigPathsDirect(configPath, smtpRequiredPaths, validatePreflightRuntimeEnvValue); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("SMTP config will be read from environment variables.")
		return true, nil
	case preflightActionSkip, "":
		return false, nil
	default:
		return false, fmt.Errorf("unknown SMTP repair action %q", action)
	}
}

func promptConfigStringValue(ctx *cli.Context, path string, fallback string) (string, error) {
	value, err := cli.InputKey(ctx.Context, ctx.UI, "privacy."+path+".value", path, strings.TrimSpace(fallback))
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", path)
	}
	return value, nil
}

func promptConfigIntValue(ctx *cli.Context, path string, fallback int) (int, error) {
	value, err := cli.InputKey(ctx.Context, ctx.UI, "privacy."+path+".value", path, strconv.Itoa(fallback))
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", path)
	}
	if err := validatePreflightRuntimeEnvValue(path, strconv.Itoa(parsed)); err != nil {
		return 0, err
	}
	return parsed, nil
}

func defaultDatabasePort(driver string, current int) int {
	if current > 0 {
		return current
	}
	if strings.EqualFold(driver, "mysql") {
		return 3306
	}
	return 5432
}

func applyConfigForceFileScalarUpdates(configPath string, updates []configloader.YAMLScalarUpdate, disabledPaths []string) error {
	if len(updates) == 0 {
		return nil
	}
	disabled, err := configloader.YAMLStringSlice(configPath, "env_override.disabled_paths")
	if err != nil {
		return err
	}
	disabled = append(disabled, disabledPaths...)
	nextUpdates := append([]configloader.YAMLScalarUpdate(nil), updates...)
	nextUpdates = append(nextUpdates, configloader.YAMLScalarUpdate{
		Kind:          configloader.YAMLScalarStringSlice,
		Path:          "env_override.disabled_paths",
		Values:        normalizeConfigPathList(disabled),
		CreateMissing: true,
	})
	return configloader.UpdateYAMLScalars(configPath, nextUpdates, configloader.WithEnvPlaceholderOverwrite())
}

func applyRuntimeEnvOnlyConfigPathsDirect(configPath string, paths []string, validate func(string, string) error) error {
	normalized := normalizeConfigPathList(paths)
	if len(normalized) == 0 {
		return nil
	}
	for _, path := range normalized {
		if _, _, err := requireConfigRuntimeEnv(path, validate); err != nil {
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

func requireConfigRuntimeEnv(path string, validate func(string, string) error) (string, string, error) {
	for _, envName := range appconfig.EnvNamesForPath(path) {
		if value, ok := os.LookupEnv(envName); ok && strings.TrimSpace(value) != "" {
			value = strings.TrimSpace(value)
			if validate != nil {
				if err := validate(path, value); err != nil {
					return envName, value, err
				}
			}
			return envName, value, nil
		}
	}
	names := appconfig.EnvNamesForPath(path)
	if len(names) == 0 {
		return "", "", fmt.Errorf("%s has no environment variable mapping", path)
	}
	return "", "", fmt.Errorf("%s requires one of %s or choose to write values to the config file", path, strings.Join(names, ", "))
}

func validatePreflightRuntimeEnvValue(path string, value string) error {
	value = strings.TrimSpace(value)
	switch path {
	case "database.port", "auth.smtp.port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("environment value for %s must be an integer", path)
		}
		if port <= 0 || port > 65535 {
			return fmt.Errorf("environment value for %s must be between 1 and 65535", path)
		}
	case "auth.signing_key", "auth.refresh_token_pepper", "auth.mfa_secret_key":
		return validatePrivacyRuntimeEnvValue(path, value)
	default:
		if value == "" {
			return fmt.Errorf("environment value for %s is required", path)
		}
	}
	return nil
}

func hasSectionDiagnostics(diagnostics []appconfig.ConfigDiagnostic, section string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Section == section {
			return true
		}
	}
	return false
}

func hasAuthCoreDiagnostics(diagnostics []appconfig.ConfigDiagnostic) bool {
	for _, diagnostic := range diagnostics {
		for _, path := range coreSecretPaths {
			if diagnostic.Path == path {
				return true
			}
		}
	}
	return false
}

func hasSMTPDiagnostics(diagnostics []appconfig.ConfigDiagnostic) bool {
	for _, diagnostic := range diagnostics {
		if strings.HasPrefix(diagnostic.Path, "auth.smtp.") {
			return true
		}
	}
	return false
}

func normalizeConfigPathList(paths []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}
	return normalized
}

func exampleConfigWriteError(configPath string) error {
	return fmt.Errorf("example config %s is read-only for generated or repaired values; copy it to a real config file or set the listed environment variables", configPath)
}

type configDiagnosticsError struct {
	configPath  string
	diagnostics []appconfig.ConfigDiagnostic
}

func newConfigDiagnosticsError(configPath string, diagnostics []appconfig.ConfigDiagnostic) error {
	return configDiagnosticsError{configPath: configPath, diagnostics: diagnostics}
}

func (e configDiagnosticsError) Error() string {
	return formatConfigDiagnostics(e.configPath, e.diagnostics, true)
}

func printConfigDiagnostics(w interface{ Write([]byte) (int, error) }, configPath string, diagnostics []appconfig.ConfigDiagnostic) {
	if w == nil || len(diagnostics) == 0 {
		return
	}
	_, _ = fmt.Fprint(w, formatConfigDiagnostics(configPath, diagnostics, false))
}

func formatConfigDiagnostics(configPath string, diagnostics []appconfig.ConfigDiagnostic, includeAdvice bool) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Config preflight found %d blocking item(s) in %s:\n", len(diagnostics), configPath)
	lastSection := ""
	for _, diagnostic := range diagnostics {
		section := diagnostic.Section
		if section == "" {
			section = "config"
		}
		if section != lastSection {
			fmt.Fprintf(&builder, "[%s]\n", section)
			lastSection = section
		}
		path := diagnostic.Path
		if path == "" {
			path = section
		}
		fmt.Fprintf(&builder, "  - %s: %s", path, diagnostic.Message)
		if len(diagnostic.EnvNames) > 0 {
			fmt.Fprintf(&builder, " (env: %s)", strings.Join(diagnostic.EnvNames, " or "))
		}
		builder.WriteByte('\n')
	}
	if includeAdvice {
		builder.WriteString("Set the listed environment variables, edit the config file, or run interactive `run` without --yes to use the guided repair flow.")
	}
	return builder.String()
}
