package initapp

// 本文件属于应用初始化装配层，负责把配置、基础设施、业务模块或传输层拼接为可运行的分层对象。

import (
	"fmt"

	"github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/internal/middleware"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	pluginhandler "github.com/rei0721/go-scaffold/internal/modules/plugins/handler"
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	httptransport "github.com/rei0721/go-scaffold/internal/transport/http"
	rpctransport "github.com/rei0721/go-scaffold/internal/transport/rpc"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/httpserver"
	"github.com/rei0721/go-scaffold/pkg/i18n"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/rpcserver"
	"github.com/rei0721/go-scaffold/pkg/web"
)

// NewTransport 装配 HTTP 传输层。
//
// Demo handler 可能为 nil，路由层据此决定是否注册 Demo Todo 接口。
func NewTransport(core Core, infra Infrastructure, modules Modules) (Transport, error) {
	corsConfig, err := NewCORS(core.Config, core.Logger)
	if err != nil {
		return Transport{}, err
	}

	router, server, err := NewHTTPServer(
		core.Config,
		core.Logger,
		core.I18n,
		infra.Database,
		corsConfig,
		modules.Demo.TodoHandler,
		modules.Demo.CustomerHandler,
		modules.IAM.Handler,
		modules.Plugins.Handler,
		modules.System.Handler,
		modules.IAM.Service,
	)
	if err != nil {
		return Transport{}, err
	}

	rpcServer, err := NewRPCServer(core.Config, core.Logger)
	if err != nil {
		return Transport{}, err
	}

	return Transport{
		Router:     router,
		HTTPServer: server,
		RPCServer:  rpcServer,
	}, nil
}

// NewCORS 生成中间件使用的 CORS 配置。
//
// 应用配置会先补默认值再应用环境覆盖，最后才校验并转换为 middleware 包结构。
func NewCORS(cfg *config.Config, log logger.Logger) (middleware.CORSConfig, error) {
	corsCfg := cfg.CORS
	corsCfg.DefaultConfig()
	corsCfg.OverrideConfig()

	if err := corsCfg.Validate(); err != nil {
		return middleware.CORSConfig{}, err
	}

	if corsCfg.Enabled {
		log.Info(
			"CORS middleware enabled",
			"allow_origins", corsCfg.AllowOrigins,
			"allow_credentials", corsCfg.AllowCredentials,
			"max_age", corsCfg.MaxAge,
		)
	} else {
		log.Info("CORS middleware disabled")
	}

	return middleware.CORSConfig{
		Enabled:          corsCfg.Enabled,
		AllowOrigins:     corsCfg.AllowOrigins,
		AllowMethods:     corsCfg.AllowMethods,
		AllowHeaders:     corsCfg.AllowHeaders,
		ExposeHeaders:    corsCfg.ExposeHeaders,
		AllowCredentials: corsCfg.AllowCredentials,
		MaxAge:           corsCfg.MaxAge,
	}, nil
}

// NewHTTPServer 创建 Gin router 和 HTTP server 包装器。
//
// gin.SetMode 会修改进程级全局状态，因此必须在创建 router 前完成。
func NewHTTPServer(
	cfg *config.Config,
	log logger.Logger,
	i18nApp i18n.I18n,
	db database.Database,
	corsConfig middleware.CORSConfig,
	todoHandler *demohandler.TodoHandler,
	customerHandler *demohandler.CustomerHandler,
	iamHandler *iamhandler.Handler,
	pluginHandler *pluginhandler.Handler,
	systemHandler *systemhandler.Handler,
	iamService iamservice.Service,
) (*web.Engine, httpserver.HTTPServer, error) {
	middlewareCfg := middleware.DefaultMiddlewareConfig()
	middlewareCfg.CORS = corsConfig
	webUICfg := cfg.WebUI
	webUICfg.ApplyDefaults()

	router := httptransport.NewRouter(httptransport.RouterDeps{
		Mode:            cfg.Server.Mode,
		Logger:          log,
		I18n:            i18nApp,
		Database:        db,
		Middleware:      middlewareCfg,
		TodoHandler:     todoHandler,
		CustomerHandler: customerHandler,
		IAMHandler:      iamHandler,
		PluginHandler:   pluginHandler,
		SystemHandler:   systemHandler,
		IAMAuth:         iamService,
		IAMAuthz:        iamService,
		WebUI: httptransport.WebUIDeps{
			Enabled:   webUICfg.EnabledValue(),
			MountPath: webUICfg.MountPath,
			DistDir:   webUICfg.DistDir,
		},
	})

	server, err := httpserver.New(router, HTTPServerConfig(cfg), log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http server: %w", err)
	}

	return router, server, nil
}

// NewRPCServer 创建 JSON-RPC 独立端口服务。
func NewRPCServer(cfg *config.Config, log logger.Logger) (rpcserver.Server, error) {
	registry, err := rpctransport.NewRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create rpc registry: %w", err)
	}

	server, err := rpcserver.New(registry, RPCServerConfig(cfg), log)
	if err != nil {
		return nil, fmt.Errorf("failed to create rpc server: %w", err)
	}
	return server, nil
}
