package cli

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"
)

// ProgramOption 定制底层 Bubble Tea 程序。
type ProgramOption = tea.ProgramOption

// Config 描述一个 CLI 应用。
type Config struct {
	// Name 是应用名称，用于 Cobra 根命令和 TUI 首页标题。
	Name string
	// Version 是应用版本号，用于 --version 和首页标题。
	Version string
	// Description 是应用描述，用于 help 和首页副标题。
	Description string

	// Stdin、Stdout、Stderr 允许调用方注入 I/O，便于测试和嵌入式运行。
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Theme 控制默认 TUI 首页样式；为空时使用 DefaultTheme。
	Theme *Theme
	// ProgramOptions 会透传给底层 Bubble Tea 程序。
	ProgramOptions []ProgramOption

	// DisableInteractiveHome 在空参数时禁用交互式首页，恢复为 Cobra help。
	DisableInteractiveHome bool
}

// App 是公开 CLI 边界，具体实现隐藏 Cobra 和 Bubble Tea。
type App interface {
	Name() string
	Version() string
	Description() string
	AddCommand(CommandSpec) error
	Run(context.Context, []string) error
	RunWithIO(context.Context, []string, io.Reader, io.Writer, io.Writer) error
}

// CommandFunc 处理已完成参数解析的命令调用。
type CommandFunc func(*Context) error

// ArgsValidator 在 flag 解析完成后校验位置参数。
type ArgsValidator func(*Context) error

// CommandSpec 声明一个命令，同时避免向调用方暴露 Cobra 细节。
type CommandSpec struct {
	// Name 是命令的唯一注册名。
	Name string
	// Use 覆盖 Cobra 的 usage 行；为空时使用 Name。
	Use string
	// Aliases 是命令别名。
	Aliases []string
	// Description 是命令短描述，用于列表和 help。
	Description string
	// Long 是命令长描述；为空时回退到 Description。
	Long string
	// Example 是命令示例文本。
	Example string

	// Flags 声明命令支持的 flag。
	Flags []FlagSpec
	// Args 校验位置参数；为空时不做额外校验。
	Args ArgsValidator
	// Run 是命令执行函数；为空时只输出该命令 help。
	Run CommandFunc
	// Commands 声明子命令。
	Commands []CommandSpec
}

// FlagType 标识支持的 flag 值类型。
type FlagType int

const (
	FlagTypeString FlagType = iota
	FlagTypeInt
	FlagTypeBool
	FlagTypeStringSlice
)

// FlagSpec 声明一个命令行 flag。
type FlagSpec struct {
	// Name 是长 flag 名称。
	Name string
	// ShortName 是短 flag 名称。
	ShortName string
	// Shorthand 是 ShortName 的兼容别名，优先级更高。
	Shorthand string
	// Type 指定 flag 值类型。
	Type FlagType
	// Required 表示该 flag 是否必填。
	Required bool
	// Default 是 flag 默认值。
	Default interface{}
	// Description 是 help 中展示的 flag 描述。
	Description string
	// EnvVar 指定环境变量回退来源。
	EnvVar string
}

// Context 是传递给命令处理函数的执行上下文，包含参数、flag 和 I/O。
type Context struct {
	context.Context

	// CommandName 是当前执行的命令名。
	CommandName string
	// CommandPath 是包含父命令的完整命令路径。
	CommandPath string
	// Args 是解析 flag 后剩余的位置参数。
	Args []string
	// Flags 是解析后的 flag 值集合。
	Flags map[string]interface{}

	// Stdin、Stdout、Stderr 是本次命令调用使用的 I/O。
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// GetString 返回字符串 flag 值；不存在或类型不匹配时返回空字符串。
func (c *Context) GetString(name string) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Flags[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt 返回整数 flag 值；不存在或类型不匹配时返回零。
func (c *Context) GetInt(name string) int {
	if c == nil {
		return 0
	}
	if v, ok := c.Flags[name]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool 返回布尔 flag 值；不存在或类型不匹配时返回 false。
func (c *Context) GetBool(name string) bool {
	if c == nil {
		return false
	}
	if v, ok := c.Flags[name]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetStringSlice 返回字符串切片 flag 值；不存在或类型不匹配时返回 nil。
func (c *Context) GetStringSlice(name string) []string {
	if c == nil {
		return nil
	}
	if v, ok := c.Flags[name]; ok {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}
