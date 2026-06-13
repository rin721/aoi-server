package config

import "strings"

// UpdateOption 配置 Manager.Update 的可选行为。
type UpdateOption func(*updateOptions)

type updateOptions struct {
	persistPaths []string
}

// WithPersistedPaths 要求 Manager.Update 在验证通过后把指定配置路径写回配置文件。
func WithPersistedPaths(paths ...string) UpdateOption {
	return func(options *updateOptions) {
		options.persistPaths = append(options.persistPaths, paths...)
	}
}

func collectUpdateOptions(options []UpdateOption) updateOptions {
	var collected updateOptions
	for _, option := range options {
		if option != nil {
			option(&collected)
		}
	}
	collected.persistPaths = normalizeConfigPaths(collected.persistPaths)
	return collected
}

func normalizeConfigPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
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
