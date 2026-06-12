package service

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/system/model"
	"github.com/rei0721/go-scaffold/pkg/database"
)

func TestSyncAPIsPersistsCurrentRoutesAndMarksStaleRecords(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	repo := newMemoryAPIRepo([]model.APIRecord{
		{
			ID:          1,
			Code:        "get /api/v1/system/apis",
			Group:       "system",
			Method:      "GET",
			Path:        "/api/v1/system/apis",
			Description: "old",
			Permission:  "permission:read",
			Status:      model.APIStatusActive,
			Source:      "router",
			SyncedAt:    now.Add(-time.Hour),
			CreatedAt:   now.Add(-time.Hour),
			UpdatedAt:   now.Add(-time.Hour),
		},
		{
			ID:          2,
			Code:        "get /api/v1/old",
			Group:       "system",
			Method:      "GET",
			Path:        "/api/v1/old",
			Description: "old route",
			Status:      model.APIStatusActive,
			Source:      "router",
			SyncedAt:    now.Add(-time.Hour),
			CreatedAt:   now.Add(-time.Hour),
			UpdatedAt:   now.Add(-time.Hour),
		},
	})
	svc := New(Config{Now: func() time.Time { return now }},
		WithRepository(repo),
		WithIDGenerator(&sequenceIDGenerator{next: 100}),
	)
	svc.RegisterAPIs([]model.APIEntry{
		{Code: "get /api/v1/system/apis", Group: "system", Method: "GET", Path: "/api/v1/system/apis", Description: "catalog", Permission: "permission:read", Order: 10},
		{Code: "post /api/v1/system/apis/sync", Group: "system", Method: "POST", Path: "/api/v1/system/apis/sync", Description: "sync", Permission: "permission:read", Order: 20},
	})

	result, err := svc.SyncAPIs(context.Background())
	if err != nil {
		t.Fatalf("SyncAPIs() error = %v", err)
	}
	if !result.Persisted || result.StorageStatus != "persisted" {
		t.Fatalf("expected persisted sync result, got %#v", result)
	}
	if result.Total != 2 || result.Created != 1 || result.Updated != 1 || result.Stale != 1 {
		t.Fatalf("unexpected sync counters: %#v", result)
	}

	oldRecord, ok := repo.record("GET", "/api/v1/old")
	if !ok || oldRecord.Status != model.APIStatusStale {
		t.Fatalf("expected old route to be stale, got %#v", oldRecord)
	}
	created, ok := repo.record("POST", "/api/v1/system/apis/sync")
	if !ok || created.ID != 100 || created.Status != model.APIStatusActive {
		t.Fatalf("expected sync route to be created, got %#v", created)
	}

	groups, err := svc.ListAPIs(context.Background())
	if err != nil {
		t.Fatalf("ListAPIs() error = %v", err)
	}
	if !apiEntrySynced(groups, "GET", "/api/v1/system/apis") || !apiEntrySynced(groups, "POST", "/api/v1/system/apis/sync") {
		t.Fatalf("expected listed API entries to include sync metadata, got %#v", groups)
	}
}

func TestSyncPermissionsRegistersRoutePermissionsAndAnnotatesCatalog(t *testing.T) {
	now := time.Date(2026, 6, 11, 13, 0, 0, 0, time.UTC)
	permissions := newMemoryPermissionStore([]model.PermissionEntry{
		{Code: "permission:read", Name: "Read permissions", Description: "Read permissions"},
	})
	svc := New(Config{Now: func() time.Time { return now }},
		WithPermissionStore(permissions),
	)
	svc.RegisterAPIs([]model.APIEntry{
		{Code: "get /api/v1/system/apis", Group: "system", Method: "GET", Path: "/api/v1/system/apis", Description: "catalog", Permission: "permission:read", Order: 10},
		{Code: "post /api/v1/system/apis/permissions/sync", Group: "system", Method: "POST", Path: "/api/v1/system/apis/permissions/sync", Description: "sync permissions", Permission: "permission:sync", Order: 20},
		{Code: "get /api/v1/system/menus", Group: "system", Method: "GET", Path: "/api/v1/system/menus", Description: "menus", Order: 30},
		{Code: "get /api/v1/broken", Group: "system", Method: "GET", Path: "/api/v1/broken", Description: "broken", Permission: "broken", Order: 40},
	})

	result, err := svc.SyncPermissions(context.Background())
	if err != nil {
		t.Fatalf("SyncPermissions() error = %v", err)
	}
	if !result.Persisted || result.StorageStatus != "persisted" {
		t.Fatalf("expected persisted permission sync, got %#v", result)
	}
	if result.Total != 2 || result.Created != 1 || result.Skipped != 1 {
		t.Fatalf("unexpected permission sync counters: %#v", result)
	}
	if !permissions.has("permission:sync") {
		t.Fatalf("expected permission:sync to be created, got %#v", permissions.records)
	}

	groups, err := svc.ListAPIs(context.Background())
	if err != nil {
		t.Fatalf("ListAPIs() error = %v", err)
	}
	if !apiEntryPermissionRegistered(groups, "GET", "/api/v1/system/apis") {
		t.Fatalf("expected permission:read route to be marked registered: %#v", groups)
	}
	if !apiEntryPermissionRegistered(groups, "POST", "/api/v1/system/apis/permissions/sync") {
		t.Fatalf("expected permission:sync route to be marked registered: %#v", groups)
	}
}

func TestListMenusIncludesSystemMenuCatalog(t *testing.T) {
	svc := New(Config{})
	groups, err := svc.ListMenus(context.Background())
	if err != nil {
		t.Fatalf("ListMenus() error = %v", err)
	}
	if !menuItemExists(groups, "system", "menus", "/menus", "permission:read") {
		t.Fatalf("expected system menu catalog entry, got %#v", groups)
	}
	if !menuItemExists(groups, "system", "apis", "/apis", "permission:read") {
		t.Fatalf("expected system API catalog entry, got %#v", groups)
	}
	if !menuItemExists(groups, "system", "dictionaries", "/dictionaries", "dictionary:read") {
		t.Fatalf("expected system dictionary catalog entry, got %#v", groups)
	}
	if !menuItemExists(groups, "system", "operation-records", "/operation-records", "operation:read") {
		t.Fatalf("expected operation history entry, got %#v", groups)
	}
	if !menuItemExists(groups, "system", "parameters", "/parameters", "parameter:read") {
		t.Fatalf("expected system parameter management entry, got %#v", groups)
	}
	if !menuItemExists(groups, "system", "system-config", "/system", "config:read") {
		t.Fatalf("expected system config entry, got %#v", groups)
	}
	if !menuItemExists(groups, "system", "server-info", "/server-info", "server:read") {
		t.Fatalf("expected server info entry, got %#v", groups)
	}
	if !menuItemExists(groups, "security", "login-logs", "/login-logs", "audit:read") {
		t.Fatalf("expected login log entry, got %#v", groups)
	}
	if !menuItemExists(groups, "security", "error-logs", "/error-logs", "operation:read") {
		t.Fatalf("expected error log entry, got %#v", groups)
	}
}

func TestListConfigReturnsRuntimeSnapshotClone(t *testing.T) {
	svc := New(Config{
		RuntimeConfig: model.ConfigSnapshot{
			Sections: []model.ConfigSection{
				{
					Code:  "server",
					Label: "System",
					Items: []model.ConfigItem{
						{Key: "server.port", Label: "Port", Value: 9999},
					},
				},
			},
		},
	})

	snapshot, err := svc.ListConfig(context.Background())
	if err != nil {
		t.Fatalf("ListConfig() error = %v", err)
	}
	if len(snapshot.Sections) != 1 || len(snapshot.Sections[0].Items) != 1 {
		t.Fatalf("unexpected config snapshot: %#v", snapshot)
	}
	snapshot.Sections[0].Items[0].Value = 10000
	snapshot.Sections[0].Items = append(snapshot.Sections[0].Items, model.ConfigItem{Key: "server.mode"})

	again, err := svc.ListConfig(context.Background())
	if err != nil {
		t.Fatalf("ListConfig() second error = %v", err)
	}
	if again.Sections[0].Items[0].Value != 9999 || len(again.Sections[0].Items) != 1 {
		t.Fatalf("expected stored snapshot to remain unchanged, got %#v", again)
	}
}

func TestGetServerInfoReportsRuntimeAndMemory(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	svc := New(Config{
		Now:       func() time.Time { return now },
		StartTime: now.Add(-time.Hour),
	})

	info, err := svc.GetServerInfo(context.Background())
	if err != nil {
		t.Fatalf("GetServerInfo() error = %v", err)
	}
	if info.OS.GoOS == "" || info.OS.GoArch == "" || info.OS.GoVersion == "" || info.OS.NumCPU <= 0 {
		t.Fatalf("expected runtime OS fields, got %#v", info.OS)
	}
	if info.Runtime.StartTime != now.Add(-time.Hour) || info.Runtime.UptimeSeconds != 3600 || info.Runtime.Uptime == "" {
		t.Fatalf("unexpected runtime info: %#v", info.Runtime)
	}
	if info.Memory.SysMB == 0 || info.Memory.HeapObjects == 0 {
		t.Fatalf("expected memory stats, got %#v", info.Memory)
	}
	if info.CPU.Cores <= 0 {
		t.Fatalf("expected host CPU core count, got %#v", info.CPU)
	}
	for _, value := range info.CPU.Percent {
		if value < 0 {
			t.Fatalf("expected non-negative CPU percentage, got %#v", info.CPU)
		}
	}
	if info.RAM.TotalMB > 0 && info.RAM.UsedMB > info.RAM.TotalMB {
		t.Fatalf("expected RAM usage to fit total, got %#v", info.RAM)
	}
	for _, item := range info.Disk {
		if item.MountPoint == "" || item.TotalMB == 0 || item.UsedPercent < 0 {
			t.Fatalf("expected valid disk entry, got %#v", item)
		}
	}
	if info.Build.GoVersion == "" {
		t.Fatalf("expected build info go version, got %#v", info.Build)
	}
	if !info.RefreshedAt.Equal(now) {
		t.Fatalf("expected refreshedAt %s, got %s", now, info.RefreshedAt)
	}
}

func TestDictionaryManagementCreatesUpdatesAndDeletesDictionariesAndItems(t *testing.T) {
	now := time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)
	repo := newMemoryAPIRepo(nil)
	svc := New(Config{Now: func() time.Time { return now }},
		WithRepository(repo),
		WithIDGenerator(&sequenceIDGenerator{next: 100}),
	)

	dictionary, err := svc.CreateDictionary(context.Background(), CreateDictionaryInput{
		Code:        " Status ",
		Description: "Workflow status",
		Name:        "Status",
	})
	if err != nil {
		t.Fatalf("CreateDictionary() error = %v", err)
	}
	if dictionary.ID != 100 || dictionary.Code != "status" || dictionary.Status != model.DictionaryStatusActive {
		t.Fatalf("unexpected dictionary: %#v", dictionary)
	}

	item, err := svc.CreateDictionaryItem(context.Background(), dictionary.ID, CreateDictionaryItemInput{
		Label: "Enabled",
		Sort:  20,
		Value: "enabled",
	})
	if err != nil {
		t.Fatalf("CreateDictionaryItem() error = %v", err)
	}
	if item.ID != 101 || item.DictionaryID != dictionary.ID || item.Status != model.DictionaryStatusActive {
		t.Fatalf("unexpected dictionary item: %#v", item)
	}

	catalog, err := svc.ListDictionaries(context.Background())
	if err != nil {
		t.Fatalf("ListDictionaries() error = %v", err)
	}
	if catalog.StorageStatus != "persisted" || catalog.Total != 1 || len(catalog.Items[0].Items) != 1 {
		t.Fatalf("unexpected dictionary catalog: %#v", catalog)
	}

	name := "Status Dictionary"
	status := model.DictionaryStatusDisabled
	updated, err := svc.UpdateDictionary(context.Background(), dictionary.ID, UpdateDictionaryInput{Name: &name, Status: &status})
	if err != nil {
		t.Fatalf("UpdateDictionary() error = %v", err)
	}
	if updated.Name != name || updated.Status != model.DictionaryStatusDisabled {
		t.Fatalf("unexpected updated dictionary: %#v", updated)
	}

	label := "Active"
	sortOrder := 5
	updatedItem, err := svc.UpdateDictionaryItem(context.Background(), item.ID, UpdateDictionaryItemInput{Label: &label, Sort: &sortOrder})
	if err != nil {
		t.Fatalf("UpdateDictionaryItem() error = %v", err)
	}
	if updatedItem.Label != label || updatedItem.Sort != sortOrder {
		t.Fatalf("unexpected updated item: %#v", updatedItem)
	}

	if err := svc.DeleteDictionaryItem(context.Background(), item.ID); err != nil {
		t.Fatalf("DeleteDictionaryItem() error = %v", err)
	}
	catalog, err = svc.ListDictionaries(context.Background())
	if err != nil {
		t.Fatalf("ListDictionaries() after item delete error = %v", err)
	}
	if len(catalog.Items[0].Items) != 0 {
		t.Fatalf("expected item to be removed from catalog, got %#v", catalog.Items[0].Items)
	}

	if err := svc.DeleteDictionary(context.Background(), dictionary.ID); err != nil {
		t.Fatalf("DeleteDictionary() error = %v", err)
	}
	catalog, err = svc.ListDictionaries(context.Background())
	if err != nil {
		t.Fatalf("ListDictionaries() after dictionary delete error = %v", err)
	}
	if catalog.Total != 0 {
		t.Fatalf("expected dictionary to be removed from catalog, got %#v", catalog)
	}
}

func TestOperationRecordManagementPersistsFiltersAndDeletesRecords(t *testing.T) {
	now := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
	repo := newMemoryAPIRepo(nil)
	svc := New(Config{Now: func() time.Time { return now }},
		WithRepository(repo),
		WithIDGenerator(&sequenceIDGenerator{next: 300}),
	)

	if err := svc.RecordOperation(context.Background(), OperationRecordInput{
		Body:      `{"name":"aoi"}`,
		IPAddress: "127.0.0.1",
		LatencyMs: 32,
		Method:    "get",
		Path:      "/api/v1/system/menus",
		Status:    200,
		TraceID:   "trace-1",
		UserAgent: "test-agent",
		UserID:    1,
		Username:  "admin",
	}); err != nil {
		t.Fatalf("RecordOperation() error = %v", err)
	}
	if err := svc.RecordOperation(context.Background(), OperationRecordInput{
		IPAddress:    "127.0.0.1",
		Method:       "delete",
		Path:         "/api/v1/system/operation-records",
		Response:     `{"deleted":true}`,
		Status:       204,
		UserID:       1,
		Username:     "admin",
		ErrorMessage: strings.Repeat("x", 9000),
	}); err != nil {
		t.Fatalf("RecordOperation() second error = %v", err)
	}

	page, err := svc.ListOperationRecords(context.Background(), OperationRecordFilter{
		Method:   "DELETE",
		Page:     1,
		PageSize: 10,
		Path:     "operation-records",
		Status:   204,
	})
	if err != nil {
		t.Fatalf("ListOperationRecords() error = %v", err)
	}
	if page.StorageStatus != "persisted" || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected operation record page: %#v", page)
	}
	record := page.Items[0]
	if record.ID != 301 || record.Method != "DELETE" || record.Username != "admin" || record.ErrorMessage == "" {
		t.Fatalf("unexpected operation record: %#v", record)
	}
	if len(record.ErrorMessage) <= 8192 || !strings.Contains(record.ErrorMessage, "truncated") {
		t.Fatalf("expected long operation payload to be truncated, got len=%d", len(record.ErrorMessage))
	}

	if err := svc.DeleteOperationRecords(context.Background(), []int64{record.ID}); err != nil {
		t.Fatalf("DeleteOperationRecords() error = %v", err)
	}
	page, err = svc.ListOperationRecords(context.Background(), OperationRecordFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListOperationRecords() after delete error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != 300 {
		t.Fatalf("expected only first operation record to remain, got %#v", page)
	}
}

func TestOperationRecordStatusClassFilters(t *testing.T) {
	now := time.Date(2026, 6, 12, 10, 30, 0, 0, time.UTC)
	repo := newMemoryAPIRepo(nil)
	svc := New(Config{Now: func() time.Time { return now }},
		WithRepository(repo),
		WithIDGenerator(&sequenceIDGenerator{next: 400}),
	)
	for _, item := range []OperationRecordInput{
		{IPAddress: "127.0.0.1", Method: "GET", Path: "/api/v1/ok", Status: 200, UserID: 1, Username: "admin"},
		{IPAddress: "127.0.0.1", Method: "GET", Path: "/api/v1/not-found", Status: 404, UserID: 1, Username: "admin"},
		{ErrorMessage: "boom", IPAddress: "127.0.0.1", Method: "POST", Path: "/api/v1/error", Status: 503, UserID: 1, Username: "admin"},
	} {
		if err := svc.RecordOperation(context.Background(), item); err != nil {
			t.Fatalf("RecordOperation() error = %v", err)
		}
	}

	page, err := svc.ListOperationRecords(context.Background(), OperationRecordFilter{Page: 1, PageSize: 10, StatusClass: "5xx"})
	if err != nil {
		t.Fatalf("ListOperationRecords(5xx) error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Status != 503 {
		t.Fatalf("expected only 5xx records, got %#v", page)
	}

	page, err = svc.ListOperationRecords(context.Background(), OperationRecordFilter{Page: 1, PageSize: 10, StatusClass: "error"})
	if err != nil {
		t.Fatalf("ListOperationRecords(error) error = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 2 {
		t.Fatalf("expected all error records, got %#v", page)
	}

	page, err = svc.ListOperationRecords(context.Background(), OperationRecordFilter{Page: 1, PageSize: 10, StatusClass: "4xx"})
	if err != nil {
		t.Fatalf("ListOperationRecords(4xx) error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Status != 404 {
		t.Fatalf("expected only 4xx records, got %#v", page)
	}

	page, err = svc.ListOperationRecords(context.Background(), OperationRecordFilter{Page: 1, PageSize: 10, Status: 404, StatusClass: "5xx"})
	if err != nil {
		t.Fatalf("ListOperationRecords(exact status) error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Status != 404 {
		t.Fatalf("expected exact status to win over status class, got %#v", page)
	}

	if _, err = svc.ListOperationRecords(context.Background(), OperationRecordFilter{Page: 1, PageSize: 10, StatusClass: "2xx"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid status class error, got %v", err)
	}
}

func TestParameterManagementCreatesFiltersUpdatesFindsAndDeletes(t *testing.T) {
	now := time.Date(2026, 6, 12, 11, 0, 0, 0, time.UTC)
	repo := newMemoryAPIRepo(nil)
	svc := New(Config{Now: func() time.Time { return now }},
		WithRepository(repo),
		WithIDGenerator(&sequenceIDGenerator{next: 500}),
	)

	parameter, err := svc.CreateParameter(context.Background(), CreateParameterInput{
		Description: "Local site name",
		Key:         "site.name",
		Name:        "Site Name",
		Value:       "Aoi Admin",
	})
	if err != nil {
		t.Fatalf("CreateParameter() error = %v", err)
	}
	if parameter.ID != 500 || parameter.Key != "site.name" || parameter.Value != "Aoi Admin" {
		t.Fatalf("unexpected parameter: %#v", parameter)
	}
	if _, err := svc.CreateParameter(context.Background(), CreateParameterInput{Name: "Duplicate", Key: "site.name", Value: "x"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("expected duplicate error, got %v", err)
	}

	page, err := svc.ListParameters(context.Background(), ParameterFilter{
		Key:            "site",
		Name:           "Site",
		Page:           1,
		PageSize:       10,
		StartCreatedAt: ptrTime(now.Add(-time.Minute)),
		EndCreatedAt:   ptrTime(now.Add(time.Minute)),
	})
	if err != nil {
		t.Fatalf("ListParameters() error = %v", err)
	}
	if page.StorageStatus != "persisted" || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected parameter page: %#v", page)
	}

	found, err := svc.FindParameterByKey(context.Background(), "site.name")
	if err != nil {
		t.Fatalf("FindParameterByKey() error = %v", err)
	}
	if found.ID != parameter.ID {
		t.Fatalf("expected parameter by key, got %#v", found)
	}

	newValue := "Aoi Console"
	newKey := "app.name"
	updated, err := svc.UpdateParameter(context.Background(), parameter.ID, UpdateParameterInput{Key: &newKey, Value: &newValue})
	if err != nil {
		t.Fatalf("UpdateParameter() error = %v", err)
	}
	if updated.Key != newKey || updated.Value != newValue {
		t.Fatalf("unexpected updated parameter: %#v", updated)
	}

	if err := svc.DeleteParameters(context.Background(), []int64{parameter.ID}); err != nil {
		t.Fatalf("DeleteParameters() error = %v", err)
	}
	page, err = svc.ListParameters(context.Background(), ParameterFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListParameters() after delete error = %v", err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("expected parameter to be soft deleted, got %#v", page)
	}
}

func TestSeedDefaultsCreatesSystemDataIdempotently(t *testing.T) {
	now := time.Date(2026, 6, 12, 14, 0, 0, 0, time.UTC)
	repo := newMemoryAPIRepo(nil)
	svc := New(Config{Now: func() time.Time { return now }},
		WithRepository(repo),
		WithIDGenerator(&sequenceIDGenerator{next: 700}),
	)

	result, err := svc.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}
	if result.StorageStatus != "persisted" || result.DictionariesCreated != 3 || result.DictionaryItemsCreated != 9 || result.ParametersCreated != 3 {
		t.Fatalf("unexpected seed result: %#v", result)
	}

	catalog, err := svc.ListDictionaries(context.Background())
	if err != nil {
		t.Fatalf("ListDictionaries() error = %v", err)
	}
	if catalog.Total != 3 || !dictionaryItemExists(catalog, "system.status", model.DictionaryStatusActive) || !dictionaryItemExists(catalog, "http.method", "DELETE") {
		t.Fatalf("expected seeded dictionaries and items, got %#v", catalog)
	}
	title, err := svc.FindParameterByKey(context.Background(), "admin.title")
	if err != nil {
		t.Fatalf("FindParameterByKey(admin.title) error = %v", err)
	}
	customTitle := "Custom Admin"
	if _, err := svc.UpdateParameter(context.Background(), title.ID, UpdateParameterInput{Value: &customTitle}); err != nil {
		t.Fatalf("UpdateParameter(admin.title) error = %v", err)
	}

	again, err := svc.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() second error = %v", err)
	}
	if again.DictionariesCreated != 0 || again.DictionaryItemsCreated != 0 || again.ParametersCreated != 0 {
		t.Fatalf("expected second seed to be idempotent, got %#v", again)
	}
	title, err = svc.FindParameterByKey(context.Background(), "admin.title")
	if err != nil {
		t.Fatalf("FindParameterByKey(admin.title) second error = %v", err)
	}
	if title.Value != customTitle {
		t.Fatalf("expected seed to preserve customized parameter, got %#v", title)
	}
}

func TestSeedDefaultsWithoutRepositoryReportsUnavailable(t *testing.T) {
	svc := New(Config{})

	result, err := svc.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}
	if result.StorageStatus != "unavailable" || result.DictionariesCreated != 0 || result.ParametersCreated != 0 {
		t.Fatalf("unexpected seed result without repository: %#v", result)
	}
}

type memoryAPIRepo struct {
	dictionaries     map[int64]model.Dictionary
	items            map[int64]model.DictionaryItem
	operationRecords map[int64]model.OperationRecord
	parameters       map[int64]model.Parameter
	records          map[string]model.APIRecord
}

func newMemoryAPIRepo(records []model.APIRecord) *memoryAPIRepo {
	repo := &memoryAPIRepo{
		dictionaries:     make(map[int64]model.Dictionary),
		items:            make(map[int64]model.DictionaryItem),
		operationRecords: make(map[int64]model.OperationRecord),
		parameters:       make(map[int64]model.Parameter),
		records:          make(map[string]model.APIRecord, len(records)),
	}
	for _, record := range records {
		repo.records[memoryAPIKey(record.Method, record.Path)] = record
	}
	return repo
}

func (r *memoryAPIRepo) CreateAPI(_ context.Context, api *model.APIRecord) error {
	r.records[memoryAPIKey(api.Method, api.Path)] = *api
	return nil
}

func (r *memoryAPIRepo) CreateDictionary(_ context.Context, dictionary *model.Dictionary) error {
	r.dictionaries[dictionary.ID] = *dictionary
	return nil
}

func (r *memoryAPIRepo) CreateDictionaryItem(_ context.Context, item *model.DictionaryItem) error {
	r.items[item.ID] = *item
	return nil
}

func (r *memoryAPIRepo) CreateOperationRecord(_ context.Context, record *model.OperationRecord) error {
	r.operationRecords[record.ID] = *record
	return nil
}

func (r *memoryAPIRepo) CreateParameter(_ context.Context, parameter *model.Parameter) error {
	r.parameters[parameter.ID] = *parameter
	return nil
}

func (r *memoryAPIRepo) DeleteDictionary(_ context.Context, id int64, deletedAt time.Time) error {
	dictionary, ok := r.dictionaries[id]
	if !ok || dictionary.DeletedAt != nil {
		return database.ErrNotFound
	}
	dictionary.DeletedAt = &deletedAt
	dictionary.UpdatedAt = deletedAt
	r.dictionaries[id] = dictionary
	for itemID, item := range r.items {
		if item.DictionaryID != id || item.DeletedAt != nil {
			continue
		}
		item.DeletedAt = &deletedAt
		item.UpdatedAt = deletedAt
		r.items[itemID] = item
	}
	return nil
}

func (r *memoryAPIRepo) DeleteDictionaryItem(_ context.Context, id int64, deletedAt time.Time) error {
	item, ok := r.items[id]
	if !ok || item.DeletedAt != nil {
		return database.ErrNotFound
	}
	item.DeletedAt = &deletedAt
	item.UpdatedAt = deletedAt
	r.items[id] = item
	return nil
}

func (r *memoryAPIRepo) DeleteOperationRecords(_ context.Context, ids []int64) error {
	for _, id := range ids {
		delete(r.operationRecords, id)
	}
	return nil
}

func (r *memoryAPIRepo) DeleteParameter(_ context.Context, id int64, deletedAt time.Time) error {
	parameter, ok := r.parameters[id]
	if !ok || parameter.DeletedAt != nil {
		return database.ErrNotFound
	}
	parameter.DeletedAt = &deletedAt
	parameter.UpdatedAt = deletedAt
	r.parameters[id] = parameter
	return nil
}

func (r *memoryAPIRepo) DeleteParameters(_ context.Context, ids []int64, deletedAt time.Time) error {
	for _, id := range ids {
		parameter, ok := r.parameters[id]
		if !ok || parameter.DeletedAt != nil {
			continue
		}
		parameter.DeletedAt = &deletedAt
		parameter.UpdatedAt = deletedAt
		r.parameters[id] = parameter
	}
	return nil
}

func (r *memoryAPIRepo) FindAPI(_ context.Context, method string, path string) (*model.APIRecord, error) {
	record, ok := r.record(method, path)
	if !ok {
		return nil, errors.New("not found")
	}
	return &record, nil
}

func (r *memoryAPIRepo) FindDictionaryByCode(_ context.Context, code string) (*model.Dictionary, error) {
	for _, dictionary := range r.dictionaries {
		if dictionary.Code == code && dictionary.DeletedAt == nil {
			return &dictionary, nil
		}
	}
	return nil, database.ErrNotFound
}

func (r *memoryAPIRepo) FindDictionaryByID(_ context.Context, id int64) (*model.Dictionary, error) {
	dictionary, ok := r.dictionaries[id]
	if !ok || dictionary.DeletedAt != nil {
		return nil, database.ErrNotFound
	}
	return &dictionary, nil
}

func (r *memoryAPIRepo) FindDictionaryItemByID(_ context.Context, id int64) (*model.DictionaryItem, error) {
	item, ok := r.items[id]
	if !ok || item.DeletedAt != nil {
		return nil, database.ErrNotFound
	}
	return &item, nil
}

func (r *memoryAPIRepo) FindParameterByID(_ context.Context, id int64) (*model.Parameter, error) {
	parameter, ok := r.parameters[id]
	if !ok || parameter.DeletedAt != nil {
		return nil, database.ErrNotFound
	}
	return &parameter, nil
}

func (r *memoryAPIRepo) FindParameterByKey(_ context.Context, key string) (*model.Parameter, error) {
	for _, parameter := range r.parameters {
		if parameter.Key == key && parameter.DeletedAt == nil {
			return &parameter, nil
		}
	}
	return nil, database.ErrNotFound
}

func (r *memoryAPIRepo) ListAPIs(context.Context) ([]model.APIRecord, error) {
	records := make([]model.APIRecord, 0, len(r.records))
	for _, record := range r.records {
		records = append(records, record)
	}
	return records, nil
}

func (r *memoryAPIRepo) ListDictionaries(context.Context) ([]model.Dictionary, error) {
	dictionaries := make([]model.Dictionary, 0, len(r.dictionaries))
	for _, dictionary := range r.dictionaries {
		if dictionary.DeletedAt != nil {
			continue
		}
		dictionaries = append(dictionaries, dictionary)
	}
	sort.SliceStable(dictionaries, func(i, j int) bool {
		return dictionaries[i].Code < dictionaries[j].Code
	})
	return dictionaries, nil
}

func (r *memoryAPIRepo) ListDictionaryItems(_ context.Context, dictionaryID int64) ([]model.DictionaryItem, error) {
	items := make([]model.DictionaryItem, 0, len(r.items))
	for _, item := range r.items {
		if item.DictionaryID != dictionaryID || item.DeletedAt != nil {
			continue
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].Value < items[j].Value
		}
		return items[i].Sort < items[j].Sort
	})
	return items, nil
}

func (r *memoryAPIRepo) ListOperationRecords(_ context.Context, filter model.OperationRecordFilter) ([]model.OperationRecord, int64, error) {
	records := make([]model.OperationRecord, 0, len(r.operationRecords))
	method := strings.ToUpper(strings.TrimSpace(filter.Method))
	path := strings.TrimSpace(filter.Path)
	for _, record := range r.operationRecords {
		if method != "" && record.Method != method {
			continue
		}
		if path != "" && !strings.Contains(record.Path, path) {
			continue
		}
		if filter.Status > 0 {
			if record.Status != filter.Status {
				continue
			}
		} else {
			switch strings.ToLower(strings.TrimSpace(filter.StatusClass)) {
			case "4xx":
				if record.Status < 400 || record.Status >= 500 {
					continue
				}
			case "5xx":
				if record.Status < 500 || record.Status >= 600 {
					continue
				}
			case "error":
				if record.Status < 400 {
					continue
				}
			}
		}
		records = append(records, record)
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].CreatedAt.Equal(records[j].CreatedAt) {
			return records[i].ID > records[j].ID
		}
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	total := int64(len(records))
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= len(records) {
		return []model.OperationRecord{}, total, nil
	}
	end := start + pageSize
	if end > len(records) {
		end = len(records)
	}
	return records[start:end], total, nil
}

func (r *memoryAPIRepo) ListParameters(_ context.Context, filter model.ParameterFilter) ([]model.Parameter, int64, error) {
	parameters := make([]model.Parameter, 0, len(r.parameters))
	name := strings.TrimSpace(filter.Name)
	key := strings.TrimSpace(filter.Key)
	for _, parameter := range r.parameters {
		if parameter.DeletedAt != nil {
			continue
		}
		if name != "" && !strings.Contains(parameter.Name, name) {
			continue
		}
		if key != "" && !strings.Contains(parameter.Key, key) {
			continue
		}
		if filter.StartCreatedAt != nil && parameter.CreatedAt.Before(*filter.StartCreatedAt) {
			continue
		}
		if filter.EndCreatedAt != nil && !parameter.CreatedAt.Before(*filter.EndCreatedAt) {
			continue
		}
		parameters = append(parameters, parameter)
	}
	sort.SliceStable(parameters, func(i, j int) bool {
		if parameters[i].CreatedAt.Equal(parameters[j].CreatedAt) {
			return parameters[i].ID > parameters[j].ID
		}
		return parameters[i].CreatedAt.After(parameters[j].CreatedAt)
	})
	total := int64(len(parameters))
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= len(parameters) {
		return []model.Parameter{}, total, nil
	}
	end := start + pageSize
	if end > len(parameters) {
		end = len(parameters)
	}
	return parameters[start:end], total, nil
}

func (r *memoryAPIRepo) SaveAPI(_ context.Context, api *model.APIRecord) error {
	r.records[memoryAPIKey(api.Method, api.Path)] = *api
	return nil
}

func (r *memoryAPIRepo) SaveDictionary(_ context.Context, dictionary *model.Dictionary) error {
	if _, ok := r.dictionaries[dictionary.ID]; !ok {
		return database.ErrNotFound
	}
	r.dictionaries[dictionary.ID] = *dictionary
	return nil
}

func (r *memoryAPIRepo) SaveDictionaryItem(_ context.Context, item *model.DictionaryItem) error {
	if _, ok := r.items[item.ID]; !ok {
		return database.ErrNotFound
	}
	r.items[item.ID] = *item
	return nil
}

func (r *memoryAPIRepo) SaveParameter(_ context.Context, parameter *model.Parameter) error {
	if _, ok := r.parameters[parameter.ID]; !ok {
		return database.ErrNotFound
	}
	r.parameters[parameter.ID] = *parameter
	return nil
}

func (r *memoryAPIRepo) record(method string, path string) (model.APIRecord, bool) {
	record, ok := r.records[memoryAPIKey(method, path)]
	return record, ok
}

type sequenceIDGenerator struct {
	next int64
}

func (g *sequenceIDGenerator) NextID() int64 {
	id := g.next
	g.next++
	return id
}

func (g *sequenceIDGenerator) NextIDString() string {
	return strconv.FormatInt(g.NextID(), 10)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func apiEntrySynced(groups []model.APIGroup, method string, path string) bool {
	for _, group := range groups {
		for _, entry := range group.Items {
			if entry.Method == method && entry.Path == path && entry.Synced && entry.SyncedAt != nil {
				return true
			}
		}
	}
	return false
}

func memoryAPIKey(method string, path string) string {
	return method + " " + path
}

type memoryPermissionStore struct {
	records map[string]model.PermissionEntry
}

func newMemoryPermissionStore(records []model.PermissionEntry) *memoryPermissionStore {
	store := &memoryPermissionStore{records: make(map[string]model.PermissionEntry, len(records))}
	for _, record := range records {
		store.records[record.Code] = record
	}
	return store
}

func (s *memoryPermissionStore) CreatePermission(_ context.Context, permission model.PermissionEntry) error {
	s.records[permission.Code] = permission
	return nil
}

func (s *memoryPermissionStore) ListPermissions(context.Context) ([]model.PermissionEntry, error) {
	records := make([]model.PermissionEntry, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	return records, nil
}

func (s *memoryPermissionStore) has(code string) bool {
	_, ok := s.records[code]
	return ok
}

func apiEntryPermissionRegistered(groups []model.APIGroup, method string, path string) bool {
	for _, group := range groups {
		for _, entry := range group.Items {
			if entry.Method == method && entry.Path == path && entry.PermissionRegistered {
				return true
			}
		}
	}
	return false
}

func menuItemExists(groups []model.MenuGroup, groupCode string, itemCode string, path string, permission string) bool {
	for _, group := range groups {
		if group.Code != groupCode {
			continue
		}
		for _, item := range group.Items {
			if item.Code == itemCode && item.Path == path && item.Permission == permission {
				return true
			}
		}
	}
	return false
}

func dictionaryItemExists(catalog model.DictionaryCatalog, code string, value string) bool {
	for _, dictionary := range catalog.Items {
		if dictionary.Code != code {
			continue
		}
		for _, item := range dictionary.Items {
			if item.Value == value {
				return true
			}
		}
	}
	return false
}
