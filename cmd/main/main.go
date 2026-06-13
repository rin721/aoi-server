// Package main is the command entry point for the scaffold service.
package main

// 本文件是编译后二进制的进程入口，只负责装配 CLI 应用并把退出语义交给命令层处理。

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

// main 创建 CLI 应用并注册顶层命令，是进程退出码和标准输出错误提示的最后边界。
func main() {
	if err := runCLI(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.GetExitCode(err))
	}
}

func newCLIApp() (cli.App, error) {
	app, err := cli.NewApp(cli.Config{
		Name:        constants.AppName,
		Version:     constants.AppVersion,
		Description: constants.AppDescription,
	})
	if err != nil {
		return nil, err
	}

	serverSpec := NewAppCommand().Spec()
	serverSpec.HomeHidden = true
	if err := app.AddCommand(serverSpec); err != nil {
		return nil, err
	}
	dbSpec := NewDBCommand().Spec()
	dbSpec.HomeHidden = true
	if err := app.AddCommand(dbSpec); err != nil {
		return nil, err
	}
	iamSpec := NewIAMCommand().Spec()
	iamSpec.HomeHidden = true
	if err := app.AddCommand(iamSpec); err != nil {
		return nil, err
	}
	for _, spec := range NewSystemCenterCommands() {
		if err := app.AddCommand(spec); err != nil {
			return nil, err
		}
	}
	return app, nil
}

func runCLI(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	app, err := newCLIApp()
	if err != nil {
		return err
	}
	return app.RunWithIO(ctx, args, stdin, stdout, stderr)
}
