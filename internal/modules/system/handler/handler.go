package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/middleware"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/internal/modules/system/model"
	"github.com/rei0721/go-scaffold/internal/modules/system/service"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/web"
	"github.com/rei0721/go-scaffold/types/result"
)

type Handler struct {
	service    service.Service
	authorizer middleware.Authorizer
	logger     logger.Logger
}

type createDictionaryRequest struct {
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Name        string `json:"name" binding:"required"`
	Status      string `json:"status"`
}

type updateDictionaryRequest struct {
	Description *string `json:"description"`
	Name        *string `json:"name"`
	Status      *string `json:"status"`
}

type createDictionaryItemRequest struct {
	Extra  string `json:"extra"`
	Label  string `json:"label" binding:"required"`
	Sort   int    `json:"sort"`
	Status string `json:"status"`
	Value  string `json:"value" binding:"required"`
}

type updateDictionaryItemRequest struct {
	Extra  *string `json:"extra"`
	Label  *string `json:"label"`
	Sort   *int    `json:"sort"`
	Status *string `json:"status"`
	Value  *string `json:"value"`
}

type createParameterRequest struct {
	Description string `json:"description"`
	Key         string `json:"key" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Value       string `json:"value" binding:"required"`
}

type updateParameterRequest struct {
	Description *string `json:"description"`
	Key         *string `json:"key"`
	Name        *string `json:"name"`
	Value       *string `json:"value"`
}

type deleteOperationRecordsRequest struct {
	IDs []systemID `json:"ids"`
}

type deleteParametersRequest struct {
	IDs []systemID `json:"ids"`
}

type systemID int64

func New(service service.Service, authorizer middleware.Authorizer, logger logger.Logger) *Handler {
	return &Handler{service: service, authorizer: authorizer, logger: logger}
}

func (h *Handler) ListMenus(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	groups, err := h.service.ListMenus(c.RequestContext())
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, h.filterMenus(c.RequestContext(), principal, groups))
}

func (h *Handler) ListAPIs(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	groups, err := h.service.ListAPIs(c.RequestContext())
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, groups)
}

func (h *Handler) ListConfig(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	snapshot, err := h.service.ListConfig(c.RequestContext())
	writeOK(c, snapshot, err, h.writeError)
}

func (h *Handler) GetServerInfo(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	info, err := h.service.GetServerInfo(c.RequestContext())
	writeOK(c, info, err, h.writeError)
}

func (h *Handler) SyncAPIs(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	syncResult, err := h.service.SyncAPIs(c.RequestContext())
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, syncResult)
}

func (h *Handler) SyncPermissions(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	syncResult, err := h.service.SyncPermissions(c.RequestContext())
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, syncResult)
}

func (h *Handler) ListDictionaries(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	catalog, err := h.service.ListDictionaries(c.RequestContext())
	writeOK(c, catalog, err, h.writeError)
}

func (h *Handler) CreateDictionary(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req createDictionaryRequest
	if !bind(c, &req) {
		return
	}
	dictionary, err := h.service.CreateDictionary(c.RequestContext(), service.CreateDictionaryInput{
		Code:        req.Code,
		Description: req.Description,
		Name:        req.Name,
		Status:      req.Status,
	})
	writeCreated(c, dictionary, err, h.writeError)
}

func (h *Handler) UpdateDictionary(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "dictionaryId")
	if !ok {
		return
	}
	var req updateDictionaryRequest
	if !bind(c, &req) {
		return
	}
	dictionary, err := h.service.UpdateDictionary(c.RequestContext(), id, service.UpdateDictionaryInput{
		Description: req.Description,
		Name:        req.Name,
		Status:      req.Status,
	})
	writeOK(c, dictionary, err, h.writeError)
}

func (h *Handler) DeleteDictionary(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "dictionaryId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteDictionary(c.RequestContext(), id), h.writeError)
}

func (h *Handler) CreateDictionaryItem(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	dictionaryID, ok := parseInt64Param(c, "dictionaryId")
	if !ok {
		return
	}
	var req createDictionaryItemRequest
	if !bind(c, &req) {
		return
	}
	item, err := h.service.CreateDictionaryItem(c.RequestContext(), dictionaryID, service.CreateDictionaryItemInput{
		Extra:  req.Extra,
		Label:  req.Label,
		Sort:   req.Sort,
		Status: req.Status,
		Value:  req.Value,
	})
	writeCreated(c, item, err, h.writeError)
}

func (h *Handler) UpdateDictionaryItem(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "itemId")
	if !ok {
		return
	}
	var req updateDictionaryItemRequest
	if !bind(c, &req) {
		return
	}
	item, err := h.service.UpdateDictionaryItem(c.RequestContext(), id, service.UpdateDictionaryItemInput{
		Extra:  req.Extra,
		Label:  req.Label,
		Sort:   req.Sort,
		Status: req.Status,
		Value:  req.Value,
	})
	writeOK(c, item, err, h.writeError)
}

func (h *Handler) DeleteDictionaryItem(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "itemId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteDictionaryItem(c.RequestContext(), id), h.writeError)
}

func (h *Handler) ListOperationRecords(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	values := c.Request().URL.Query()
	page, ok := parseIntQuery(c, "page", 1)
	if !ok {
		return
	}
	pageSize, ok := parseIntQuery(c, "pageSize", 10)
	if !ok {
		return
	}
	status, ok := parseIntQuery(c, "status", 0)
	if !ok {
		return
	}
	records, err := h.service.ListOperationRecords(c.RequestContext(), service.OperationRecordFilter{
		Method:      values.Get("method"),
		Page:        page,
		PageSize:    pageSize,
		Path:        values.Get("path"),
		Status:      status,
		StatusClass: values.Get("statusClass"),
	})
	writeOK(c, records, err, h.writeError)
}

func (h *Handler) DeleteOperationRecords(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req deleteOperationRecordsRequest
	if !bind(c, &req) {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteOperationRecords(c.RequestContext(), req.int64s()), h.writeError)
}

func (h *Handler) ListParameters(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	values := c.Request().URL.Query()
	page, ok := parseIntQuery(c, "page", 1)
	if !ok {
		return
	}
	pageSize, ok := parseIntQuery(c, "pageSize", 10)
	if !ok {
		return
	}
	startCreatedAt, ok := parseTimeQuery(c, "startCreatedAt", false)
	if !ok {
		return
	}
	endCreatedAt, ok := parseTimeQuery(c, "endCreatedAt", true)
	if !ok {
		return
	}
	parameters, err := h.service.ListParameters(c.RequestContext(), service.ParameterFilter{
		EndCreatedAt:   endCreatedAt,
		Key:            values.Get("key"),
		Name:           values.Get("name"),
		Page:           page,
		PageSize:       pageSize,
		StartCreatedAt: startCreatedAt,
	})
	writeOK(c, parameters, err, h.writeError)
}

func (h *Handler) CreateParameter(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req createParameterRequest
	if !bind(c, &req) {
		return
	}
	parameter, err := h.service.CreateParameter(c.RequestContext(), service.CreateParameterInput{
		Description: req.Description,
		Key:         req.Key,
		Name:        req.Name,
		Value:       req.Value,
	})
	writeCreated(c, parameter, err, h.writeError)
}

func (h *Handler) GetParameter(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "parameterId")
	if !ok {
		return
	}
	parameter, err := h.service.FindParameter(c.RequestContext(), id)
	writeOK(c, parameter, err, h.writeError)
}

func (h *Handler) GetParameterByKey(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	parameter, err := h.service.FindParameterByKey(c.RequestContext(), c.Request().URL.Query().Get("key"))
	writeOK(c, parameter, err, h.writeError)
}

func (h *Handler) UpdateParameter(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "parameterId")
	if !ok {
		return
	}
	var req updateParameterRequest
	if !bind(c, &req) {
		return
	}
	parameter, err := h.service.UpdateParameter(c.RequestContext(), id, service.UpdateParameterInput{
		Description: req.Description,
		Key:         req.Key,
		Name:        req.Name,
		Value:       req.Value,
	})
	writeOK(c, parameter, err, h.writeError)
}

func (h *Handler) DeleteParameter(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "parameterId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteParameter(c.RequestContext(), id), h.writeError)
}

func (h *Handler) DeleteParameters(c web.Context) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req deleteParametersRequest
	if !bind(c, &req) {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteParameters(c.RequestContext(), req.int64s()), h.writeError)
}

func (h *Handler) RegisterAPIs(entries []model.APIEntry) {
	h.service.RegisterAPIs(entries)
}

func (h *Handler) RecordOperation(ctx context.Context, input service.OperationRecordInput) error {
	return h.service.RecordOperation(ctx, input)
}

func (h *Handler) filterMenus(ctx context.Context, principal iamservice.Principal, groups []model.MenuGroup) []model.MenuGroup {
	filtered := make([]model.MenuGroup, 0, len(groups))
	for _, group := range groups {
		items := make([]model.MenuItem, 0, len(group.Items))
		for _, item := range group.Items {
			if item.Permission == "" || h.allowed(ctx, principal, item.Permission) {
				items = append(items, item)
			}
		}
		if len(items) == 0 {
			continue
		}
		group.Items = items
		filtered = append(filtered, group)
	}
	return filtered
}

func (h *Handler) allowed(ctx context.Context, principal iamservice.Principal, permission string) bool {
	if h.authorizer == nil {
		return false
	}
	obj, act := permissionObjectAction(permission)
	if obj == "" || act == "" {
		return false
	}
	allowed, err := h.authorizer.Authorize(ctx, principal, obj, act)
	return err == nil && allowed
}

func (h *Handler) writeError(c web.Context, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		result.Fail(c, http.StatusRequestTimeout, "request canceled")
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrDuplicate):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrNotFound):
		result.NotFound(c, err.Error())
	case errors.Is(err, service.ErrStorageUnavailable):
		result.Fail(c, http.StatusServiceUnavailable, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("system request failed", "error", err)
		}
		result.InternalError(c, "internal server error")
	}
}

func requirePrincipal(c web.Context) (iamservice.Principal, bool) {
	principal, ok := middleware.GetPrincipal(c)
	if !ok {
		result.Unauthorized(c, "missing principal")
		return iamservice.Principal{}, false
	}
	return principal, true
}

func bind(c web.Context, dest any) bool {
	if err := c.BindJSON(dest); err != nil {
		result.BadRequest(c, err.Error())
		return false
	}
	return true
}

func parseInt64Param(c web.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		result.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return id, true
}

func parseIntQuery(c web.Context, name string, fallback int) (int, bool) {
	raw := strings.TrimSpace(c.Request().URL.Query().Get(name))
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		result.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return value, true
}

func parseTimeQuery(c web.Context, name string, endExclusive bool) (*time.Time, bool) {
	raw := strings.TrimSpace(c.Request().URL.Query().Get(name))
	if raw == "" {
		return nil, true
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		value, err := time.Parse(layout, raw)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" && endExclusive {
			value = value.AddDate(0, 0, 1)
		}
		return &value, true
	}
	result.BadRequest(c, "invalid "+name)
	return nil, false
}

func (r deleteOperationRecordsRequest) int64s() []int64 {
	ids := make([]int64, 0, len(r.IDs))
	for _, id := range r.IDs {
		ids = append(ids, int64(id))
	}
	return ids
}

func (r deleteParametersRequest) int64s() []int64 {
	ids := make([]int64, 0, len(r.IDs))
	for _, id := range r.IDs {
		ids = append(ids, int64(id))
	}
	return ids
}

func (id *systemID) UnmarshalJSON(raw []byte) error {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return service.ErrInvalidInput
	}
	if strings.HasPrefix(value, `"`) {
		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return err
		}
		value = strings.TrimSpace(unquoted)
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	*id = systemID(parsed)
	return nil
}

func writeOK(c web.Context, data any, err error, writeError func(web.Context, error)) {
	if err != nil {
		writeError(c, err)
		return
	}
	result.OK(c, data)
}

func writeCreated(c web.Context, data any, err error, writeError func(web.Context, error)) {
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result.Success(data))
}

func permissionObjectAction(code string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(code), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	return parts[0], parts[1]
}
