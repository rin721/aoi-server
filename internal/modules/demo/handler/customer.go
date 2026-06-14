package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/rei0721/go-scaffold/internal/middleware"
	"github.com/rei0721/go-scaffold/internal/modules/demo/service"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/internal/ports"
	"github.com/rei0721/go-scaffold/types/result"
)

// CustomerHandler 将客户资源示例暴露为受保护的 HTTP 接口。
type CustomerHandler struct {
	service service.CustomerService
	logger  ports.Logger
}

type createCustomerRequest struct {
	CustomerName      string `json:"customerName" binding:"required"`
	CustomerPhoneData string `json:"customerPhoneData" binding:"required"`
}

type updateCustomerRequest struct {
	CustomerName      *string `json:"customerName"`
	CustomerPhoneData *string `json:"customerPhoneData"`
}

// NewCustomerHandler 创建客户资源示例 handler。
func NewCustomerHandler(service service.CustomerService, logger ports.Logger) *CustomerHandler {
	return &CustomerHandler{service: service, logger: logger}
}

// Create 处理新增客户资源请求，并把资源归属绑定到当前登录主体。
func (h *CustomerHandler) Create(c ports.HTTPContext) {
	principal, ok := customerPrincipal(c)
	if !ok {
		return
	}
	var req createCustomerRequest
	if err := c.BindJSON(&req); err != nil {
		result.BadRequest(c, err.Error())
		return
	}
	customer, err := h.service.Create(c.RequestContext(), service.CreateCustomerInput{
		Principal:         principal,
		CustomerName:      req.CustomerName,
		CustomerPhoneData: req.CustomerPhoneData,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result.Success(customer))
}

// List 按当前主体的资源可见范围返回客户列表。
func (h *CustomerHandler) List(c ports.HTTPContext) {
	principal, ok := customerPrincipal(c)
	if !ok {
		return
	}
	page := parsePositiveIntQuery(c, "page", 1)
	pageSize := parsePositiveIntQuery(c, "pageSize", 10)
	customers, err := h.service.List(c.RequestContext(), service.ListCustomerInput{
		Principal: principal,
		Keyword:   strings.TrimSpace(c.Request().URL.Query().Get("keyword")),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, customers)
}

// Get 查询当前主体可见范围内的单个客户资源。
func (h *CustomerHandler) Get(c ports.HTTPContext) {
	principal, ok := customerPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	customer, err := h.service.Get(c.RequestContext(), service.CustomerIdentityInput{Principal: principal, ID: id})
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, customer)
}

// Update 处理客户资源局部更新请求。
func (h *CustomerHandler) Update(c ports.HTTPContext) {
	principal, ok := customerPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateCustomerRequest
	if err := c.BindJSON(&req); err != nil {
		result.BadRequest(c, err.Error())
		return
	}
	customer, err := h.service.Update(c.RequestContext(), service.UpdateCustomerInput{
		Principal:         principal,
		ID:                id,
		CustomerName:      req.CustomerName,
		CustomerPhoneData: req.CustomerPhoneData,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, customer)
}

// Delete 在当前主体可见范围内软删除客户资源。
func (h *CustomerHandler) Delete(c ports.HTTPContext) {
	principal, ok := customerPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.RequestContext(), service.CustomerIdentityInput{Principal: principal, ID: id}); err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, map[string]bool{"deleted": true})
}

func (h *CustomerHandler) writeError(c ports.HTTPContext, err error) {
	switch {
	case errors.Is(err, service.ErrCustomerPrincipalRequired):
		result.Unauthorized(c, err.Error())
	case errors.Is(err, service.ErrCustomerNameRequired), errors.Is(err, service.ErrCustomerPhoneRequired):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrCustomerNotFound):
		result.NotFound(c, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("customer request failed", "error", err)
		}
		result.InternalError(c, "internal server error")
	}
}

func customerPrincipal(c ports.HTTPContext) (iamservice.Principal, bool) {
	principal, ok := middleware.GetPrincipal(c)
	if !ok {
		result.Unauthorized(c, "missing principal")
		return iamservice.Principal{}, false
	}
	return principal, true
}

func parsePositiveIntQuery(c ports.HTTPContext, name string, fallback int) int {
	raw := strings.TrimSpace(c.Request().URL.Query().Get(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
