package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/rei0721/go-scaffold/pkg/configloader"
)

const envOverrideDisabledPathsConfigPath = "env_override.disabled_paths"

func (m *manager) persistConfigUpdate(newCfg *Config, paths []string, options updateOptions) error {
	if strings.TrimSpace(m.configPath) == "" {
		return fmt.Errorf("configuration file path is not available")
	}
	if options.envManagedPersistMode < EnvManagedPersistReject || options.envManagedPersistMode > EnvManagedPersistRuntimeEnvOnly {
		return fmt.Errorf("unsupported env managed persist mode %d", options.envManagedPersistMode)
	}

	updates := make([]configloader.YAMLScalarUpdate, 0, len(paths))
	for _, path := range paths {
		if options.envManagedPersistMode != EnvManagedPersistForceFile {
			if envName, ok := activeConfigPathEnvName(path); ok {
				return fmt.Errorf("%s is managed by environment variable %s", path, envName)
			}
		}
		if options.envManagedPersistMode == EnvManagedPersistRuntimeEnvOnly {
			if managed, err := m.configPathHasEnvPlaceholder(path); err != nil {
				return err
			} else if managed {
				return missingRuntimeEnvError(path)
			}
		}

		value, err := configValueByMapstructurePath(reflect.ValueOf(newCfg).Elem(), path)
		if err != nil {
			return err
		}
		update, err := yamlScalarUpdateFromConfigValue(path, value)
		if err != nil {
			return err
		}
		updates = append(updates, update)
	}

	var yamlOptions []configloader.YAMLUpdateOption
	if options.envManagedPersistMode == EnvManagedPersistForceFile {
		yamlOptions = append(yamlOptions, configloader.WithEnvPlaceholderOverwrite())
	}
	return configloader.UpdateYAMLScalars(m.configPath, updates, yamlOptions...)
}

func (m *manager) applyRuntimeEnvOnlyPersistPaths(newCfg *Config, paths []string) (map[string]struct{}, bool, error) {
	runtimeOnlyPaths := make(map[string]struct{}, len(paths))
	disabledPaths := disabledConfigPathSet(newCfg.EnvOverride.DisabledPaths)
	root := reflect.ValueOf(newCfg).Elem()
	for _, path := range paths {
		_, alreadyDisabled := disabledPaths[path]
		managed, err := m.configPathIsEnvManaged(path)
		if err != nil {
			return nil, false, err
		}
		if !managed && !alreadyDisabled {
			continue
		}

		envName, raw, ok := activeConfigPathEnv(path)
		if !ok {
			return nil, false, missingRuntimeEnvError(path)
		}
		field, err := configValueByMapstructurePath(root, path)
		if err != nil {
			return nil, false, err
		}
		if !setValueFromEnv(field, raw) {
			return nil, false, fmt.Errorf("%s cannot be set from environment variable %s", path, envName)
		}
		runtimeOnlyPaths[path] = struct{}{}
	}
	metadataChanged := removeDisabledEnvOverridePaths(&newCfg.EnvOverride, runtimeOnlyPaths)
	return runtimeOnlyPaths, metadataChanged, nil
}

func (m *manager) configPathIsEnvManaged(path string) (bool, error) {
	if _, _, ok := activeConfigPathEnv(path); ok {
		return true, nil
	}
	return m.configPathHasEnvPlaceholder(path)
}

func (m *manager) configPathHasEnvPlaceholder(path string) (bool, error) {
	if strings.TrimSpace(m.configPath) == "" {
		return false, fmt.Errorf("configuration file path is not available")
	}
	return configloader.YAMLPathContainsEnvPlaceholder(m.configPath, path)
}

func missingRuntimeEnvError(path string) error {
	candidates := EnvNamesForPath(path)
	if len(candidates) == 0 {
		return fmt.Errorf("%s is managed by environment placeholder but has no environment variable mapping", path)
	}
	return fmt.Errorf("%s is managed by environment placeholder; set one of %s or choose force file persistence", path, strings.Join(candidates, ", "))
}

func removeConfigPaths(paths []string, remove map[string]struct{}) []string {
	if len(remove) == 0 {
		return paths
	}
	kept := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, ok := remove[path]; ok {
			continue
		}
		kept = append(kept, path)
	}
	return kept
}

func addDisabledEnvOverridePaths(cfg *Config, paths []string) bool {
	if cfg == nil {
		return false
	}
	changed := false
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(cfg.EnvOverride.DisabledPaths)+len(paths))
	for _, path := range normalizeConfigPaths(cfg.EnvOverride.DisabledPaths) {
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}
	for _, path := range normalizeConfigPaths(paths) {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
		changed = true
	}
	cfg.EnvOverride.DisabledPaths = normalized
	return changed
}

func removeDisabledEnvOverridePaths(cfg *EnvOverrideConfig, paths map[string]struct{}) bool {
	if cfg == nil || len(paths) == 0 {
		return false
	}
	next := make([]string, 0, len(cfg.DisabledPaths))
	changed := false
	for _, path := range normalizeConfigPaths(cfg.DisabledPaths) {
		if _, ok := paths[path]; ok {
			changed = true
			continue
		}
		next = append(next, path)
	}
	cfg.DisabledPaths = next
	return changed
}

func configValueByMapstructurePath(root reflect.Value, path string) (reflect.Value, error) {
	current := root
	for _, segment := range strings.Split(path, ".") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return reflect.Value{}, fmt.Errorf("invalid config key %q", path)
		}
		if current.Kind() == reflect.Pointer {
			if current.IsNil() {
				return reflect.Value{}, fmt.Errorf("%s is nil", path)
			}
			current = current.Elem()
		}
		if current.Kind() == reflect.Slice || current.Kind() == reflect.Array {
			index, err := strconv.Atoi(segment)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s sequence index must be an integer", path)
			}
			if index < 0 || index >= current.Len() {
				return reflect.Value{}, fmt.Errorf("%s sequence index is out of range", path)
			}
			current = current.Index(index)
			continue
		}
		if current.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("%s is not editable", path)
		}
		field, ok := mapstructureValueField(current, segment)
		if !ok {
			return reflect.Value{}, fmt.Errorf("unknown config key %s", path)
		}
		current = field
	}
	return current, nil
}

func mapstructureValueField(value reflect.Value, segment string) (reflect.Value, bool) {
	valueType := value.Type()
	for index := 0; index < value.NumField(); index++ {
		fieldType := valueType.Field(index)
		tag := strings.Split(fieldType.Tag.Get("mapstructure"), ",")[0]
		if tag == segment {
			return value.Field(index), true
		}
	}
	return reflect.Value{}, false
}

func yamlScalarUpdateFromConfigValue(path string, value reflect.Value) (configloader.YAMLScalarUpdate, error) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return configloader.YAMLScalarUpdate{}, fmt.Errorf("%s is nil", path)
		}
		value = value.Elem()
	}

	update := configloader.YAMLScalarUpdate{Path: path}
	switch value.Kind() {
	case reflect.String:
		update.Kind = configloader.YAMLScalarString
		update.Value = value.String()
	case reflect.Bool:
		update.Kind = configloader.YAMLScalarBool
		update.Value = strconv.FormatBool(value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		update.Kind = configloader.YAMLScalarInt
		update.Value = strconv.FormatInt(value.Int(), 10)
	case reflect.Slice:
		if value.Type().Elem().Kind() != reflect.String {
			return configloader.YAMLScalarUpdate{}, fmt.Errorf("%s is not a persistable scalar value", path)
		}
		update.Kind = configloader.YAMLScalarStringSlice
		update.CreateMissing = path == envOverrideDisabledPathsConfigPath
		update.Values = make([]string, 0, value.Len())
		for index := 0; index < value.Len(); index++ {
			update.Values = append(update.Values, value.Index(index).String())
		}
	default:
		return configloader.YAMLScalarUpdate{}, fmt.Errorf("%s is not a persistable scalar value", path)
	}
	return update, nil
}

func activeConfigPathEnvName(path string) (string, bool) {
	envName, _, ok := activeConfigPathEnv(path)
	return envName, ok
}

func activeConfigPathEnv(path string) (string, string, bool) {
	envName, ok := configPathEnvName(path)
	if !ok {
		return "", "", false
	}
	for _, candidate := range envNameCandidates(envName) {
		if value, ok := os.LookupEnv(candidate); ok && value != "" {
			return candidate, value, true
		}
	}
	return "", "", false
}

// EnvNamesForPath 返回配置路径可使用的环境变量名，按实际覆盖优先级排序。
func EnvNamesForPath(path string) []string {
	envName, ok := configPathEnvName(path)
	if !ok {
		return nil
	}
	return envNameCandidates(envName)
}

func configPathEnvName(path string) (string, bool) {
	field, ok := configStructFieldByMapstructurePath(reflect.TypeOf(Config{}), path)
	if !ok {
		return "", false
	}
	envName := strings.TrimSpace(field.Tag.Get(envNameTag))
	return envName, envName != "" && envName != "-"
}

func configStructFieldByMapstructurePath(root reflect.Type, path string) (reflect.StructField, bool) {
	current := root
	var field reflect.StructField
	for _, segment := range strings.Split(path, ".") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return reflect.StructField{}, false
		}
		if current.Kind() == reflect.Pointer {
			current = current.Elem()
		}
		if current.Kind() == reflect.Slice || current.Kind() == reflect.Array {
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 {
				return reflect.StructField{}, false
			}
			current = current.Elem()
			continue
		}
		if current.Kind() != reflect.Struct {
			return reflect.StructField{}, false
		}
		next, ok := mapstructureTypeField(current, segment)
		if !ok {
			return reflect.StructField{}, false
		}
		field = next
		current = next.Type
	}
	return field, true
}

func mapstructureTypeField(valueType reflect.Type, segment string) (reflect.StructField, bool) {
	for index := 0; index < valueType.NumField(); index++ {
		field := valueType.Field(index)
		tag := strings.Split(field.Tag.Get("mapstructure"), ",")[0]
		if tag == segment {
			return field, true
		}
	}
	return reflect.StructField{}, false
}
