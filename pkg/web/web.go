package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	urlpath "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var ErrStaticSPAIndexMissing = errors.New("static spa index.html missing")

// Context is the HTTP request boundary exposed to internal layers.
type Context interface {
	Request() *http.Request
	RequestContext() context.Context
	GetHeader(name string) string
	Header(name, value string)
	Set(key string, value any)
	Get(key any) (any, bool)
	Param(name string) string
	BindJSON(dest any) error
	JSON(status int, body any)
	AbortWithStatusJSON(status int, body any)
	Next()
	Path() string
	Method() string
	ClientIP() string
	Status() int
}

// HandlerFunc is a transport handler that is independent from the underlying router.
type HandlerFunc func(Context)

// Router is the route registration surface used by internal transport code.
type Router interface {
	Use(...HandlerFunc)
	GET(string, HandlerFunc)
	POST(string, HandlerFunc)
	PATCH(string, HandlerFunc)
	PUT(string, HandlerFunc)
	DELETE(string, HandlerFunc)
	ANY(string, HandlerFunc)
	Group(string) Router
}

// Engine wraps the underlying HTTP router while exposing only project-owned types.
type Engine struct {
	engine *gin.Engine
}

// RouteInfo exposes registered HTTP route metadata without leaking Gin types.
type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}

type group struct {
	group *gin.RouterGroup
}

type contextAdapter struct {
	ctx *gin.Context
}

// CORSConfig configures the CORS middleware without exposing gin-contrib/cors.
type CORSConfig struct {
	Enabled          bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// StaticSPAConfig 描述一个静态单页应用的挂载点和构建产物目录。
type StaticSPAConfig struct {
	MountPath string
	DistDir   string
}

// New creates a router engine for the given mode.
func New(mode string) *Engine {
	if mode != "" {
		gin.SetMode(mode)
	}
	return &Engine{engine: gin.New()}
}

// Recovery returns the default recovery middleware.
func Recovery() HandlerFunc {
	return wrapGinHandler(gin.Recovery())
}

// CORS returns a configured CORS middleware.
func CORS(cfg CORSConfig) HandlerFunc {
	if !cfg.Enabled {
		return func(c Context) {
			c.Next()
		}
	}
	return wrapGinHandler(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	}))
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.engine.ServeHTTP(w, r)
}

func (e *Engine) Routes() []RouteInfo {
	routes := e.engine.Routes()
	out := make([]RouteInfo, 0, len(routes))
	for _, route := range routes {
		out = append(out, RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		})
	}
	return out
}

// MountStaticSPA 在指定前缀托管静态单页应用，并把非资源路由回退到 index.html。
func (e *Engine) MountStaticSPA(cfg StaticSPAConfig) error {
	mountPath := normalizeMountPath(cfg.MountPath)
	if mountPath == "" || mountPath == "/" {
		return errors.New("mount path must be a non-root absolute path")
	}

	indexPath := filepath.Join(cfg.DistDir, "index.html")
	if info, err := os.Stat(indexPath); err != nil || info.IsDir() {
		return ErrStaticSPAIndexMissing
	}

	handler := func(c *gin.Context) {
		serveStaticSPA(c, cfg.DistDir, indexPath)
	}
	e.engine.GET(mountPath, handler)
	e.engine.GET(mountPath+"/*filepath", handler)
	return nil
}

func (e *Engine) Use(handlers ...HandlerFunc) {
	e.engine.Use(wrapHandlers(handlers)...)
}

func (e *Engine) GET(path string, handler HandlerFunc) {
	e.engine.GET(path, wrapHandler(handler))
}

func (e *Engine) POST(path string, handler HandlerFunc) {
	e.engine.POST(path, wrapHandler(handler))
}

func (e *Engine) PATCH(path string, handler HandlerFunc) {
	e.engine.PATCH(path, wrapHandler(handler))
}

func (e *Engine) PUT(path string, handler HandlerFunc) {
	e.engine.PUT(path, wrapHandler(handler))
}

func (e *Engine) DELETE(path string, handler HandlerFunc) {
	e.engine.DELETE(path, wrapHandler(handler))
}

func (e *Engine) ANY(path string, handler HandlerFunc) {
	e.engine.Any(path, wrapHandler(handler))
}

func (e *Engine) Group(path string) Router {
	return &group{group: e.engine.Group(path)}
}

func (g *group) Use(handlers ...HandlerFunc) {
	g.group.Use(wrapHandlers(handlers)...)
}

func (g *group) GET(path string, handler HandlerFunc) {
	g.group.GET(path, wrapHandler(handler))
}

func (g *group) POST(path string, handler HandlerFunc) {
	g.group.POST(path, wrapHandler(handler))
}

func (g *group) PATCH(path string, handler HandlerFunc) {
	g.group.PATCH(path, wrapHandler(handler))
}

func (g *group) PUT(path string, handler HandlerFunc) {
	g.group.PUT(path, wrapHandler(handler))
}

func (g *group) DELETE(path string, handler HandlerFunc) {
	g.group.DELETE(path, wrapHandler(handler))
}

func (g *group) ANY(path string, handler HandlerFunc) {
	g.group.Any(path, wrapHandler(handler))
}

func (g *group) Group(path string) Router {
	return &group{group: g.group.Group(path)}
}

func serveStaticSPA(c *gin.Context, distDir string, indexPath string) {
	cleanPath := cleanSPARequestPath(c.Param("filepath"))
	if cleanPath == "" {
		c.File(indexPath)
		return
	}

	filePath, ok := safeJoin(distDir, cleanPath)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		c.File(filePath)
		return
	}
	if isStaticAssetPath(cleanPath) {
		c.Status(http.StatusNotFound)
		return
	}
	c.File(indexPath)
}

func cleanSPARequestPath(value string) string {
	value = strings.TrimPrefix(value, "/")
	if value == "" {
		return ""
	}
	cleaned := urlpath.Clean("/" + value)
	if cleaned == "/" || cleaned == "." {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}

func safeJoin(root string, cleanPath string) (string, bool) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	filePath := filepath.Join(absRoot, filepath.FromSlash(cleanPath))
	rel, err := filepath.Rel(absRoot, filePath)
	if err != nil {
		return "", false
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", false
	}
	return filePath, true
}

func isStaticAssetPath(value string) bool {
	if strings.HasPrefix(value, "_nuxt/") || strings.HasPrefix(value, "assets/") {
		return true
	}
	return urlpath.Ext(value) != ""
}

func normalizeMountPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") {
		return value
	}
	if value == "/" {
		return "/"
	}
	return "/" + strings.Trim(strings.TrimRight(value, "/"), "/")
}

func (c contextAdapter) Request() *http.Request {
	return c.ctx.Request
}

func (c contextAdapter) RequestContext() context.Context {
	return c.ctx.Request.Context()
}

func (c contextAdapter) GetHeader(name string) string {
	return c.ctx.GetHeader(name)
}

func (c contextAdapter) Header(name, value string) {
	c.ctx.Header(name, value)
}

func (c contextAdapter) Set(key string, value any) {
	c.ctx.Set(key, value)
}

func (c contextAdapter) Get(key any) (any, bool) {
	return c.ctx.Get(key)
}

func (c contextAdapter) Param(name string) string {
	return c.ctx.Param(name)
}

func (c contextAdapter) BindJSON(dest any) error {
	return c.ctx.ShouldBindJSON(dest)
}

func (c contextAdapter) JSON(status int, body any) {
	c.ctx.JSON(status, body)
}

func (c contextAdapter) Data(status int, contentType string, data []byte) {
	c.ctx.Data(status, contentType, data)
}

func (c contextAdapter) AbortWithStatusJSON(status int, body any) {
	c.ctx.AbortWithStatusJSON(status, body)
}

func (c contextAdapter) Next() {
	c.ctx.Next()
}

func (c contextAdapter) Path() string {
	return c.ctx.Request.URL.Path
}

func (c contextAdapter) Method() string {
	return c.ctx.Request.Method
}

func (c contextAdapter) ClientIP() string {
	return c.ctx.ClientIP()
}

func (c contextAdapter) Status() int {
	return c.ctx.Writer.Status()
}

func wrapHandlers(handlers []HandlerFunc) []gin.HandlerFunc {
	wrapped := make([]gin.HandlerFunc, 0, len(handlers))
	for _, handler := range handlers {
		wrapped = append(wrapped, wrapHandler(handler))
	}
	return wrapped
}

func wrapHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(contextAdapter{ctx: c})
	}
}

func wrapGinHandler(handler gin.HandlerFunc) HandlerFunc {
	return func(c Context) {
		adapter, ok := c.(contextAdapter)
		if !ok {
			c.Next()
			return
		}
		handler(adapter.ctx)
	}
}
