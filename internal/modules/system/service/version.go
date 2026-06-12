package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/system/model"
	"github.com/rei0721/go-scaffold/internal/modules/system/repository"
	"github.com/rei0721/go-scaffold/pkg/database"
)

type VersionFilter struct {
	EndCreatedAt   *time.Time
	Page           int
	PageSize       int
	StartCreatedAt *time.Time
	VersionCode    string
	VersionName    string
}

type ExportVersionInput struct {
	APICodes        []string
	CreatedBy       int64
	CreatorUsername string
	Description     string
	DictionaryCodes []string
	MenuCodes       []string
	VersionCode     string
	VersionName     string
}

type ImportVersionInput struct {
	CreatedBy       int64
	CreatorUsername string
	VersionData     string
}

func (s *service) ListVersionSources(ctx context.Context) (model.VersionSourceCatalog, error) {
	menus, err := s.ListMenus(ctx)
	if err != nil {
		return model.VersionSourceCatalog{}, err
	}
	apis, err := s.ListAPIs(ctx)
	if err != nil {
		return model.VersionSourceCatalog{}, err
	}
	catalog, err := s.ListDictionaries(ctx)
	if err != nil {
		return model.VersionSourceCatalog{}, err
	}
	return model.VersionSourceCatalog{
		APICount:        countAPIs(apis),
		APIs:            apis,
		Dictionaries:    catalog.Items,
		DictionaryCount: len(catalog.Items),
		MenuCount:       countMenus(menus),
		Menus:           menus,
		StorageStatus:   catalog.StorageStatus,
	}, nil
}

func (s *service) ListVersions(ctx context.Context, input VersionFilter) (model.VersionPage, error) {
	page := normalizePage(input.Page)
	pageSize := normalizePageSize(input.PageSize)
	result := model.VersionPage{Page: page, PageSize: pageSize, StorageStatus: "unavailable"}
	if s.repo == nil {
		return result, nil
	}
	if input.StartCreatedAt != nil && input.EndCreatedAt != nil && !input.StartCreatedAt.Before(*input.EndCreatedAt) {
		return result, ErrInvalidInput
	}
	versions, total, err := s.repo.ListVersions(ctx, model.VersionFilter{
		EndCreatedAt:   input.EndCreatedAt,
		Page:           page,
		PageSize:       pageSize,
		StartCreatedAt: input.StartCreatedAt,
		VersionCode:    strings.TrimSpace(input.VersionCode),
		VersionName:    strings.TrimSpace(input.VersionName),
	})
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			return result, nil
		}
		return result, err
	}
	result.Items = versions
	result.StorageStatus = "persisted"
	result.Total = total
	return result, nil
}

func (s *service) FindVersion(ctx context.Context, id int64) (*model.VersionDetail, error) {
	version, err := s.findVersionRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	pkg, err := decodeVersionPackage(version.VersionData)
	if err != nil {
		return nil, err
	}
	return &model.VersionDetail{Item: *version, Package: pkg}, nil
}

func (s *service) GetVersionPackage(ctx context.Context, id int64) (model.VersionPackage, error) {
	version, err := s.findVersionRecord(ctx, id)
	if err != nil {
		return model.VersionPackage{}, err
	}
	return decodeVersionPackage(version.VersionData)
}

func (s *service) ExportVersion(ctx context.Context, input ExportVersionInput) (*model.VersionDetail, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	name := strings.TrimSpace(input.VersionName)
	code := normalizeVersionCode(input.VersionCode)
	if name == "" || !validVersionCode(code) {
		return nil, ErrInvalidInput
	}
	pkg, err := s.buildVersionPackage(ctx, ExportVersionInput{
		APICodes:        input.APICodes,
		CreatedBy:       input.CreatedBy,
		CreatorUsername: strings.TrimSpace(input.CreatorUsername),
		Description:     strings.TrimSpace(input.Description),
		DictionaryCodes: input.DictionaryCodes,
		MenuCodes:       input.MenuCodes,
		VersionCode:     code,
		VersionName:     name,
	})
	if err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, err
	}
	now := s.now()
	version := &model.Version{
		ID:                s.ids.NextID(),
		APICount:          len(pkg.APIs),
		CreatedAt:         now,
		CreatedBy:         input.CreatedBy,
		CreatedByUsername: strings.TrimSpace(input.CreatorUsername),
		Description:       strings.TrimSpace(input.Description),
		DictionaryCount:   len(pkg.Dictionaries),
		MenuCount:         countMenus(pkg.Menus),
		Source:            model.VersionSourceExport,
		UpdatedAt:         now,
		VersionCode:       code,
		VersionData:       string(raw),
		VersionName:       name,
	}
	if err := s.repo.CreateVersion(ctx, version); err != nil {
		if repository.IsStorageUnavailable(err) {
			return nil, ErrStorageUnavailable
		}
		return nil, err
	}
	return &model.VersionDetail{Item: *version, Package: pkg}, nil
}

func (s *service) ImportVersion(ctx context.Context, input ImportVersionInput) (model.VersionImportResult, error) {
	result := model.VersionImportResult{
		ImportedAt:    s.now(),
		StorageStatus: "unavailable",
		APIsSkipped:   0,
		MenusSkipped:  0,
	}
	if s.repo == nil {
		return result, ErrStorageUnavailable
	}
	pkg, err := decodeVersionPackage(input.VersionData)
	if err != nil {
		return result, err
	}
	result.MenusSkipped = countMenus(pkg.Menus)
	result.APIsSkipped = len(pkg.APIs)
	createdDictionaries, skippedDictionaries, createdItems, err := s.importVersionDictionaries(ctx, pkg.Dictionaries)
	if err != nil {
		if repository.IsStorageUnavailable(err) {
			return result, ErrStorageUnavailable
		}
		return result, err
	}
	result.DictionariesCreated = createdDictionaries
	result.DictionariesSkipped = skippedDictionaries
	result.DictionaryItemsCreated = createdItems

	raw, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return result, err
	}
	now := s.now()
	version := &model.Version{
		ID:                s.ids.NextID(),
		APICount:          len(pkg.APIs),
		CreatedAt:         now,
		CreatedBy:         input.CreatedBy,
		CreatedByUsername: strings.TrimSpace(input.CreatorUsername),
		Description:       strings.TrimSpace(pkg.Version.Description),
		DictionaryCount:   len(pkg.Dictionaries),
		MenuCount:         countMenus(pkg.Menus),
		Source:            model.VersionSourceImport,
		UpdatedAt:         now,
		VersionCode:       normalizeVersionCode(pkg.Version.Code),
		VersionData:       string(raw),
		VersionName:       strings.TrimSpace(pkg.Version.Name),
	}
	if err := s.repo.CreateVersion(ctx, version); err != nil {
		if repository.IsStorageUnavailable(err) {
			return result, ErrStorageUnavailable
		}
		return result, err
	}
	result.Item = *version
	result.StorageStatus = "persisted"
	return result, nil
}

func (s *service) DeleteVersion(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrStorageUnavailable
	}
	if _, err := s.repo.FindVersionByID(ctx, id); err != nil {
		return mapVersionLookupError(err)
	}
	if err := s.repo.DeleteVersion(ctx, id, s.now()); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) DeleteVersions(ctx context.Context, ids []int64) error {
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
	if err := s.repo.DeleteVersions(ctx, normalized, s.now()); err != nil {
		if repository.IsStorageUnavailable(err) {
			return ErrStorageUnavailable
		}
		return err
	}
	return nil
}

func (s *service) buildVersionPackage(ctx context.Context, input ExportVersionInput) (model.VersionPackage, error) {
	menuCodes := normalizedStringSet(input.MenuCodes)
	apiCodes := normalizedStringSet(input.APICodes)
	dictionaryCodes := normalizedStringSet(input.DictionaryCodes)
	if len(menuCodes) == 0 && len(apiCodes) == 0 && len(dictionaryCodes) == 0 {
		return model.VersionPackage{}, ErrInvalidInput
	}
	menus, err := s.ListMenus(ctx)
	if err != nil {
		return model.VersionPackage{}, err
	}
	apis, err := s.ListAPIs(ctx)
	if err != nil {
		return model.VersionPackage{}, err
	}
	dictionaryCatalog, err := s.ListDictionaries(ctx)
	if err != nil {
		return model.VersionPackage{}, err
	}
	if len(dictionaryCodes) > 0 && dictionaryCatalog.StorageStatus != "persisted" {
		return model.VersionPackage{}, ErrStorageUnavailable
	}
	pkg := model.VersionPackage{
		APIs:         selectedAPIs(apis, apiCodes),
		Dictionaries: selectedDictionaries(dictionaryCatalog.Items, dictionaryCodes),
		Menus:        selectedMenus(menus, menuCodes),
		Version: model.VersionPackageInfo{
			Code:        normalizeVersionCode(input.VersionCode),
			Description: strings.TrimSpace(input.Description),
			ExportTime:  s.now(),
			Name:        strings.TrimSpace(input.VersionName),
		},
	}
	if countMenus(pkg.Menus) == 0 && len(pkg.APIs) == 0 && len(pkg.Dictionaries) == 0 {
		return model.VersionPackage{}, ErrInvalidInput
	}
	return pkg, nil
}

func (s *service) importVersionDictionaries(ctx context.Context, dictionaries []model.Dictionary) (int, int, int, error) {
	createdDictionaries := 0
	skippedDictionaries := 0
	createdItems := 0
	for _, src := range dictionaries {
		code := normalizeDictionaryCode(src.Code)
		name := strings.TrimSpace(src.Name)
		if !validDictionaryCode(code) || name == "" {
			return createdDictionaries, skippedDictionaries, createdItems, ErrInvalidInput
		}
		status, err := normalizeDictionaryStatus(src.Status)
		if err != nil {
			status = model.DictionaryStatusActive
		}
		dictionary, err := s.repo.FindDictionaryByCode(ctx, code)
		if err == nil {
			skippedDictionaries++
		} else if errors.Is(err, database.ErrNotFound) {
			now := s.now()
			dictionary = &model.Dictionary{
				ID:          s.ids.NextID(),
				Code:        code,
				CreatedAt:   now,
				Description: strings.TrimSpace(src.Description),
				Name:        name,
				Status:      status,
				UpdatedAt:   now,
			}
			if err := s.repo.CreateDictionary(ctx, dictionary); err != nil {
				if isStorageDuplicate(err) {
					skippedDictionaries++
				} else {
					return createdDictionaries, skippedDictionaries, createdItems, err
				}
			} else {
				createdDictionaries++
			}
		} else {
			return createdDictionaries, skippedDictionaries, createdItems, err
		}
		if dictionary == nil {
			continue
		}
		items, err := s.importVersionDictionaryItems(ctx, dictionary.ID, src.Items)
		if err != nil {
			return createdDictionaries, skippedDictionaries, createdItems, err
		}
		createdItems += items
	}
	return createdDictionaries, skippedDictionaries, createdItems, nil
}

func (s *service) importVersionDictionaryItems(ctx context.Context, dictionaryID int64, srcItems []model.DictionaryItem) (int, error) {
	existing, err := s.repo.ListDictionaryItems(ctx, dictionaryID)
	if err != nil {
		return 0, err
	}
	byValue := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		byValue[item.Value] = struct{}{}
	}
	created := 0
	for _, src := range srcItems {
		value := strings.TrimSpace(src.Value)
		if value == "" {
			return created, ErrInvalidInput
		}
		if _, ok := byValue[value]; ok {
			continue
		}
		label := strings.TrimSpace(src.Label)
		if label == "" {
			label = value
		}
		status, err := normalizeDictionaryStatus(src.Status)
		if err != nil {
			status = model.DictionaryStatusActive
		}
		now := s.now()
		item := &model.DictionaryItem{
			ID:           s.ids.NextID(),
			CreatedAt:    now,
			DictionaryID: dictionaryID,
			Extra:        strings.TrimSpace(src.Extra),
			Label:        label,
			Sort:         src.Sort,
			Status:       status,
			UpdatedAt:    now,
			Value:        value,
		}
		if err := s.repo.CreateDictionaryItem(ctx, item); err != nil {
			if isStorageDuplicate(err) {
				continue
			}
			return created, err
		}
		byValue[value] = struct{}{}
		created++
	}
	return created, nil
}

func (s *service) findVersionRecord(ctx context.Context, id int64) (*model.Version, error) {
	if s.repo == nil {
		return nil, ErrStorageUnavailable
	}
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	version, err := s.repo.FindVersionByID(ctx, id)
	if err != nil {
		return nil, mapVersionLookupError(err)
	}
	return version, nil
}

func decodeVersionPackage(raw string) (model.VersionPackage, error) {
	var pkg model.VersionPackage
	if strings.TrimSpace(raw) == "" {
		return pkg, ErrInvalidInput
	}
	if err := json.Unmarshal([]byte(raw), &pkg); err != nil {
		return pkg, ErrInvalidInput
	}
	pkg.Version.Name = strings.TrimSpace(pkg.Version.Name)
	pkg.Version.Code = normalizeVersionCode(pkg.Version.Code)
	pkg.Version.Description = strings.TrimSpace(pkg.Version.Description)
	if pkg.Version.Name == "" || !validVersionCode(pkg.Version.Code) {
		return pkg, ErrInvalidInput
	}
	if pkg.Version.ExportTime.IsZero() {
		pkg.Version.ExportTime = time.Now().UTC()
	}
	if countMenus(pkg.Menus) == 0 && len(pkg.APIs) == 0 && len(pkg.Dictionaries) == 0 {
		return pkg, ErrInvalidInput
	}
	return pkg, nil
}

func selectedMenus(groups []model.MenuGroup, selected map[string]struct{}) []model.MenuGroup {
	if len(selected) == 0 {
		return nil
	}
	out := make([]model.MenuGroup, 0, len(groups))
	for _, group := range groups {
		items := make([]model.MenuItem, 0, len(group.Items))
		_, includeGroup := selected[normalizeSelector(group.Code)]
		for _, item := range group.Items {
			_, includeFull := selected[normalizeSelector(group.Code+":"+item.Code)]
			_, includeItem := selected[normalizeSelector(item.Code)]
			if includeGroup || includeFull || includeItem {
				items = append(items, item)
			}
		}
		if len(items) == 0 {
			continue
		}
		group.Items = items
		out = append(out, group)
	}
	return out
}

func selectedAPIs(groups []model.APIGroup, selected map[string]struct{}) []model.APIEntry {
	if len(selected) == 0 {
		return nil
	}
	out := make([]model.APIEntry, 0)
	for _, group := range groups {
		for _, item := range group.Items {
			if _, ok := selected[normalizeSelector(item.Code)]; ok {
				out = append(out, item)
				continue
			}
			if _, ok := selected[normalizeSelector(apiKey(item.Method, item.Path))]; ok {
				out = append(out, item)
				continue
			}
		}
	}
	return out
}

func selectedDictionaries(dictionaries []model.Dictionary, selected map[string]struct{}) []model.Dictionary {
	if len(selected) == 0 {
		return nil
	}
	out := make([]model.Dictionary, 0, len(dictionaries))
	for _, dictionary := range dictionaries {
		if _, ok := selected[normalizeSelector(dictionary.Code)]; ok {
			out = append(out, dictionary)
		}
	}
	return out
}

func countMenus(groups []model.MenuGroup) int {
	total := 0
	for _, group := range groups {
		total += len(group.Items)
	}
	return total
}

func countAPIs(groups []model.APIGroup) int {
	total := 0
	for _, group := range groups {
		total += len(group.Items)
	}
	return total
}

func normalizedStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalizeSelector(value)
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}
	return set
}

func normalizeSelector(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeVersionCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validVersionCode(code string) bool {
	if code == "" {
		return false
	}
	for _, char := range code {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '_' || char == '-' || char == '.' || char == ':':
		default:
			return false
		}
	}
	return true
}

func mapVersionLookupError(err error) error {
	switch {
	case errors.Is(err, database.ErrNotFound):
		return ErrNotFound
	case repository.IsStorageUnavailable(err):
		return ErrStorageUnavailable
	default:
		return err
	}
}
