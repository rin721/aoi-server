package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/rei0721/go-scaffold/pkg/configloader"
)

func (m *manager) persistConfigUpdate(newCfg *Config, paths []string) error {
	if strings.TrimSpace(m.configPath) == "" {
		return fmt.Errorf("configuration file path is not available")
	}

	updates := make([]configloader.YAMLScalarUpdate, 0, len(paths))
	for _, path := range paths {
		if envName, ok := activeConfigPathEnvName(path); ok {
			return fmt.Errorf("%s is managed by environment variable %s", path, envName)
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

	return configloader.UpdateYAMLScalars(m.configPath, updates)
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
	envName, ok := configPathEnvName(path)
	if !ok {
		return "", false
	}
	for _, candidate := range envNameCandidates(envName) {
		if value, ok := os.LookupEnv(candidate); ok && value != "" {
			return candidate, true
		}
	}
	return "", false
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
