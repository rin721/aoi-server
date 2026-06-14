package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rei0721/go-scaffold/internal/ports"
)

// CORSMiddleware 返回基于内部 HTTP 端口的 CORS 中间件。
func CORSMiddleware(cfg CORSConfig) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(origin, cfg.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}
		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		if c.Method() == http.MethodOptions {
			if len(cfg.AllowMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			}
			if len(cfg.AllowHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
			}
			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}
			c.Data(http.StatusNoContent, "text/plain; charset=utf-8", nil)
			return
		}

		c.Next()
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, item := range allowed {
		item = strings.TrimSpace(item)
		if item == "*" || item == origin {
			return true
		}
	}
	return false
}
