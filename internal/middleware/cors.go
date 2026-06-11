package middleware

// 本文件定义 Gin 中间件能力，约束请求进入业务 handler 前后的链路上下文、副作用和错误输出。

import "github.com/rei0721/go-scaffold/pkg/web"

// CORSMiddleware 返回 CORS 中间件
// 基于配置处理跨域资源共享(CORS)
// 参数:
//
//	cfg: CORS 配置,包含允许的源、方法、请求头等
//
// 返回:
//
//	web.HandlerFunc: CORS 中间件
//
// 功能:
//  1. 处理 OPTIONS 预检请求
//  2. 设置 CORS 相关响应头
//  3. 支持配置化的源匹配(精确匹配和通配符)
//
// 使用场景:
//
//	当前端应用和后端 API 不在同一域名时,需要启用 CORS
//	例如: 前端在 http://localhost:3000, 后端在 http://localhost:8080
//
// 中间件顺序:
//
//	建议在 TraceID 之后,Logger 之前注册
//	这样可以确保预检请求也能被正确追踪和记录
func CORSMiddleware(cfg CORSConfig) web.HandlerFunc {
	return web.CORS(web.CORSConfig{
		Enabled:          cfg.Enabled,
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
