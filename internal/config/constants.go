package config

// 本文件属于配置子系统，处理配置加载、环境变量覆盖、运行时快照或跨分区校验。

import (
	"strings"
	"unicode"

	app "github.com/rei0721/go-scaffold/types/constants"
)

// EnvPrefix returns the dynamic prefix used by configuration env overrides.
//
// Rules:
//   - derive the application token from app.AppPrefix
//   - normalize it to uppercase underscore-separated env-token format
//   - configuration override names use <APP_PREFIX>_APP_<MODULE>_<FIELD>
//
// Example: app.AppPrefix = "Rin" returns "RIN_APP".
func EnvPrefix() string {
	prefix := normalizeEnvToken(app.AppPrefix)
	if prefix == "" {
		return "APP"
	}
	return prefix + "_APP"
}

// EnvConfigPathName returns the dynamic env name for the config file path.
//
// Example: app.AppPrefix = "Rin" returns "RIN_CONFIG_PATH".
func EnvConfigPathName() string {
	prefix := normalizeEnvToken(app.AppPrefix)
	if prefix == "" {
		return "CONFIG_PATH"
	}
	return prefix + "_CONFIG_PATH"
}

// EnvPrefixJoin joins an envname tag value with the dynamic application prefix.
func EnvPrefixJoin(field string) string {
	field = strings.TrimSpace(field)
	prefix := EnvPrefix()
	if field == "" {
		return prefix
	}
	if prefix == "" || strings.HasPrefix(field, prefix+"_") {
		return field
	}
	return prefix + "_" + field
}

// normalizeEnvToken 将动态应用前缀规整为环境变量安全片段，避免大小写和非法字符影响 env key 拼接。
func normalizeEnvToken(value string) string {
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToUpper(r))
			lastUnderscore = false
			continue
		}
		if !lastUnderscore && builder.Len() > 0 {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

const (
	// EnvFilePath 定义配置管理器自动加载的本地环境变量文件名称。
	EnvFilePath = ".env"
	// EnvFilePathExample 定义环境变量示例文件名称，用于文档和初始化模板。
	EnvFilePathExample = ".env.example"
)

const (
	// DefaultSeparator 定义环境变量中切片配置的默认分隔符。
	DefaultSeparator = ","
)

const (
	// AppServerName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppServerName = "server"
	// AppDatabaseName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppDatabaseName = "database"
	// AppRedisName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppRedisName = "redis"
	// AppLoggerName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppLoggerName = "logger"
	// AppI18nName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppI18nName = "i18n"
	// AppExecutorName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppExecutorName = "executor"
	// AppStorageName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppStorageName = "storage"
	// AppDemoName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppDemoName = "demo"
	// AppCORSName 定义聚合配置中的分区名称，供加载、校验和热更新差异判断复用。
	AppCORSName      = "cors"
	AppRPCName       = "rpc"
	AppAuthName      = "auth"
	AppMigrationName = "migration"
	AppWebUIName     = "webui"
)
