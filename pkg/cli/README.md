# pkg/cli

`pkg/cli` 是项目级终端工具包。它用 Cobra 负责命令路由和 flag 解析，
用 Charm Bubble Tea/Lip Gloss v2 提供默认交互式终端首页。

本包对外暴露项目自己的 `CommandSpec`、`FlagSpec` 和 `Context`，不要求业务层
直接依赖 Cobra 命令对象。底层库细节、I/O 注入、错误映射、帮助输出和默认
TUI 首页都收敛在 `pkg/cli` 内部。

## 基本用法

```go
app, err := cli.NewApp(cli.Config{
    Name:        "mytool",
    Version:     "1.0.0",
    Description: "我的 CLI 工具",
})
if err != nil {
    return err
}

err = app.AddCommand(cli.CommandSpec{
    Name:        "generate",
    Description: "生成代码",
    Flags: []cli.FlagSpec{
        {
            Name:        "model",
            ShortName:   "m",
            Type:        cli.FlagTypeString,
            Required:    true,
            Description: "模型名称",
        },
        {
            Name:        "output",
            ShortName:   "o",
            Type:        cli.FlagTypeString,
            Default:     "./models",
            EnvVar:      "OUTPUT_DIR",
            Description: "输出目录",
        },
    },
    Run: func(ctx *cli.Context) error {
        fmt.Fprintf(ctx.Stdout, "生成 %s\n", ctx.GetString("model"))
        return nil
    },
})
if err != nil {
    return err
}

if err := app.Run(context.Background(), os.Args[1:]); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(cli.GetExitCode(err))
}
```

## 运行行为

- `app.Run(ctx, nil)` 默认启动内置 Bubble Tea 交互式首页。
- `--help`、`help`、命令路由和 flag 解析由 Cobra 处理。
- 首页展示已注册命令，按 `enter` 查看当前命令的帮助信息。
- 首页按 `q`、`esc` 或 `ctrl+c` 退出，并返回 `CancelledError`。
- 如需空参数时显示普通 Cobra help，可设置 `Config.DisableInteractiveHome`。

## Flag 类型

当前支持的 flag 类型：

```go
cli.FlagTypeString
cli.FlagTypeInt
cli.FlagTypeBool
cli.FlagTypeStringSlice
```

`FlagSpec.EnvVar` 会在运行时解析；当环境变量存在时，它会作为该 flag 的默认值。

## 错误和退出码

本包保留稳定的进程退出语义：

| 错误类型 | 退出码 |
| --- | --- |
| `UsageError` | `2` |
| `CommandError` | `1` |
| `CancelledError` | `130` |

进程边界统一使用 `cli.GetExitCode(err)` 提取退出码。
