package handler

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/middleware"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/internal/modules/system/model"
	"github.com/rei0721/go-scaffold/internal/modules/system/service"
	"github.com/rei0721/go-scaffold/internal/ports"
	"github.com/rei0721/go-scaffold/types/result"
)

type Handler struct {
	service    service.Service
	authorizer middleware.Authorizer
	logger     ports.Logger
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

type deleteVersionsRequest struct {
	IDs []systemID `json:"ids"`
}

type updateConfigRequest struct {
	Items   []updateConfigItemRequest `json:"items" binding:"required"`
	Persist bool                      `json:"persist"`
}

type updateConfigItemRequest struct {
	Key   string `json:"key" binding:"required"`
	Value any    `json:"value"`
}

type exportVersionRequest struct {
	APICodes        []string `json:"apiCodes"`
	Description     string   `json:"description"`
	DictionaryCodes []string `json:"dictionaryCodes"`
	MenuCodes       []string `json:"menuCodes"`
	VersionCode     string   `json:"versionCode" binding:"required"`
	VersionName     string   `json:"versionName" binding:"required"`
}

type importVersionRequest struct {
	VersionData string `json:"versionData" binding:"required"`
}

type upsertMediaCategoryRequest struct {
	ID       systemID `json:"id"`
	ParentID systemID `json:"parentId"`
	Name     string   `json:"name" binding:"required"`
	Sort     int      `json:"sort"`
}

type importMediaURLsRequest struct {
	CategoryID systemID             `json:"categoryId"`
	Items      []importMediaURLItem `json:"items"`
	Text       string               `json:"text"`
}

type importMediaURLItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type updateMediaAssetRequest struct {
	DisplayName string `json:"displayName" binding:"required"`
}

type checkMediaResumableUploadRequest struct {
	CategoryID systemID `json:"categoryId"`
	ChunkSize  int64    `json:"chunkSize"`
	ChunkTotal int      `json:"chunkTotal"`
	FileHash   string   `json:"fileHash" binding:"required"`
	FileName   string   `json:"fileName" binding:"required"`
	SizeBytes  int64    `json:"sizeBytes" binding:"required"`
}

type mediaResumableSessionRequest struct {
	FileHash  string   `json:"fileHash" binding:"required"`
	SessionID systemID `json:"sessionId" binding:"required"`
}

type systemID int64

func New(service service.Service, authorizer middleware.Authorizer, logger ports.Logger) *Handler {
	return &Handler{service: service, authorizer: authorizer, logger: logger}
}

func (h *Handler) ListMenus(c ports.HTTPContext) {
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

func (h *Handler) ListAPIs(c ports.HTTPContext) {
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

func (h *Handler) ListConfig(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	snapshot, err := h.service.ListConfig(c.RequestContext())
	writeOK(c, snapshot, err, h.writeError)
}

func (h *Handler) UpdateConfig(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req updateConfigRequest
	if !bind(c, &req) {
		return
	}
	input := service.UpdateConfigInput{Items: make([]service.UpdateConfigItem, 0, len(req.Items)), Persist: req.Persist}
	for _, item := range req.Items {
		input.Items = append(input.Items, service.UpdateConfigItem{
			Key:   item.Key,
			Value: item.Value,
		})
	}
	snapshot, err := h.service.UpdateConfig(c.RequestContext(), input)
	writeOK(c, snapshot, err, h.writeError)
}

func (h *Handler) GetServerInfo(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	info, err := h.service.GetServerInfo(c.RequestContext())
	writeOK(c, info, err, h.writeError)
}

func (h *Handler) SyncAPIs(c ports.HTTPContext) {
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

func (h *Handler) SyncPermissions(c ports.HTTPContext) {
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

func (h *Handler) ListDictionaries(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	catalog, err := h.service.ListDictionaries(c.RequestContext())
	writeOK(c, catalog, err, h.writeError)
}

func (h *Handler) CreateDictionary(c ports.HTTPContext) {
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

func (h *Handler) UpdateDictionary(c ports.HTTPContext) {
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

func (h *Handler) DeleteDictionary(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "dictionaryId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteDictionary(c.RequestContext(), id), h.writeError)
}

func (h *Handler) CreateDictionaryItem(c ports.HTTPContext) {
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

func (h *Handler) UpdateDictionaryItem(c ports.HTTPContext) {
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

func (h *Handler) DeleteDictionaryItem(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "itemId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteDictionaryItem(c.RequestContext(), id), h.writeError)
}

func (h *Handler) ListOperationRecords(c ports.HTTPContext) {
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

func (h *Handler) DeleteOperationRecords(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req deleteOperationRecordsRequest
	if !bind(c, &req) {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteOperationRecords(c.RequestContext(), req.int64s()), h.writeError)
}

func (h *Handler) ListParameters(c ports.HTTPContext) {
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

func (h *Handler) CreateParameter(c ports.HTTPContext) {
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

func (h *Handler) GetParameter(c ports.HTTPContext) {
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

func (h *Handler) GetParameterByKey(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	parameter, err := h.service.FindParameterByKey(c.RequestContext(), c.Request().URL.Query().Get("key"))
	writeOK(c, parameter, err, h.writeError)
}

func (h *Handler) UpdateParameter(c ports.HTTPContext) {
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

func (h *Handler) DeleteParameter(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "parameterId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteParameter(c.RequestContext(), id), h.writeError)
}

func (h *Handler) DeleteParameters(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req deleteParametersRequest
	if !bind(c, &req) {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteParameters(c.RequestContext(), req.int64s()), h.writeError)
}

func (h *Handler) ListVersionSources(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	sources, err := h.service.ListVersionSources(c.RequestContext())
	writeOK(c, sources, err, h.writeError)
}

func (h *Handler) ListVersions(c ports.HTTPContext) {
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
	versions, err := h.service.ListVersions(c.RequestContext(), service.VersionFilter{
		EndCreatedAt:   endCreatedAt,
		Page:           page,
		PageSize:       pageSize,
		StartCreatedAt: startCreatedAt,
		VersionCode:    values.Get("versionCode"),
		VersionName:    values.Get("versionName"),
	})
	writeOK(c, versions, err, h.writeError)
}

func (h *Handler) GetVersion(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "versionId")
	if !ok {
		return
	}
	version, err := h.service.FindVersion(c.RequestContext(), id)
	writeOK(c, version, err, h.writeError)
}

func (h *Handler) ExportVersion(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req exportVersionRequest
	if !bind(c, &req) {
		return
	}
	version, err := h.service.ExportVersion(c.RequestContext(), service.ExportVersionInput{
		APICodes:        req.APICodes,
		CreatedBy:       principal.UserID,
		CreatorUsername: principal.Username,
		Description:     req.Description,
		DictionaryCodes: req.DictionaryCodes,
		MenuCodes:       req.MenuCodes,
		VersionCode:     req.VersionCode,
		VersionName:     req.VersionName,
	})
	writeCreated(c, version, err, h.writeError)
}

func (h *Handler) ImportVersion(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req importVersionRequest
	if !bind(c, &req) {
		return
	}
	importResult, err := h.service.ImportVersion(c.RequestContext(), service.ImportVersionInput{
		CreatedBy:       principal.UserID,
		CreatorUsername: principal.Username,
		VersionData:     req.VersionData,
	})
	writeCreated(c, importResult, err, h.writeError)
}

func (h *Handler) DownloadVersion(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "versionId")
	if !ok {
		return
	}
	pkg, err := h.service.GetVersionPackage(c.RequestContext(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="system-version-`+strconv.FormatInt(id, 10)+`.json"`)
	result.OK(c, pkg)
}

func (h *Handler) DeleteVersion(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "versionId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteVersion(c.RequestContext(), id), h.writeError)
}

func (h *Handler) DeleteVersions(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req deleteVersionsRequest
	if !bind(c, &req) {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteVersions(c.RequestContext(), req.int64s()), h.writeError)
}

func (h *Handler) ListMediaCategories(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	catalog, err := h.service.ListMediaCategories(c.RequestContext())
	writeOK(c, catalog, err, h.writeError)
}

func (h *Handler) UpsertMediaCategory(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	var req upsertMediaCategoryRequest
	if !bind(c, &req) {
		return
	}
	category, err := h.service.UpsertMediaCategory(c.RequestContext(), service.UpsertMediaCategoryInput{
		ID:       int64(req.ID),
		Name:     req.Name,
		ParentID: int64(req.ParentID),
		Sort:     req.Sort,
	})
	if int64(req.ID) > 0 {
		writeOK(c, category, err, h.writeError)
		return
	}
	writeCreated(c, category, err, h.writeError)
}

func (h *Handler) DeleteMediaCategory(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "categoryId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteMediaCategory(c.RequestContext(), id), h.writeError)
}

func (h *Handler) ListMediaAssets(c ports.HTTPContext) {
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
	categoryID, ok := parseInt64Query(c, "categoryId", 0)
	if !ok {
		return
	}
	assets, err := h.service.ListMediaAssets(c.RequestContext(), service.MediaAssetFilter{
		CategoryID: categoryID,
		Keyword:    values.Get("keyword"),
		Page:       page,
		PageSize:   pageSize,
	})
	writeOK(c, assets, err, h.writeError)
}

func (h *Handler) UploadMediaAsset(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	req := c.Request()
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		result.BadRequest(c, err.Error())
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		result.BadRequest(c, "missing file")
		return
	}
	defer file.Close()
	categoryID, ok := parseInt64Form(c, "categoryId", 0)
	if !ok {
		return
	}
	asset, err := h.service.UploadMediaAsset(c.RequestContext(), service.UploadMediaAssetInput{
		CategoryID:         categoryID,
		Filename:           header.Filename,
		Reader:             file,
		Size:               header.Size,
		UploadedBy:         principal.UserID,
		UploadedByUsername: principal.Username,
	})
	writeCreated(c, asset, err, h.writeError)
}

func (h *Handler) CheckMediaResumableUpload(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req checkMediaResumableUploadRequest
	if !bind(c, &req) {
		return
	}
	check, err := h.service.CheckMediaResumableUpload(c.RequestContext(), service.CheckMediaResumableUploadInput{
		CategoryID:         int64(req.CategoryID),
		ChunkSize:          req.ChunkSize,
		ChunkTotal:         req.ChunkTotal,
		FileHash:           req.FileHash,
		Filename:           req.FileName,
		SizeBytes:          req.SizeBytes,
		UploadedBy:         principal.UserID,
		UploadedByUsername: principal.Username,
	})
	writeOK(c, check, err, h.writeError)
}

func (h *Handler) UploadMediaChunk(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	req := c.Request()
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		result.BadRequest(c, err.Error())
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		result.BadRequest(c, "missing file")
		return
	}
	defer file.Close()
	sessionID, ok := parseInt64Form(c, "sessionId", 0)
	if !ok || sessionID <= 0 {
		result.BadRequest(c, "invalid sessionId")
		return
	}
	chunkIndex, ok := parseIntForm(c, "chunkIndex", -1)
	if !ok || chunkIndex < 0 {
		result.BadRequest(c, "invalid chunkIndex")
		return
	}
	chunkTotal, ok := parseIntForm(c, "chunkTotal", 0)
	if !ok {
		return
	}
	chunk, err := h.service.UploadMediaChunk(c.RequestContext(), service.UploadMediaChunkInput{
		ChunkHash:          req.FormValue("chunkHash"),
		ChunkIndex:         chunkIndex,
		ChunkTotal:         chunkTotal,
		FileHash:           req.FormValue("fileHash"),
		Filename:           req.FormValue("fileName"),
		Reader:             file,
		SessionID:          sessionID,
		Size:               header.Size,
		UploadedBy:         principal.UserID,
		UploadedByUsername: principal.Username,
	})
	writeCreated(c, chunk, err, h.writeError)
}

func (h *Handler) CompleteMediaResumableUpload(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req mediaResumableSessionRequest
	if !bind(c, &req) {
		return
	}
	complete, err := h.service.CompleteMediaResumableUpload(c.RequestContext(), service.CompleteMediaResumableUploadInput{
		FileHash:   req.FileHash,
		SessionID:  int64(req.SessionID),
		UploadedBy: principal.UserID,
	})
	writeCreated(c, complete, err, h.writeError)
}

func (h *Handler) AbortMediaResumableUpload(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req mediaResumableSessionRequest
	if !bind(c, &req) {
		return
	}
	abort, err := h.service.AbortMediaResumableUpload(c.RequestContext(), service.AbortMediaResumableUploadInput{
		FileHash:   req.FileHash,
		SessionID:  int64(req.SessionID),
		UploadedBy: principal.UserID,
	})
	writeOK(c, abort, err, h.writeError)
}

func (h *Handler) ImportMediaURLs(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req importMediaURLsRequest
	if !bind(c, &req) {
		return
	}
	importResult, err := h.service.ImportMediaURLs(c.RequestContext(), service.ImportMediaURLsInput{
		CategoryID:         int64(req.CategoryID),
		Items:              req.items(),
		UploadedBy:         principal.UserID,
		UploadedByUsername: principal.Username,
	})
	writeCreated(c, importResult, err, h.writeError)
}

func (h *Handler) UpdateMediaAsset(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "assetId")
	if !ok {
		return
	}
	var req updateMediaAssetRequest
	if !bind(c, &req) {
		return
	}
	asset, err := h.service.UpdateMediaAsset(c.RequestContext(), id, service.UpdateMediaAssetInput{DisplayName: req.DisplayName})
	writeOK(c, asset, err, h.writeError)
}

func (h *Handler) DownloadMediaAsset(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "assetId")
	if !ok {
		return
	}
	download, err := h.service.DownloadMediaAsset(c.RequestContext(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": download.Filename}))
	c.Header("Content-Length", strconv.Itoa(len(download.Data)))
	c.Data(http.StatusOK, download.ContentType, download.Data)
}

func (h *Handler) DeleteMediaAsset(c ports.HTTPContext) {
	if _, ok := requirePrincipal(c); !ok {
		return
	}
	id, ok := parseInt64Param(c, "assetId")
	if !ok {
		return
	}
	writeOK(c, map[string]bool{"deleted": true}, h.service.DeleteMediaAsset(c.RequestContext(), id), h.writeError)
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

func (h *Handler) writeError(c ports.HTTPContext, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		result.Fail(c, http.StatusRequestTimeout, "request canceled")
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrDuplicate), errors.Is(err, service.ErrExternalMedia):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrNotFound):
		result.NotFound(c, err.Error())
	case errors.Is(err, service.ErrConfigUnavailable), errors.Is(err, service.ErrStorageUnavailable):
		result.Fail(c, http.StatusServiceUnavailable, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("system request failed", "error", err)
		}
		result.InternalError(c, "internal server error")
	}
}

func requirePrincipal(c ports.HTTPContext) (iamservice.Principal, bool) {
	principal, ok := middleware.GetPrincipal(c)
	if !ok {
		result.Unauthorized(c, "missing principal")
		return iamservice.Principal{}, false
	}
	return principal, true
}

func bind(c ports.HTTPContext, dest any) bool {
	if err := c.BindJSON(dest); err != nil {
		result.BadRequest(c, err.Error())
		return false
	}
	return true
}

func parseInt64Param(c ports.HTTPContext, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		result.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return id, true
}

func parseIntQuery(c ports.HTTPContext, name string, fallback int) (int, bool) {
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

func parseInt64Query(c ports.HTTPContext, name string, fallback int64) (int64, bool) {
	raw := strings.TrimSpace(c.Request().URL.Query().Get(name))
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		result.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return value, true
}

func parseInt64Form(c ports.HTTPContext, name string, fallback int64) (int64, bool) {
	raw := strings.TrimSpace(c.Request().FormValue(name))
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		result.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return value, true
}

func parseIntForm(c ports.HTTPContext, name string, fallback int) (int, bool) {
	raw := strings.TrimSpace(c.Request().FormValue(name))
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

func parseTimeQuery(c ports.HTTPContext, name string, endExclusive bool) (*time.Time, bool) {
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

func (r deleteVersionsRequest) int64s() []int64 {
	ids := make([]int64, 0, len(r.IDs))
	for _, id := range r.IDs {
		ids = append(ids, int64(id))
	}
	return ids
}

func (r importMediaURLsRequest) items() []service.MediaURLImportItem {
	items := make([]service.MediaURLImportItem, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, service.MediaURLImportItem{Name: item.Name, URL: item.URL})
	}
	if len(items) > 0 {
		return items
	}
	for _, line := range strings.Split(r.Text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := ""
		rawURL := line
		if left, right, ok := strings.Cut(line, "|"); ok {
			name = strings.TrimSpace(left)
			rawURL = strings.TrimSpace(right)
		}
		items = append(items, service.MediaURLImportItem{Name: name, URL: rawURL})
	}
	return items
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

func writeOK(c ports.HTTPContext, data any, err error, writeError func(ports.HTTPContext, error)) {
	if err != nil {
		writeError(c, err)
		return
	}
	result.OK(c, data)
}

func writeCreated(c ports.HTTPContext, data any, err error, writeError func(ports.HTTPContext, error)) {
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
