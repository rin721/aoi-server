package handler

// 本文件定义 Demo Todo 的 HTTP 适配层，负责请求绑定、错误映射和统一 Result 响应。

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/rei0721/go-scaffold/internal/modules/demo/service"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/web"
	"github.com/rei0721/go-scaffold/types/result"
)

// TodoHandler 将 Demo Todo 服务暴露为 HTTP handler，负责请求绑定、状态码选择和统一响应。
type TodoHandler struct {
	service service.TodoService
	logger  logger.Logger
}

type createTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type updateTodoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

// NewTodoHandler 创建 TodoHandler，并显式注入服务与日志依赖以便测试替换。
func NewTodoHandler(service service.TodoService, logger logger.Logger) *TodoHandler {
	return &TodoHandler{service: service, logger: logger}
}

// Create 处理新建 Todo 请求，绑定 JSON 后把业务校验与事务交给服务层完成。
func (h *TodoHandler) Create(c web.Context) {
	var req createTodoRequest
	if err := c.BindJSON(&req); err != nil {
		result.BadRequest(c, err.Error())
		return
	}

	todo, err := h.service.Create(c.RequestContext(), service.CreateTodoInput{
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result.Success(todo))
}

// List 处理 Todo 列表请求，当前示例使用固定分页上限以避免开放无限制查询。
func (h *TodoHandler) List(c web.Context) {
	todos, err := h.service.List(c.RequestContext())
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, todos)
}

// Get 根据路径参数查询单个 Todo，并把不存在场景映射为统一失败响应。
func (h *TodoHandler) Get(c web.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	todo, err := h.service.Get(c.RequestContext(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, todo)
}

// Update 处理部分字段更新请求，只有 JSON 中出现的字段会传递给服务层。
func (h *TodoHandler) Update(c web.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateTodoRequest
	if err := c.BindJSON(&req); err != nil {
		result.BadRequest(c, err.Error())
		return
	}

	todo, err := h.service.Update(c.RequestContext(), id, service.UpdateTodoInput{
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, todo)
}

// Delete 根据路径参数删除 Todo，并保持删除成功响应不泄露存储层细节。
func (h *TodoHandler) Delete(c web.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.RequestContext(), id); err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, map[string]any{"deleted": true})
}

// parseID 从路径参数解析 Todo ID，并把非法输入转换为统一的客户端错误响应。
func parseID(c web.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		result.BadRequest(c, "invalid id")
		return 0, false
	}
	return uint(value), true
}

// writeError 将服务层领域错误映射为 Result 响应，避免 handler 泄露存储层错误细节。
func (h *TodoHandler) writeError(c web.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTodoTitleRequired):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrTodoNotFound):
		result.NotFound(c, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("todo request failed", "error", err)
		}
		result.InternalError(c, "internal server error")
	}
}
