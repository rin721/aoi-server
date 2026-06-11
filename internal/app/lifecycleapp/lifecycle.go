// Package lifecycleapp 管理应用传输层启动和资源关闭顺序。
package lifecycleapp

// 本文件定义应用运行期生命周期编排，统一处理 HTTP 服务启动、资源关闭和日志同步。

import (
	"context"
	"fmt"
	"time"

	"github.com/rei0721/go-scaffold/internal/app/initapp"
)

// Start 启动 HTTP server。
//
// 传输层必须已完成初始化；nil HTTPServer 表示装配阶段失败或被错误跳过，应立即返回错误。
func Start(ctx context.Context, transport initapp.Transport) error {
	if transport.HTTPServer == nil {
		return fmt.Errorf("http server is not initialized")
	}
	if err := transport.HTTPServer.Start(ctx); err != nil {
		return fmt.Errorf("start HTTP server: %w", err)
	}
	if transport.RPCServer != nil {
		if err := transport.RPCServer.Start(ctx); err != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = transport.HTTPServer.Shutdown(shutdownCtx)
			return fmt.Errorf("start RPC server: %w", err)
		}
	}
	return nil
}

// Shutdown 按固定顺序释放应用资源。
//
// 关闭过程采用最佳努力策略：某个资源关闭失败不会阻断后续资源释放，最终返回错误数量汇总。
func Shutdown(ctx context.Context, core initapp.Core, infra initapp.Infrastructure, transport initapp.Transport) error {
	log := core.Logger
	if log != nil {
		log.Info("shutting down application...")
	}

	var errs []error

	if transport.HTTPServer != nil {
		if err := transport.HTTPServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("http server shutdown: %w", err))
			if log != nil {
				log.Error("failed to shutdown HTTP server", "error", err)
			}
		} else if log != nil {
			log.Info("HTTP server stopped")
		}
	}

	if transport.RPCServer != nil {
		if err := transport.RPCServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("rpc server shutdown: %w", err))
			if log != nil {
				log.Error("failed to shutdown RPC server", "error", err)
			}
		} else if log != nil {
			log.Info("RPC server stopped")
		}
	}

	if infra.Storage != nil {
		if err := infra.Storage.Close(); err != nil {
			errs = append(errs, fmt.Errorf("storage close: %w", err))
			if log != nil {
				log.Error("failed to close storage", "error", err)
			}
		} else if log != nil {
			log.Info("storage closed")
		}
	}

	if infra.Executor != nil {
		infra.Executor.Shutdown()
		if log != nil {
			log.Info("executor stopped")
		}
	}

	if infra.Cache != nil {
		if err := infra.Cache.Close(); err != nil {
			errs = append(errs, fmt.Errorf("cache close: %w", err))
			if log != nil {
				log.Error("failed to close cache", "error", err)
			}
		} else if log != nil {
			log.Info("cache closed")
		}
	}

	if infra.Database != nil {
		if err := infra.Database.Close(); err != nil {
			errs = append(errs, fmt.Errorf("database close: %w", err))
			if log != nil {
				log.Error("failed to close database", "error", err)
			}
		} else if log != nil {
			log.Info("database connection closed")
		}
	}

	if log != nil {
		_ = log.Sync()
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(errs))
	}
	if log != nil {
		log.Info("application shutdown complete")
	}
	return nil
}
