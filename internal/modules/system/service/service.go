package service

import (
	"context"
	"errors"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/system/model"
	"github.com/rei0721/go-scaffold/internal/modules/system/repository"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/hostmetrics"
	"github.com/rei0721/go-scaffold/pkg/utils"
)

type Config struct {
	DemoEnabled   bool
	RuntimeConfig model.ConfigSnapshot
	Now           func() time.Time
	StartTime     time.Time
}

type Service interface {
	CreateDictionary(context.Context, CreateDictionaryInput) (*model.Dictionary, error)
	CreateDictionaryItem(context.Context, int64, CreateDictionaryItemInput) (*model.DictionaryItem, error)
	CreateParameter(context.Context, CreateParameterInput) (*model.Parameter, error)
	DeleteDictionary(context.Context, int64) error
	DeleteDictionaryItem(context.Context, int64) error
	DeleteOperationRecords(context.Context, []int64) error
	DeleteParameter(context.Context, int64) error
	DeleteParameters(context.Context, []int64) error
	FindParameter(context.Context, int64) (*model.Parameter, error)
	FindParameterByKey(context.Context, string) (*model.Parameter, error)
	GetServerInfo(context.Context) (model.ServerInfo, error)
	ListAPIs(context.Context) ([]model.APIGroup, error)
	ListConfig(context.Context) (model.ConfigSnapshot, error)
	ListDictionaries(context.Context) (model.DictionaryCatalog, error)
	ListMenus(context.Context) ([]model.MenuGroup, error)
	ListOperationRecords(context.Context, OperationRecordFilter) (model.OperationRecordPage, error)
	ListParameters(context.Context, ParameterFilter) (model.ParameterPage, error)
	RecordOperation(context.Context, OperationRecordInput) error
	RegisterAPIs([]model.APIEntry)
	SyncAPIs(context.Context) (model.APISyncResult, error)
	SyncPermissions(context.Context) (model.PermissionSyncResult, error)
	UpdateDictionary(context.Context, int64, UpdateDictionaryInput) (*model.Dictionary, error)
	UpdateDictionaryItem(context.Context, int64, UpdateDictionaryItemInput) (*model.DictionaryItem, error)
	UpdateParameter(context.Context, int64, UpdateParameterInput) (*model.Parameter, error)
}

type Option func(*service)

type PermissionStore interface {
	CreatePermission(context.Context, model.PermissionEntry) error
	ListPermissions(context.Context) ([]model.PermissionEntry, error)
}

type CreateDictionaryInput struct {
	Code        string
	Description string
	Name        string
	Status      string
}

type UpdateDictionaryInput struct {
	Description *string
	Name        *string
	Status      *string
}

type CreateDictionaryItemInput struct {
	Extra  string
	Label  string
	Sort   int
	Status string
	Value  string
}

type UpdateDictionaryItemInput struct {
	Extra  *string
	Label  *string
	Sort   *int
	Status *string
	Value  *string
}

type CreateParameterInput struct {
	Description string
	Key         string
	Name        string
	Value       string
}

type UpdateParameterInput struct {
	Description *string
	Key         *string
	Name        *string
	Value       *string
}

type OperationRecordInput struct {
	Body         string
	ErrorMessage string
	IPAddress    string
	LatencyMs    int64
	Method       string
	Path         string
	Response     string
	Status       int
	TraceID      string
	UserAgent    string
	UserID       int64
	Username     string
}

type OperationRecordFilter struct {
	Method   string
	Page     int
	PageSize int
	Path     string
	Status   int
}

type ParameterFilter struct {
	EndCreatedAt   *time.Time
	Key            string
	Name           string
	Page           int
	PageSize       int
	StartCreatedAt *time.Time
}

var (
	ErrDuplicate          = errors.New("system resource already exists")
	ErrInvalidInput       = errors.New("invalid system input")
	ErrNotFound           = errors.New("system resource not found")
	ErrStorageUnavailable = errors.New("system storage unavailable")
)

type service struct {
	cfg             Config
	ids             utils.IDGenerator
	mu              sync.RWMutex
	apis            []model.APIEntry
	repo            repository.Repository
	permissionStore PermissionStore
}

func WithRepository(repo repository.Repository) Option {
	return func(s *service) {
		s.repo = repo
	}
}

func WithIDGenerator(ids utils.IDGenerator) Option {
	return func(s *service) {
		s.ids = ids
	}
}

func WithPermissionStore(store PermissionStore) Option {
	return func(s *service) {
		s.permissionStore = store
	}
}

func New(cfg Config, options ...Option) Service {
	s := &service{cfg: cfg}
	for _, option := range options {
		option(s)
	}
	if s.ids == nil {
		s.ids = utils.DefaultSnowflake()
	}
	return s
}

func (s *service) ListMenus(context.Context) ([]model.MenuGroup, error) {
	groups := cloneGroups(baseMenus)
	if s.cfg.DemoEnabled {
		groups = append(groups, cloneGroup(demoMenu))
	}
	sortMenus(groups)
	return groups, nil
}

func (s *service) ListConfig(context.Context) (model.ConfigSnapshot, error) {
	return cloneConfigSnapshot(s.cfg.RuntimeConfig), nil
}

func (s *service) GetServerInfo(ctx context.Context) (model.ServerInfo, error) {
	now := s.now()
	start := s.cfg.StartTime
	if !start.IsZero() {
		start = start.UTC()
	}
	if start.IsZero() || start.After(now) {
		start = now
	}
	uptime := now.Sub(start)
	if uptime < 0 {
		uptime = 0
	}

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	host := hostmetrics.Collect(ctx)

	info := model.ServerInfo{
		Build: buildInfo(),
		CPU:   mapServerCPU(host.CPU),
		Disk:  mapServerDisks(host.Disk),
		GC: model.ServerGCInfo{
			NextGCMB:     bytesToMB(stats.NextGC),
			NumGC:        stats.NumGC,
			PauseTotalNs: stats.PauseTotalNs,
		},
		Memory: model.ServerMemoryInfo{
			AllocMB:        bytesToMB(stats.Alloc),
			HeapAllocMB:    bytesToMB(stats.HeapAlloc),
			HeapIdleMB:     bytesToMB(stats.HeapIdle),
			HeapInuseMB:    bytesToMB(stats.HeapInuse),
			HeapObjects:    stats.HeapObjects,
			HeapReleasedMB: bytesToMB(stats.HeapReleased),
			HeapSysMB:      bytesToMB(stats.HeapSys),
			StackInuseMB:   bytesToMB(stats.StackInuse),
			StackSysMB:     bytesToMB(stats.StackSys),
			SysMB:          bytesToMB(stats.Sys),
			TotalAllocMB:   bytesToMB(stats.TotalAlloc),
		},
		OS: model.ServerOSInfo{
			Compiler:     runtime.Compiler,
			GoArch:       runtime.GOARCH,
			GoOS:         runtime.GOOS,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		RAM:         mapServerRAM(host.RAM),
		RefreshedAt: now,
		Runtime: model.ServerRuntimeInfo{
			StartTime:     start,
			Uptime:        formatRuntimeDuration(uptime),
			UptimeSeconds: int64(uptime.Seconds()),
		},
	}
	if stats.LastGC > 0 {
		lastGCAt := time.Unix(0, int64(stats.LastGC)).UTC()
		info.GC.LastGCAt = &lastGCAt
	}
	return info, nil
}

func (s *service) ListDictionaries(ctx context.Context) (model.DictionaryCatalog, error) {
	catalog := model.DictionaryCatalog{StorageStatus: "unavailable"}
	if s.repo == nil {
		return catalog, nil
	}
	dictionaries, err := s.repo.ListDictionaries(ctx)
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			return catalog, nil
		}
		return catalog, err
	}
	for i := range dictionaries {
		items, err := s.repo.ListDictionaryItems(ctx, dictionaries[i].ID)
		if err != nil {
			if repository.IsStorageUnavailable(err) {
				return model.DictionaryCatalog{StorageStatus: "unavailable"}, nil
			}
			return catalog, err
		}
		dictionaries[i].Items = items
	}
	catalog.Items = dictionaries
	catalog.StorageStatus = "persisted"
	catalog.Total = len(dictionaries)
	return catalog, nil
}

func (s *service) CreateDictionary(ctx context.Context, input CreateDictionaryInput) (*model.Dictionary, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	code := normalizeDictionaryCode(input.Code)
	name := strings.TrimSpace(input.Name)
	if !validDictionaryCode(code) || name == "" {
		return nil, ErrInvalidInput
	}
	status, err := normalizeDictionaryStatus(input.Status)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.FindDictionaryByCode(ctx, code); err == nil {
		return nil, ErrDuplicate
	} else if !errors.Is(err, database.ErrNotFound) {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	now := s.now()
	dictionary := &model.Dictionary{
		ID:          s.ids.NextID(),
		Code:        code,
		Description: strings.TrimSpace(input.Description),
		Name:        name,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
		Items:       []model.DictionaryItem{},
	}
	if err := s.repo.CreateDictionary(ctx, dictionary); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	return dictionary, nil
}

func (s *service) UpdateDictionary(ctx context.Context, id int64, input UpdateDictionaryInput) (*model.Dictionary, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	dictionary, err := s.repo.FindDictionaryByID(ctx, id)
	if err != nil {
		return nil, mapDictionaryLookupError(err)
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrInvalidInput
		}
		dictionary.Name = name
	}
	if input.Description != nil {
		dictionary.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		status, err := normalizeDictionaryStatus(*input.Status)
		if err != nil {
			return nil, err
		}
		dictionary.Status = status
	}
	if err := s.repo.SaveDictionary(ctx, dictionary); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	items, err := s.repo.ListDictionaryItems(ctx, dictionary.ID)
	if err != nil && !repository.IsStorageUnavailable(err) {
		return nil, err
	}
	dictionary.Items = items
	return dictionary, nil
}

func (s *service) DeleteDictionary(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	if _, err := s.repo.FindDictionaryByID(ctx, id); err != nil {
		return mapDictionaryLookupError(err)
	}
	if err := s.repo.DeleteDictionary(ctx, id, s.now()); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) CreateDictionaryItem(ctx context.Context, dictionaryID int64, input CreateDictionaryItemInput) (*model.DictionaryItem, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	if _, err := s.repo.FindDictionaryByID(ctx, dictionaryID); err != nil {
		return nil, mapDictionaryLookupError(err)
	}
	label := strings.TrimSpace(input.Label)
	value := strings.TrimSpace(input.Value)
	if label == "" || value == "" {
		return nil, ErrInvalidInput
	}
	status, err := normalizeDictionaryStatus(input.Status)
	if err != nil {
		return nil, err
	}
	now := s.now()
	item := &model.DictionaryItem{
		ID:           s.ids.NextID(),
		DictionaryID: dictionaryID,
		Extra:        strings.TrimSpace(input.Extra),
		Label:        label,
		Sort:         input.Sort,
		Status:       status,
		Value:        value,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateDictionaryItem(ctx, item); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	return item, nil
}

func (s *service) UpdateDictionaryItem(ctx context.Context, id int64, input UpdateDictionaryItemInput) (*model.DictionaryItem, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	item, err := s.repo.FindDictionaryItemByID(ctx, id)
	if err != nil {
		return nil, mapDictionaryLookupError(err)
	}
	if input.Label != nil {
		label := strings.TrimSpace(*input.Label)
		if label == "" {
			return nil, ErrInvalidInput
		}
		item.Label = label
	}
	if input.Value != nil {
		value := strings.TrimSpace(*input.Value)
		if value == "" {
			return nil, ErrInvalidInput
		}
		item.Value = value
	}
	if input.Extra != nil {
		item.Extra = strings.TrimSpace(*input.Extra)
	}
	if input.Sort != nil {
		item.Sort = *input.Sort
	}
	if input.Status != nil {
		status, err := normalizeDictionaryStatus(*input.Status)
		if err != nil {
			return nil, err
		}
		item.Status = status
	}
	if err := s.repo.SaveDictionaryItem(ctx, item); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	return item, nil
}

func (s *service) DeleteDictionaryItem(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	if _, err := s.repo.FindDictionaryItemByID(ctx, id); err != nil {
		return mapDictionaryLookupError(err)
	}
	if err := s.repo.DeleteDictionaryItem(ctx, id, s.now()); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) ListParameters(ctx context.Context, input ParameterFilter) (model.ParameterPage, error) {
	page := normalizePage(input.Page)
	pageSize := normalizePageSize(input.PageSize)
	result := model.ParameterPage{Page: page, PageSize: pageSize, StorageStatus: "unavailable"}
	if s.repo == nil {
		return result, nil
	}
	if input.StartCreatedAt != nil && input.EndCreatedAt != nil && !input.StartCreatedAt.Before(*input.EndCreatedAt) {
		return result, ErrInvalidInput
	}
	parameters, total, err := s.repo.ListParameters(ctx, model.ParameterFilter{
		EndCreatedAt:   input.EndCreatedAt,
		Key:            strings.TrimSpace(input.Key),
		Name:           strings.TrimSpace(input.Name),
		Page:           page,
		PageSize:       pageSize,
		StartCreatedAt: input.StartCreatedAt,
	})
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			return result, nil
		}
		return result, err
	}
	result.Items = parameters
	result.StorageStatus = "persisted"
	result.Total = total
	return result, nil
}

func (s *service) CreateParameter(ctx context.Context, input CreateParameterInput) (*model.Parameter, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	name := strings.TrimSpace(input.Name)
	key := strings.TrimSpace(input.Key)
	value := strings.TrimSpace(input.Value)
	if name == "" || key == "" || value == "" {
		return nil, ErrInvalidInput
	}
	if _, err := s.repo.FindParameterByKey(ctx, key); err == nil {
		return nil, ErrDuplicate
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, mapParameterLookupError(err)
	}
	now := s.now()
	parameter := &model.Parameter{
		ID:          s.ids.NextID(),
		CreatedAt:   now,
		Description: strings.TrimSpace(input.Description),
		Key:         key,
		Name:        name,
		UpdatedAt:   now,
		Value:       value,
	}
	if err := s.repo.CreateParameter(ctx, parameter); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	return parameter, nil
}

func (s *service) UpdateParameter(ctx context.Context, id int64, input UpdateParameterInput) (*model.Parameter, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	parameter, err := s.repo.FindParameterByID(ctx, id)
	if err != nil {
		return nil, mapParameterLookupError(err)
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrInvalidInput
		}
		parameter.Name = name
	}
	if input.Key != nil {
		key := strings.TrimSpace(*input.Key)
		if key == "" {
			return nil, ErrInvalidInput
		}
		if key != parameter.Key {
			existing, err := s.repo.FindParameterByKey(ctx, key)
			if err == nil && existing.ID != parameter.ID {
				return nil, ErrDuplicate
			}
			if err != nil && !errors.Is(err, database.ErrNotFound) {
				return nil, mapParameterLookupError(err)
			}
		}
		parameter.Key = key
	}
	if input.Value != nil {
		value := strings.TrimSpace(*input.Value)
		if value == "" {
			return nil, ErrInvalidInput
		}
		parameter.Value = value
	}
	if input.Description != nil {
		parameter.Description = strings.TrimSpace(*input.Description)
	}
	if err := s.repo.SaveParameter(ctx, parameter); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	return parameter, nil
}

func (s *service) DeleteParameter(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	if _, err := s.repo.FindParameterByID(ctx, id); err != nil {
		return mapParameterLookupError(err)
	}
	if err := s.repo.DeleteParameter(ctx, id, s.now()); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) DeleteParameters(ctx context.Context, ids []int64) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	if len(ids) == 0 {
		return ErrInvalidInput
	}
	normalized := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return ErrInvalidInput
		}
		normalized = append(normalized, id)
	}
	if err := s.repo.DeleteParameters(ctx, normalized, s.now()); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) FindParameter(ctx context.Context, id int64) (*model.Parameter, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	parameter, err := s.repo.FindParameterByID(ctx, id)
	if err != nil {
		return nil, mapParameterLookupError(err)
	}
	return parameter, nil
}

func (s *service) FindParameterByKey(ctx context.Context, key string) (*model.Parameter, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidInput
	}
	parameter, err := s.repo.FindParameterByKey(ctx, key)
	if err != nil {
		return nil, mapParameterLookupError(err)
	}
	return parameter, nil
}

func (s *service) RecordOperation(ctx context.Context, input OperationRecordInput) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	method := strings.ToUpper(strings.TrimSpace(input.Method))
	path := strings.TrimSpace(input.Path)
	if method == "" || path == "" {
		return ErrInvalidInput
	}
	now := s.now()
	record := &model.OperationRecord{
		ID:           s.ids.NextID(),
		Body:         trimOperationPayload(input.Body),
		CreatedAt:    now,
		ErrorMessage: trimOperationPayload(input.ErrorMessage),
		IPAddress:    strings.TrimSpace(input.IPAddress),
		LatencyMs:    input.LatencyMs,
		Method:       method,
		Path:         path,
		Response:     trimOperationPayload(input.Response),
		Status:       input.Status,
		TraceID:      strings.TrimSpace(input.TraceID),
		UserAgent:    trimOperationPayload(input.UserAgent),
		UserID:       input.UserID,
		Username:     strings.TrimSpace(input.Username),
	}
	if err := s.repo.CreateOperationRecord(ctx, record); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) ListOperationRecords(ctx context.Context, input OperationRecordFilter) (model.OperationRecordPage, error) {
	page := normalizePage(input.Page)
	pageSize := normalizePageSize(input.PageSize)
	result := model.OperationRecordPage{Page: page, PageSize: pageSize, StorageStatus: "unavailable"}
	if s.repo == nil {
		return result, nil
	}
	status := input.Status
	if status < 0 || status > 999 {
		return result, ErrInvalidInput
	}
	records, total, err := s.repo.ListOperationRecords(ctx, model.OperationRecordFilter{
		Method:   strings.ToUpper(strings.TrimSpace(input.Method)),
		Page:     page,
		PageSize: pageSize,
		Path:     strings.TrimSpace(input.Path),
		Status:   status,
	})
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			return result, nil
		}
		return result, err
	}
	result.Items = records
	result.StorageStatus = "persisted"
	result.Total = total
	return result, nil
}

func (s *service) DeleteOperationRecords(ctx context.Context, ids []int64) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	if len(ids) == 0 {
		return ErrInvalidInput
	}
	normalized := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return ErrInvalidInput
		}
		normalized = append(normalized, id)
	}
	if err := s.repo.DeleteOperationRecords(ctx, normalized); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) RegisterAPIs(entries []model.APIEntry) {
	cloned := append([]model.APIEntry(nil), entries...)
	sortAPIs(cloned)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.apis = cloned
}

func (s *service) ListAPIs(ctx context.Context) ([]model.APIGroup, error) {
	s.mu.RLock()
	entries := append([]model.APIEntry(nil), s.apis...)
	s.mu.RUnlock()
	sortAPIs(entries)
	if err := s.applySyncMetadata(ctx, entries); err != nil {
		return nil, err
	}
	if err := s.applyPermissionMetadata(ctx, entries); err != nil {
		return nil, err
	}

	return groupAPIs(entries), nil
}

func (s *service) SyncAPIs(ctx context.Context) (model.APISyncResult, error) {
	s.mu.RLock()
	entries := append([]model.APIEntry(nil), s.apis...)
	s.mu.RUnlock()
	sortAPIs(entries)

	now := s.now()
	result := model.APISyncResult{
		Groups:        groupAPIs(entries),
		StorageStatus: "memory",
		SyncedAt:      now,
		Total:         len(entries),
	}
	if s.repo == nil {
		return result, nil
	}

	existing, err := s.repo.ListAPIs(ctx)
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			result.StorageStatus = "unavailable"
			return result, nil
		}
		return result, err
	}

	existingByKey := make(map[string]*model.APIRecord, len(existing))
	seen := make(map[string]struct{}, len(entries))
	for i := range existing {
		existingByKey[apiKey(existing[i].Method, existing[i].Path)] = &existing[i]
	}

	for _, entry := range entries {
		key := apiKey(entry.Method, entry.Path)
		seen[key] = struct{}{}
		record, ok := existingByKey[key]
		if !ok {
			record = &model.APIRecord{
				ID:        s.ids.NextID(),
				CreatedAt: now,
			}
			result.Created++
		} else {
			result.Updated++
		}
		applyEntryToRecord(record, entry, now)
		if ok {
			if err := s.repo.SaveAPI(ctx, record); err != nil {
				return result, err
			}
			continue
		}
		if err := s.repo.CreateAPI(ctx, record); err != nil {
			return result, err
		}
	}

	for _, record := range existing {
		if _, ok := seen[apiKey(record.Method, record.Path)]; ok || record.Status == model.APIStatusStale {
			continue
		}
		record.Status = model.APIStatusStale
		record.UpdatedAt = now
		if err := s.repo.SaveAPI(ctx, &record); err != nil {
			return result, err
		}
		result.Stale++
	}

	synced := make(map[string]model.APIRecord, len(entries))
	for _, entry := range entries {
		synced[apiKey(entry.Method, entry.Path)] = model.APIRecord{Status: model.APIStatusActive, SyncedAt: now}
	}
	annotated := append([]model.APIEntry(nil), entries...)
	applySyncMetadataFromRecords(annotated, synced)
	if err := s.applyPermissionMetadata(ctx, annotated); err != nil {
		return result, err
	}
	result.Groups = groupAPIs(annotated)
	result.Persisted = true
	result.StorageStatus = "persisted"
	return result, nil
}

func (s *service) SyncPermissions(ctx context.Context) (model.PermissionSyncResult, error) {
	s.mu.RLock()
	entries := append([]model.APIEntry(nil), s.apis...)
	s.mu.RUnlock()

	now := s.now()
	specs := permissionSpecsFromAPIs(entries)
	result := model.PermissionSyncResult{
		Items:         make([]model.PermissionSyncItem, 0, len(specs)),
		StorageStatus: "unavailable",
		SyncedAt:      now,
		Total:         len(specs),
	}
	if s.permissionStore == nil {
		return result, nil
	}

	existing, err := s.permissionStore.ListPermissions(ctx)
	if err != nil {
		return result, err
	}
	existingByCode := make(map[string]model.PermissionEntry, len(existing))
	for _, permission := range existing {
		existingByCode[normalizePermissionCode(permission.Code)] = permission
	}

	for _, spec := range specs {
		item := model.PermissionSyncItem{
			Code:        spec.Code,
			Description: spec.Description,
			Name:        spec.Name,
		}
		if _, ok := existingByCode[spec.Code]; ok {
			item.Exists = true
			result.Skipped++
			result.Items = append(result.Items, item)
			continue
		}
		if err := s.permissionStore.CreatePermission(ctx, spec); err != nil {
			return result, err
		}
		item.Created = true
		result.Created++
		result.Items = append(result.Items, item)
	}
	result.Persisted = true
	result.StorageStatus = "persisted"
	return result, nil
}

func sortMenus(groups []model.MenuGroup) {
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Order == groups[j].Order {
			return groups[i].Code < groups[j].Code
		}
		return groups[i].Order < groups[j].Order
	})
	for i := range groups {
		sort.SliceStable(groups[i].Items, func(a, b int) bool {
			if groups[i].Items[a].Order == groups[i].Items[b].Order {
				return groups[i].Items[a].Code < groups[i].Items[b].Code
			}
			return groups[i].Items[a].Order < groups[i].Items[b].Order
		})
	}
}

func sortAPIs(entries []model.APIEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Group != entries[j].Group {
			return entries[i].Group < entries[j].Group
		}
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		if entries[i].Order == entries[j].Order {
			return entries[i].Method < entries[j].Method
		}
		return entries[i].Order < entries[j].Order
	})
}

func groupAPIs(entries []model.APIEntry) []model.APIGroup {
	byGroup := make(map[string][]model.APIEntry)
	for _, entry := range entries {
		group := normalizeGroup(entry.Group)
		if group == "" {
			group = "other"
		}
		entry.Group = group
		byGroup[group] = append(byGroup[group], entry)
	}

	groups := make([]model.APIGroup, 0, len(byGroup))
	for group, items := range byGroup {
		groups = append(groups, model.APIGroup{
			Code:  group,
			Label: apiGroupLabel(group),
			Count: len(items),
			Items: items,
		})
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if apiGroupOrder(groups[i].Code) == apiGroupOrder(groups[j].Code) {
			return groups[i].Code < groups[j].Code
		}
		return apiGroupOrder(groups[i].Code) < apiGroupOrder(groups[j].Code)
	})
	return groups
}

func (s *service) applySyncMetadata(ctx context.Context, entries []model.APIEntry) error {
	if s.repo == nil {
		return nil
	}
	records, err := s.repo.ListAPIs(ctx)
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil
		}
		return err
	}
	byKey := make(map[string]model.APIRecord, len(records))
	for _, record := range records {
		byKey[apiKey(record.Method, record.Path)] = record
	}
	applySyncMetadataFromRecords(entries, byKey)
	return nil
}

func (s *service) applyPermissionMetadata(ctx context.Context, entries []model.APIEntry) error {
	if s.permissionStore == nil {
		return nil
	}
	permissions, err := s.permissionStore.ListPermissions(ctx)
	if err != nil {
		return err
	}
	registered := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		if code := normalizePermissionCode(permission.Code); code != "" {
			registered[code] = struct{}{}
		}
	}
	for i := range entries {
		code := normalizePermissionCode(entries[i].Permission)
		if code == "" {
			continue
		}
		_, entries[i].PermissionRegistered = registered[code]
	}
	return nil
}

func applySyncMetadataFromRecords(entries []model.APIEntry, records map[string]model.APIRecord) {
	for i := range entries {
		record, ok := records[apiKey(entries[i].Method, entries[i].Path)]
		if !ok || record.Status != model.APIStatusActive {
			continue
		}
		syncedAt := record.SyncedAt
		entries[i].Synced = true
		entries[i].SyncedAt = &syncedAt
	}
}

func applyEntryToRecord(record *model.APIRecord, entry model.APIEntry, now time.Time) {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.Code = entry.Code
	record.Group = normalizeGroup(entry.Group)
	record.Method = strings.ToUpper(strings.TrimSpace(entry.Method))
	record.Path = entry.Path
	record.Description = entry.Description
	record.Permission = entry.Permission
	record.Status = model.APIStatusActive
	record.Source = "router"
	record.SyncedAt = now
	record.UpdatedAt = now
}

func permissionSpecsFromAPIs(entries []model.APIEntry) []model.PermissionEntry {
	byCode := make(map[string]model.PermissionEntry)
	for _, entry := range entries {
		code := normalizePermissionCode(entry.Permission)
		if code == "" || !validPermissionCode(code) {
			continue
		}
		if _, ok := byCode[code]; ok {
			continue
		}
		byCode[code] = model.PermissionEntry{
			Code:        code,
			Name:        permissionName(code),
			Description: permissionDescription(entry),
		}
	}
	specs := make([]model.PermissionEntry, 0, len(byCode))
	for _, spec := range byCode {
		specs = append(specs, spec)
	}
	sort.SliceStable(specs, func(i, j int) bool {
		return specs[i].Code < specs[j].Code
	})
	return specs
}

func normalizePermissionCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validPermissionCode(code string) bool {
	obj, act, ok := strings.Cut(code, ":")
	return ok && strings.TrimSpace(obj) != "" && strings.TrimSpace(act) != ""
}

func normalizeDictionaryCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validDictionaryCode(code string) bool {
	if code == "" {
		return false
	}
	for _, char := range code {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '_' || char == '-' || char == ':' || char == '.':
		default:
			return false
		}
	}
	return true
}

func normalizeDictionaryStatus(value string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == "" {
		return model.DictionaryStatusActive, nil
	}
	switch status {
	case model.DictionaryStatusActive, model.DictionaryStatusDisabled:
		return status, nil
	default:
		return "", ErrInvalidInput
	}
}

func mapDictionaryLookupError(err error) error {
	switch {
	case errors.Is(err, database.ErrNotFound):
		return ErrNotFound
	case repository.IsStorageUnavailable(err):
		return ErrStorageUnavailable
	default:
		return err
	}
}

func mapParameterLookupError(err error) error {
	switch {
	case errors.Is(err, database.ErrNotFound):
		return ErrNotFound
	case repository.IsStorageUnavailable(err):
		return ErrStorageUnavailable
	default:
		return err
	}
}

func normalizePage(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

func normalizePageSize(value int) int {
	if value < 1 {
		return 10
	}
	if value > 100 {
		return 100
	}
	return value
}

func trimOperationPayload(value string) string {
	value = strings.TrimSpace(value)
	const maxOperationPayloadBytes = 8192
	if len(value) <= maxOperationPayloadBytes {
		return value
	}
	return value[:maxOperationPayloadBytes] + "...(truncated)"
}

func permissionName(code string) string {
	obj, act, ok := strings.Cut(code, ":")
	if !ok {
		return code
	}
	return strings.ToUpper(obj[:1]) + obj[1:] + " " + act
}

func permissionDescription(entry model.APIEntry) string {
	if entry.Description != "" {
		return entry.Description
	}
	return entry.Method + " " + entry.Path
}

func apiKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func (s *service) now() time.Time {
	if s.cfg.Now != nil {
		return s.cfg.Now().UTC()
	}
	return time.Now().UTC()
}

func normalizeGroup(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func apiGroupLabel(code string) string {
	switch code {
	case "auth":
		return "认证"
	case "demo":
		return "示例"
	case "orgs":
		return "组织/IAM"
	case "plugins":
		return "插件"
	case "system":
		return "系统"
	default:
		return code
	}
}

func apiGroupOrder(code string) int {
	switch code {
	case "auth":
		return 10
	case "orgs":
		return 20
	case "system":
		return 30
	case "plugins":
		return 40
	case "demo":
		return 90
	default:
		return 100
	}
}

func cloneGroups(src []model.MenuGroup) []model.MenuGroup {
	out := make([]model.MenuGroup, 0, len(src))
	for _, group := range src {
		out = append(out, cloneGroup(group))
	}
	return out
}

func cloneGroup(src model.MenuGroup) model.MenuGroup {
	dst := src
	dst.Items = append([]model.MenuItem(nil), src.Items...)
	return dst
}

func cloneConfigSnapshot(src model.ConfigSnapshot) model.ConfigSnapshot {
	dst := model.ConfigSnapshot{
		Sections: make([]model.ConfigSection, 0, len(src.Sections)),
	}
	for _, section := range src.Sections {
		cloned := section
		cloned.Items = append([]model.ConfigItem(nil), section.Items...)
		dst.Sections = append(dst.Sections, cloned)
	}
	return dst
}

func buildInfo() model.ServerBuildInfo {
	out := model.ServerBuildInfo{
		GoVersion: runtime.Version(),
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return out
	}
	out.GoVersion = info.GoVersion
	out.Module = info.Main.Path
	out.Path = info.Path
	out.Version = info.Main.Version
	out.Settings = make([]model.ServerBuildSetting, 0, len(info.Settings))
	for _, setting := range info.Settings {
		key := strings.TrimSpace(setting.Key)
		if key == "" {
			continue
		}
		out.Settings = append(out.Settings, model.ServerBuildSetting{
			Key:   key,
			Value: setting.Value,
		})
	}
	sort.SliceStable(out.Settings, func(i, j int) bool {
		return out.Settings[i].Key < out.Settings[j].Key
	})
	return out
}

func bytesToMB(value uint64) uint64 {
	const bytesPerMB = 1024 * 1024
	return value / bytesPerMB
}

func mapServerCPU(src hostmetrics.CPUInfo) model.ServerCPUInfo {
	return model.ServerCPUInfo{
		Cores:   src.Cores,
		Percent: append([]float64(nil), src.Percent...),
	}
}

func mapServerRAM(src hostmetrics.RAMInfo) model.ServerRAMInfo {
	return model.ServerRAMInfo{
		TotalMB:     src.TotalMB,
		UsedMB:      src.UsedMB,
		UsedPercent: src.UsedPercent,
	}
}

func mapServerDisks(src []hostmetrics.DiskInfo) []model.ServerDiskInfo {
	out := make([]model.ServerDiskInfo, 0, len(src))
	for _, item := range src {
		out = append(out, model.ServerDiskInfo{
			FSType:      item.FSType,
			MountPoint:  item.MountPoint,
			TotalGB:     item.TotalGB,
			TotalMB:     item.TotalMB,
			UsedGB:      item.UsedGB,
			UsedMB:      item.UsedMB,
			UsedPercent: item.UsedPercent,
		})
	}
	return out
}

func formatRuntimeDuration(value time.Duration) string {
	if value <= 0 {
		return "0s"
	}
	return value.Truncate(time.Second).String()
}

var baseMenus = []model.MenuGroup{
	{
		Code:  "workspace",
		Label: "工作台",
		Order: 10,
		Items: []model.MenuItem{
			{Code: "dashboard", Label: "仪表盘", Icon: "layout-dashboard", Path: "/", Mobile: true, Order: 10},
			{Code: "organizations", Label: "组织", Icon: "building-2", Path: "/organizations", Permission: "org:read", Mobile: true, Order: 20},
			{Code: "users", Label: "用户", Icon: "users", Path: "/users", Permission: "user:read", Mobile: true, Order: 30},
			{Code: "roles", Label: "角色权限", Icon: "shield-check", Path: "/roles", Permission: "role:read", Mobile: true, Order: 40},
		},
	},
	{
		Code:  "security",
		Label: "安全审计",
		Order: 20,
		Items: []model.MenuItem{
			{Code: "sessions", Label: "会话", Icon: "monitor-check", Path: "/sessions", Permission: "session:read", Order: 10},
			{Code: "audit-logs", Label: "审计日志", Icon: "scroll-text", Path: "/audit-logs", Permission: "audit:read", Order: 20},
			{Code: "security", Label: "安全", Icon: "lock-keyhole", Path: "/security", Order: 30},
		},
	},
	{
		Code:  "system",
		Label: "系统管理",
		Order: 30,
		Items: []model.MenuItem{
			{Code: "menus", Label: "菜单管理", Icon: "panel-left", Path: "/menus", Permission: "permission:read", Order: 10},
			{Code: "apis", Label: "API 管理", Icon: "code-2", Path: "/apis", Permission: "permission:read", Order: 20},
			{Code: "dictionaries", Label: "字典管理", Icon: "book-open", Path: "/dictionaries", Permission: "dictionary:read", Order: 30},
			{Code: "operation-records", Label: "操作历史", Icon: "history", Path: "/operation-records", Permission: "operation:read", Order: 40},
			{Code: "parameters", Label: "参数管理", Icon: "compass", Path: "/parameters", Permission: "parameter:read", Order: 50},
			{Code: "system-config", Label: "系统配置", Icon: "settings", Path: "/system", Permission: "config:read", Order: 60},
			{Code: "server-info", Label: "服务器状态", Icon: "activity", Path: "/server-info", Permission: "server:read", Order: 70},
		},
	},
}

var demoMenu = model.MenuGroup{
	Code:  "examples",
	Label: "示例",
	Order: 90,
	Items: []model.MenuItem{
		{Code: "todos", Label: "Demo Todo", Icon: "list-checks", Path: "/todos", Order: 10},
	},
}
