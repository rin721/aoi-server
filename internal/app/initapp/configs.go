package initapp

// 本文件属于应用初始化装配层，负责把配置、基础设施、业务模块或传输层拼接为可运行的分层对象。

import (
	"reflect"
	"time"

	"github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cache"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/executor"
	"github.com/rei0721/go-scaffold/pkg/httpserver"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/rpcserver"
	"github.com/rei0721/go-scaffold/pkg/storage"
)

// IsRedisConfigChanged 判断 Redis 配置是否发生变化。
func IsRedisConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.Redis != newCfg.Redis
}

// IsDatabaseConfigChanged 判断数据库配置是否发生变化。
func IsDatabaseConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.Database != newCfg.Database
}

// IsServerConfigChanged 判断 HTTP server 配置是否发生变化。
func IsServerConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.Server != newCfg.Server
}

// IsRPCConfigChanged 判断 RPC server 配置是否发生变化。
func IsRPCConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.RPC != newCfg.RPC
}

// IsLoggerConfigChanged 判断日志配置是否发生变化。
func IsLoggerConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.Logger != newCfg.Logger
}

// IsExecutorConfigChanged 判断执行器配置是否发生变化。
//
// Pools 是切片字段，不能直接比较，因此使用 DeepEqual 保证池数量和每个池参数都纳入 reload 判断。
func IsExecutorConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.Executor.Enabled != newCfg.Executor.Enabled ||
		!reflect.DeepEqual(oldCfg.Executor.Pools, newCfg.Executor.Pools)
}

// IsStorageConfigChanged 判断存储配置是否发生变化。
func IsStorageConfigChanged(oldCfg, newCfg *config.Config) bool {
	if oldCfg == newCfg {
		return false
	}
	return oldCfg.Storage != newCfg.Storage
}

// RedisCacheConfig 将应用 Redis 配置转换为缓存包配置。
//
// 应用配置中的超时单位是秒，传给底层包前必须转换为 time.Duration。
func RedisCacheConfig(cfg *config.Config) *cache.Config {
	return &cache.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,
	}
}

// DatabaseConfig 将应用数据库配置转换为 database 包配置。
func DatabaseConfig(cfg *config.Config) *database.Config {
	return &database.Config{
		Driver:       database.Driver(cfg.Database.Driver),
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}
}

// LoggerConfig 将应用日志配置转换为 logger 包配置。
func LoggerConfig(cfg *config.Config) *logger.Config {
	return &logger.Config{
		Level:         cfg.Logger.Level,
		Format:        cfg.Logger.Format,
		ConsoleFormat: cfg.Logger.ConsoleFormat,
		FileFormat:    cfg.Logger.FileFormat,
		Output:        cfg.Logger.Output,
		FilePath:      cfg.Logger.FilePath,
		MaxSize:       cfg.Logger.MaxSize,
		MaxBackups:    cfg.Logger.MaxBackups,
		MaxAge:        cfg.Logger.MaxAge,
	}
}

// ExecutorConfigs 将应用执行器池配置转换为 executor 包配置列表。
//
// 每个池的 Expiry 同样以秒为单位配置，进入执行器前统一转换为 time.Duration。
func ExecutorConfigs(cfg *config.Config) []executor.Config {
	configs := make([]executor.Config, 0, len(cfg.Executor.Pools))
	for _, poolCfg := range cfg.Executor.Pools {
		configs = append(configs, executor.Config{
			Name:        executor.PoolName(poolCfg.Name),
			Size:        poolCfg.Size,
			Expiry:      time.Duration(poolCfg.Expiry) * time.Second,
			NonBlocking: poolCfg.NonBlocking,
		})
	}
	return configs
}

// HTTPServerConfig 将应用 server 配置转换为 httpserver 包配置。
func HTTPServerConfig(cfg *config.Config) *httpserver.Config {
	return &httpserver.Config{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
}

// RPCServerConfig 将应用 rpc 配置转换为 rpcserver 包配置。
func RPCServerConfig(cfg *config.Config) *rpcserver.Config {
	rpcCfg := cfg.RPC
	rpcCfg.ApplyDefaults()
	return &rpcserver.Config{
		Enabled:      rpcCfg.Enabled,
		Host:         rpcCfg.Host,
		Port:         rpcCfg.Port,
		ReadTimeout:  time.Duration(rpcCfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(rpcCfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(rpcCfg.IdleTimeout) * time.Second,
	}
}

// NormalizedStorageConfig 返回已补默认值并应用环境覆盖的存储配置。
func NormalizedStorageConfig(cfg *config.Config) config.StorageConfig {
	storageCfg := cfg.Storage
	storageCfg.DefaultConfig()
	storageCfg.OverrideConfig()
	return storageCfg
}

// StorageConfig 将标准化后的存储配置转换为 storage 包配置。
func StorageConfig(cfg *config.Config) *storage.Config {
	storageCfg := NormalizedStorageConfig(cfg)
	return storageCfg.ToPkgConfig()
}
