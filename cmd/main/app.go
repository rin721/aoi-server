package main

// 本文件定义 server 子命令，是 CLI 层进入应用组装、启动和优雅关闭链路的边界。

import (
	"fmt"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

// AppCommand 表示启动 HTTP 服务的 CLI 子命令，是命令层进入应用生命周期的主入口。
type AppCommand struct{}

// NewAppCommand 创建 server 子命令实例，保持命令注册阶段无运行时副作用。
func NewAppCommand() *AppCommand {
	return &AppCommand{}
}

// Name 返回 server 子命令在 CLI 注册表中的唯一名称。
func (c *AppCommand) Name() string {
	return constants.AppServerCommandName
}

// Description 返回 server 子命令的短描述，用于帮助输出和命令索引。
func (c *AppCommand) Description() string {
	return "Run server"
}

// Usage 返回 server 子命令的调用格式，说明可接受的配置路径参数。
func (c *AppCommand) Usage() string {
	return fmt.Sprintf("%s [--config=<name>]", constants.AppServerCommandName)
}

// Flags 声明 server 子命令接受的配置文件参数，默认值来自全局应用常量。
func (c *AppCommand) Flags() []cli.FlagSpec {
	return []cli.FlagSpec{
		{
			Name:        "config",
			ShortName:   "c",
			Type:        cli.FlagTypeString,
			Required:    false,
			Default:     constants.AppDefaultConfigPath,
			Description: "Config file path",
			EnvVar:      appconfig.EnvConfigPathName(),
		},
	}
}

// Spec 将 server 命令转换为 pkg/cli 的新声明式命令规格。
func (c *AppCommand) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        c.Name(),
		Use:         c.Usage(),
		Description: c.Description(),
		Flags:       c.Flags(),
		Run:         c.Execute,
	}
}

// Execute 解析 server 子命令参数并进入应用启动流程，错误会原样返回给 CLI 框架。
func (c *AppCommand) Execute(ctx *cli.Context) error {
	runApp(ctx.GetString("config"))
	return nil
}
