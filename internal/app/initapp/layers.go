// Package initapp 定义应用装配所需的分层结构和构建函数。
//
// 本包位于组合根内部，负责把配置转换为可注入的核心服务、基础设施、业务模块和传输层。
package initapp

// 本文件属于应用初始化装配层，负责把配置、基础设施、业务模块或传输层拼接为可运行的分层对象。

import (
	"github.com/rei0721/go-scaffold/internal/config"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	demorepository "github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	demoservice "github.com/rei0721/go-scaffold/internal/modules/demo/service"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	iamrepository "github.com/rei0721/go-scaffold/internal/modules/iam/repository"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	pluginhandler "github.com/rei0721/go-scaffold/internal/modules/plugins/handler"
	pluginservice "github.com/rei0721/go-scaffold/internal/modules/plugins/service"
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
	"github.com/rei0721/go-scaffold/pkg/cache"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/executor"
	"github.com/rei0721/go-scaffold/pkg/httpserver"
	"github.com/rei0721/go-scaffold/pkg/i18n"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/rpcserver"
	"github.com/rei0721/go-scaffold/pkg/storage"
	"github.com/rei0721/go-scaffold/pkg/utils"
	"github.com/rei0721/go-scaffold/pkg/web"
)

// Core 保存所有层共享的核心服务。
//
// Core 是后续装配的输入边界：基础设施、模块和传输层只能依赖这里暴露的跨层能力。
type Core struct {
	Config        *config.Config
	ConfigManager config.Manager
	Logger        logger.Logger
	I18n          i18n.I18n
	I18nUtils     *utils.I18nUtils
	IDGenerator   utils.IDGenerator
}

// Infrastructure 保存可被业务模块和传输层复用的基础设施组件。
//
// Database 是启动期硬依赖；Cache、Executor 和 Storage 可能因配置禁用而为 nil，
// 调用方必须把 nil 视为“该能力未启用”。
type Infrastructure struct {
	Database database.Database
	Cache    cache.Cache
	Executor executor.Manager
	Storage  storage.Storage
}

// Modules 汇总当前应用启用的业务模块。
type Modules struct {
	Demo    DemoModule
	IAM     IAMModule
	Plugins PluginsModule
	System  SystemModule
}

// DemoModule 保存 Demo Todo 模块的三层对象。
//
// 当 Demo 被配置禁用时，本结构体保持零值，路由层应据此跳过 Demo 接口注册。
type DemoModule struct {
	TodoRepository demorepository.TodoRepository
	TodoService    demoservice.TodoService
	TodoHandler    *demohandler.TodoHandler
}

type IAMModule struct {
	Repository iamrepository.Repository
	Service    iamservice.Service
	Handler    *iamhandler.Handler
}

type PluginsModule struct {
	Service pluginservice.Service
	Handler *pluginhandler.Handler
}

type SystemModule struct {
	Service systemservice.Service
	Handler *systemhandler.Handler
}

// Transport 保存对外服务入口。
type Transport struct {
	Router     *web.Engine
	HTTPServer httpserver.HTTPServer
	RPCServer  rpcserver.Server
}
