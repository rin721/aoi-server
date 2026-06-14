package initapp

import (
	"fmt"

	"github.com/rei0721/go-scaffold/internal/app/adapters"
	"github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/internal/middleware"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	pluginhandler "github.com/rei0721/go-scaffold/internal/modules/plugins/handler"
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	"github.com/rei0721/go-scaffold/internal/ports"
	httptransport "github.com/rei0721/go-scaffold/internal/transport/http"
	rpctransport "github.com/rei0721/go-scaffold/internal/transport/rpc"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/httpserver"
	"github.com/rei0721/go-scaffold/pkg/i18n"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/rpcserver"
	"github.com/rei0721/go-scaffold/pkg/web"
)

// NewTransport 装配 HTTP 和 RPC 传输入口。
func NewTransport(core Core, infra Infrastructure, modules Modules) (Transport, error) {
	corsConfig, err := NewCORS(core.Config, core.Logger)
	if err != nil {
		return Transport{}, err
	}
	attachWebInitialSetupService(core, infra, modules)

	router, server, err := NewHTTPServer(
		core.Config,
		core.Logger,
		core.I18n,
		infra.Database,
		core.IDGenerator,
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

// NewHTTPServer 创建 HTTP router 和 HTTP server 包装器。
func NewHTTPServer(
	cfg *config.Config,
	log logger.Logger,
	i18nApp i18n.I18n,
	db database.Database,
	traceIDGenerator ports.IDGenerator,
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

	engine := web.New(cfg.Server.Mode)
	router := adapters.NewHTTPEngine(engine)
	httptransport.NewRouter(httptransport.RouterDeps{
		Router:           router,
		RouteLister:      router,
		StaticSPA:        router,
		Logger:           log,
		I18n:             i18nApp,
		Database:         adapters.NewDatabase(db),
		TraceIDGenerator: traceIDGenerator,
		Middleware:       middlewareCfg,
		TodoHandler:      todoHandler,
		CustomerHandler:  customerHandler,
		IAMHandler:       iamHandler,
		PluginHandler:    pluginHandler,
		SystemHandler:    systemHandler,
		IAMAuth:          iamService,
		IAMAuthz:         iamService,
		WebUI: httptransport.WebUIDeps{
			Enabled:   webUICfg.EnabledValue(),
			MountPath: webUICfg.MountPath,
			DistDir:   webUICfg.DistDir,
		},
	})

	server, err := httpserver.New(engine, HTTPServerConfig(cfg), log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http server: %w", err)
	}

	return engine, server, nil
}

// NewRPCServer 创建 JSON-RPC 独立端口服务。
func NewRPCServer(cfg *config.Config, log logger.Logger) (rpcserver.Server, error) {
	registry := rpcserver.NewRegistry()
	if err := rpctransport.Register(adapters.NewRPCRegistry(registry)); err != nil {
		return nil, fmt.Errorf("failed to create rpc registry: %w", err)
	}

	server, err := rpcserver.New(registry, RPCServerConfig(cfg), log)
	if err != nil {
		return nil, fmt.Errorf("failed to create rpc server: %w", err)
	}
	return server, nil
}
