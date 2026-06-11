package httptransport

// 本测试文件固定 HTTP 路由、中间件顺序和错误响应契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	"github.com/rei0721/go-scaffold/internal/middleware"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	"github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	demoservice "github.com/rei0721/go-scaffold/internal/modules/demo/service"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/web"
	apperrors "github.com/rei0721/go-scaffold/types/errors"
)

type integrationResponse struct {
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
	TraceID    string          `json:"traceId"`
	ServerTime int64           `json:"serverTime"`
}

// TestNewRouterDemoTodoIntegration 固定 HTTP 路由、中间件顺序和错误响应契约，确保后续注释补全或结构调整不改变该场景。
func TestNewRouterDemoTodoIntegration(t *testing.T) {
	router, db := newDemoIntegrationRouter(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	}()

	preflight := performRawRequest(router, http.MethodOptions, "/api/v1/demo/todos", "", map[string]string{
		"Origin":                        "http://example.com",
		"Access-Control-Request-Method": http.MethodPost,
	})
	if preflight.Code != http.StatusNoContent {
		t.Fatalf("expected CORS preflight status %d, got %d with body %s", http.StatusNoContent, preflight.Code, preflight.Body.String())
	}
	if got := preflight.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("expected preflight CORS origin header, got %q", got)
	}

	createRecorder, createBody := performIntegrationRequest(t, router, http.MethodPost, "/api/v1/demo/todos", `{
		"title": "  Router integration  ",
		"description": "  through handler and service  "
	}`, map[string]string{
		"Origin":     "http://example.com",
		"X-Trace-ID": "trace-demo",
	})
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createRecorder.Code, createRecorder.Body.String())
	}
	assertIntegrationSuccess(t, createBody)
	if got := createRecorder.Header().Get("X-Trace-ID"); got != "trace-demo" {
		t.Fatalf("expected trace header to round-trip, got %q", got)
	}
	if got := createRecorder.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("expected CORS origin header, got %q", got)
	}
	created := decodeIntegrationData[model.Todo](t, createBody)
	if created.ID == 0 {
		t.Fatal("expected created todo to have an id")
	}
	if created.Title != "Router integration" {
		t.Fatalf("expected trimmed title, got %q", created.Title)
	}
	if created.Description != "through handler and service" {
		t.Fatalf("expected trimmed description, got %q", created.Description)
	}

	listRecorder, listBody := performIntegrationRequest(t, router, http.MethodGet, "/api/v1/demo/todos", "", nil)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, listRecorder.Code, listRecorder.Body.String())
	}
	listed := decodeIntegrationData[[]model.Todo](t, listBody)
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("expected list to include created todo %d, got %#v", created.ID, listed)
	}

	getRecorder, getBody := performIntegrationRequest(t, router, http.MethodGet, "/api/v1/demo/todos/"+uintPath(created.ID), "", nil)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getRecorder.Code, getRecorder.Body.String())
	}
	gotTodo := decodeIntegrationData[model.Todo](t, getBody)
	if gotTodo.ID != created.ID || gotTodo.Title != created.Title {
		t.Fatalf("expected fetched todo %#v, got %#v", created, gotTodo)
	}

	updateRecorder, updateBody := performIntegrationRequest(t, router, http.MethodPut, "/api/v1/demo/todos/"+uintPath(created.ID), `{
		"title": "  Updated through router  ",
		"completed": true
	}`, nil)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateRecorder.Code, updateRecorder.Body.String())
	}
	updated := decodeIntegrationData[model.Todo](t, updateBody)
	if updated.Title != "Updated through router" || !updated.Completed {
		t.Fatalf("expected updated todo to be completed with trimmed title, got %#v", updated)
	}

	deleteRecorder, deleteBody := performIntegrationRequest(t, router, http.MethodDelete, "/api/v1/demo/todos/"+uintPath(created.ID), "", nil)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusOK, deleteRecorder.Code, deleteRecorder.Body.String())
	}
	deleted := decodeIntegrationData[map[string]bool](t, deleteBody)
	if !deleted["deleted"] {
		t.Fatalf("expected deleted flag, got %#v", deleted)
	}

	missingRecorder, missingBody := performIntegrationRequest(t, router, http.MethodGet, "/api/v1/demo/todos/"+uintPath(created.ID), "", nil)
	if missingRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing get status %d, got %d with body %s", http.StatusNotFound, missingRecorder.Code, missingRecorder.Body.String())
	}
	if missingBody.Code != apperrors.ErrResourceNotFound {
		t.Fatalf("expected not found error code %d, got %d", apperrors.ErrResourceNotFound, missingBody.Code)
	}
}

// TestNewRouterRecoveryIncludesTraceID 固定 HTTP 路由、中间件顺序和错误响应契约，确保后续注释补全或结构调整不改变该场景。
func TestNewRouterRecoveryIncludesTraceID(t *testing.T) {
	log := &recordingLogger{}
	router := newTestRouter(RouterDeps{
		Logger: log,
		Middleware: middleware.MiddlewareConfig{
			Recovery: middleware.RecoveryConfig{Enabled: true},
			TraceID:  middleware.TraceIDConfig{Enabled: true, HeaderName: "X-Request-ID"},
		},
	})
	router.GET("/panic", func(c web.Context) {
		panic("boom")
	})

	recorder, body := performIntegrationRequest(t, router, http.MethodGet, "/panic", "", map[string]string{
		"X-Request-ID": "panic-trace",
	})
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected panic status %d, got %d with body %s", http.StatusInternalServerError, recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("X-Request-ID") != "panic-trace" {
		t.Fatalf("expected response trace header panic-trace, got %q", recorder.Header().Get("X-Request-ID"))
	}
	if body.Code != apperrors.ErrInternalServer {
		t.Fatalf("expected internal error code %d, got %d", apperrors.ErrInternalServer, body.Code)
	}
	if body.TraceID != "panic-trace" {
		t.Fatalf("expected response trace id panic-trace, got %q", body.TraceID)
	}
	if !log.hasEntry("panic recovered") {
		t.Fatalf("expected recovery logger entry, got %#v", log.entries)
	}
}

// newDemoIntegrationRouter 构造当前测试场景所需的最小依赖集合，避免测试直接耦合生产装配流程。
func newDemoIntegrationRouter(t *testing.T) (*web.Engine, database.Database) {
	t.Helper()

	db, err := database.New(&database.Config{
		Driver: database.DriverSQLite,
		DBName: filepath.Join(t.TempDir(), "demo-router.db"),
	})
	if err != nil {
		t.Fatalf("create sqlite database: %v", err)
	}
	if _, err := dbapp.ApplyDemoSchema(context.Background(), db, string(database.DriverSQLite)); err != nil {
		_ = db.Close()
		t.Fatalf("apply todo schema: %v", err)
	}

	todoService := demoservice.NewTodoService(db, repository.NewTodoRepository(db))
	todoHandler := demohandler.NewTodoHandler(todoService, &recordingLogger{})

	return newTestRouter(RouterDeps{
		Database:    db,
		Logger:      &recordingLogger{},
		TodoHandler: todoHandler,
		Middleware: middleware.MiddlewareConfig{
			Recovery: middleware.RecoveryConfig{Enabled: true},
			Logger:   middleware.LoggerConfig{Enabled: true},
			TraceID:  middleware.TraceIDConfig{Enabled: true, HeaderName: "X-Trace-ID"},
			CORS: middleware.CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"http://example.com"},
				AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
				AllowHeaders: []string{"Content-Type", "X-Trace-ID"},
				MaxAge:       60,
			},
		},
	}), db
}

// performIntegrationRequest 执行测试 HTTP 请求并返回响应记录器，封装路由调用细节。
func performIntegrationRequest(t *testing.T, router http.Handler, method string, path string, body string, headers map[string]string) (*httptest.ResponseRecorder, integrationResponse) {
	t.Helper()

	recorder := performRawRequest(router, method, path, body, headers)
	var response integrationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response body %q: %v", recorder.Body.String(), err)
	}
	return recorder, response
}

// performRawRequest 执行测试 HTTP 请求并返回响应记录器，封装路由调用细节。
func performRawRequest(router http.Handler, method string, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Host = "api.local"
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// decodeIntegrationData 解析测试响应中的数据负载，并在格式不符合契约时直接终止用例。
func decodeIntegrationData[T any](t *testing.T, response integrationResponse) T {
	t.Helper()

	var data T
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("failed to decode response data %q: %v", string(response.Data), err)
	}
	return data
}

// assertIntegrationSuccess 校验测试响应或状态中的关键字段，使测试断言聚焦在对外契约而非重复解析细节。
func assertIntegrationSuccess(t *testing.T, response integrationResponse) {
	t.Helper()

	if response.Code != 0 {
		t.Fatalf("expected success response code 0, got %d", response.Code)
	}
	if response.Message != "success" {
		t.Fatalf("expected success message, got %q", response.Message)
	}
	if len(response.Data) == 0 {
		t.Fatal("expected response data")
	}
}

// uintPath 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func uintPath(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

type recordingLogger struct {
	mu      sync.Mutex
	entries []string
}

// Debug 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (l *recordingLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.record(msg)
}

// Info 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (l *recordingLogger) Info(msg string, keysAndValues ...interface{}) {
	l.record(msg)
}

// Warn 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (l *recordingLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.record(msg)
}

// Error 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (l *recordingLogger) Error(msg string, keysAndValues ...interface{}) {
	l.record(msg)
}

// Fatal 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (l *recordingLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.record(msg)
}

// With 实现测试日志桩的字段绑定入口，返回自身以保持 logger.Logger 链式调用契约。
func (l *recordingLogger) With(keysAndValues ...interface{}) logger.Logger {
	return l
}

// Sync 实现测试日志桩的刷新入口，测试环境不持有真实缓冲区。
func (l *recordingLogger) Sync() error {
	return nil
}

// Reload 实现测试桩的配置重载入口，用于验证调用路径而不触发真实资源替换。
func (l *recordingLogger) Reload(*logger.Config) error {
	return nil
}

// record 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func (l *recordingLogger) record(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, msg)
}

// hasEntry 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func (l *recordingLogger) hasEntry(msg string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, entry := range l.entries {
		if entry == msg {
			return true
		}
	}
	return false
}
