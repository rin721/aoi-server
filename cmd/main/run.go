package main

// 本文件封装主应用启动流程，把 CLI 参数解析、应用构建、信号监听和 shutdown 超时串成进程生命周期。

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rei0721/go-scaffold/internal/app"
	"github.com/rei0721/go-scaffold/types/constants"
)

// runApp 装配应用、启动 HTTP 服务并等待系统信号或启动错误，最终按统一超时执行优雅关闭。
func runApp(configPath string) {
	application, err := app.New(app.Options{
		ConfigPath: configPath,
	})
	if err != nil {
		os.Stderr.WriteString("failed to initialize application: " + err.Error() + "\n")
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := application.Run(); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-quit:
		application.Core.Logger.Info("received shutdown signal", "signal", sig.String())
	case err := <-errChan:
		application.Core.Logger.Error("server error", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.AppShutdownTimeout)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		application.Core.Logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	application.Core.Logger.Info("application exited gracefully")
}
