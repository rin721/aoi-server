package constants

// 本文件定义跨层共享常量，避免内部基础设施包反向污染公共 types 边界。

import "time"

const (
	// AppDefaultConfigPath 定义未显式传参时应用读取的默认配置文件路径。
	AppDefaultConfigPath = "configs/config.yaml"
	// AppShutdownTimeout 定义进程收到退出信号后的优雅关闭最长等待时间。
	AppShutdownTimeout = 30 * time.Second
	// EnvConfigPathName 定义覆盖配置文件路径的环境变量名称。
	EnvConfigPathName = "RIN_CONFIG_PATH"

	// AppPrefix 定义运行时环境变量前缀，配置包会基于它构造动态覆盖键。
	AppPrefix = "Rin"
	// AppName 定义 CLI 与元信息输出中的应用名称。
	AppName = "go-scaffold"
	// AppDescription 定义 CLI 帮助信息中的应用描述。
	AppDescription = "This is a go backend scaffold"
	// AppServerCommandName 定义启动服务子命令的稳定名称。
	AppServerCommandName = "server"
	// AppVersion 定义当前脚手架二进制暴露的版本号。
	AppVersion = "0.1.2"
)
