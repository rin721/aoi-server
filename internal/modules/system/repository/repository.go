package repository

import (
	"context"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/system/model"
	"github.com/rei0721/go-scaffold/pkg/database"
)

type Repository interface {
	CreateAPI(context.Context, *model.APIRecord) error
	CreateDictionary(context.Context, *model.Dictionary) error
	CreateDictionaryItem(context.Context, *model.DictionaryItem) error
	CreateOperationRecord(context.Context, *model.OperationRecord) error
	CreateParameter(context.Context, *model.Parameter) error
	DeleteDictionary(context.Context, int64, time.Time) error
	DeleteDictionaryItem(context.Context, int64, time.Time) error
	DeleteOperationRecords(context.Context, []int64) error
	DeleteParameter(context.Context, int64, time.Time) error
	DeleteParameters(context.Context, []int64, time.Time) error
	FindAPI(context.Context, string, string) (*model.APIRecord, error)
	FindDictionaryByCode(context.Context, string) (*model.Dictionary, error)
	FindDictionaryByID(context.Context, int64) (*model.Dictionary, error)
	FindDictionaryItemByID(context.Context, int64) (*model.DictionaryItem, error)
	FindParameterByID(context.Context, int64) (*model.Parameter, error)
	FindParameterByKey(context.Context, string) (*model.Parameter, error)
	ListAPIs(context.Context) ([]model.APIRecord, error)
	ListDictionaries(context.Context) ([]model.Dictionary, error)
	ListDictionaryItems(context.Context, int64) ([]model.DictionaryItem, error)
	ListOperationRecords(context.Context, model.OperationRecordFilter) ([]model.OperationRecord, int64, error)
	ListParameters(context.Context, model.ParameterFilter) ([]model.Parameter, int64, error)
	SaveAPI(context.Context, *model.APIRecord) error
	SaveDictionary(context.Context, *model.Dictionary) error
	SaveDictionaryItem(context.Context, *model.DictionaryItem) error
	SaveParameter(context.Context, *model.Parameter) error
}

type repository struct {
	db database.Executor
}

func New(db database.Executor) Repository {
	return &repository{db: db}
}

func (r *repository) CreateAPI(ctx context.Context, api *model.APIRecord) error {
	return r.db.Create(ctx, api)
}

func (r *repository) CreateDictionary(ctx context.Context, dictionary *model.Dictionary) error {
	return r.db.Create(ctx, dictionary)
}

func (r *repository) CreateDictionaryItem(ctx context.Context, item *model.DictionaryItem) error {
	return r.db.Create(ctx, item)
}

func (r *repository) CreateOperationRecord(ctx context.Context, record *model.OperationRecord) error {
	return r.db.Create(ctx, record)
}

func (r *repository) CreateParameter(ctx context.Context, parameter *model.Parameter) error {
	return r.db.Create(ctx, parameter)
}

func (r *repository) DeleteDictionary(ctx context.Context, id int64, deletedAt time.Time) error {
	_, err := r.db.Update(ctx, &model.Dictionary{}, map[string]any{
		"deleted_at": deletedAt,
		"updated_at": deletedAt,
	}, database.Where("id = ?", id), alive())
	if err != nil {
		return err
	}
	_, err = r.db.Update(ctx, &model.DictionaryItem{}, map[string]any{
		"deleted_at": deletedAt,
		"updated_at": deletedAt,
	}, database.Where("dictionary_id = ?", id), alive())
	return err
}

func (r *repository) DeleteDictionaryItem(ctx context.Context, id int64, deletedAt time.Time) error {
	_, err := r.db.Update(ctx, &model.DictionaryItem{}, map[string]any{
		"deleted_at": deletedAt,
		"updated_at": deletedAt,
	}, database.Where("id = ?", id), alive())
	return err
}

func (r *repository) DeleteOperationRecords(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Delete(ctx, &model.OperationRecord{}, database.Where("id IN ?", ids))
	return err
}

func (r *repository) DeleteParameter(ctx context.Context, id int64, deletedAt time.Time) error {
	_, err := r.db.Update(ctx, &model.Parameter{}, map[string]any{
		"deleted_at": deletedAt,
		"updated_at": deletedAt,
	}, database.Where("id = ?", id), alive())
	return err
}

func (r *repository) DeleteParameters(ctx context.Context, ids []int64, deletedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Update(ctx, &model.Parameter{}, map[string]any{
		"deleted_at": deletedAt,
		"updated_at": deletedAt,
	}, database.Where("id IN ?", ids), alive())
	return err
}

func (r *repository) FindAPI(ctx context.Context, method string, path string) (*model.APIRecord, error) {
	var api model.APIRecord
	err := r.db.First(ctx, &api,
		database.Where("http_method = ?", method),
		database.Where("path = ?", path),
	)
	if err != nil {
		return nil, err
	}
	return &api, nil
}

func (r *repository) FindDictionaryByCode(ctx context.Context, code string) (*model.Dictionary, error) {
	var dictionary model.Dictionary
	if err := r.db.First(ctx, &dictionary, database.Where("code = ?", code), alive()); err != nil {
		return nil, err
	}
	return &dictionary, nil
}

func (r *repository) FindDictionaryByID(ctx context.Context, id int64) (*model.Dictionary, error) {
	var dictionary model.Dictionary
	if err := r.db.First(ctx, &dictionary, database.Where("id = ?", id), alive()); err != nil {
		return nil, err
	}
	return &dictionary, nil
}

func (r *repository) FindDictionaryItemByID(ctx context.Context, id int64) (*model.DictionaryItem, error) {
	var item model.DictionaryItem
	if err := r.db.First(ctx, &item, database.Where("id = ?", id), alive()); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) FindParameterByID(ctx context.Context, id int64) (*model.Parameter, error) {
	var parameter model.Parameter
	if err := r.db.First(ctx, &parameter, database.Where("id = ?", id), alive()); err != nil {
		return nil, err
	}
	return &parameter, nil
}

func (r *repository) FindParameterByKey(ctx context.Context, key string) (*model.Parameter, error) {
	var parameter model.Parameter
	if err := r.db.First(ctx, &parameter, database.Where("param_key = ?", key), alive()); err != nil {
		return nil, err
	}
	return &parameter, nil
}

func (r *repository) ListAPIs(ctx context.Context) ([]model.APIRecord, error) {
	var apis []model.APIRecord
	err := r.db.Find(ctx, &apis, database.Order("api_group ASC, path ASC, http_method ASC"))
	return apis, err
}

func (r *repository) ListDictionaries(ctx context.Context) ([]model.Dictionary, error) {
	var dictionaries []model.Dictionary
	err := r.db.Find(ctx, &dictionaries, alive(), database.Order("code ASC"))
	return dictionaries, err
}

func (r *repository) ListDictionaryItems(ctx context.Context, dictionaryID int64) ([]model.DictionaryItem, error) {
	var items []model.DictionaryItem
	err := r.db.Find(ctx, &items,
		database.Where("dictionary_id = ?", dictionaryID),
		alive(),
		database.Order("sort_order ASC, value ASC"),
	)
	return items, err
}

func (r *repository) ListOperationRecords(ctx context.Context, filter model.OperationRecordFilter) ([]model.OperationRecord, int64, error) {
	opts := operationRecordOptions(filter)
	var total int64
	var err error
	total, err = r.db.Count(ctx, &model.OperationRecord{}, opts...)
	if err != nil {
		return nil, 0, err
	}
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	opts = append(opts,
		database.Order("created_at DESC, id DESC"),
		database.Limit(pageSize),
		database.Offset((page-1)*pageSize),
	)
	var records []model.OperationRecord
	err = r.db.Find(ctx, &records, opts...)
	return records, total, err
}

func (r *repository) ListParameters(ctx context.Context, filter model.ParameterFilter) ([]model.Parameter, int64, error) {
	opts := parameterOptions(filter)
	var total int64
	var err error
	total, err = r.db.Count(ctx, &model.Parameter{}, opts...)
	if err != nil {
		return nil, 0, err
	}
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	opts = append(opts,
		database.Order("created_at DESC, id DESC"),
		database.Limit(pageSize),
		database.Offset((page-1)*pageSize),
	)
	var parameters []model.Parameter
	err = r.db.Find(ctx, &parameters, opts...)
	return parameters, total, err
}

func (r *repository) SaveAPI(ctx context.Context, api *model.APIRecord) error {
	api.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, api)
}

func (r *repository) SaveDictionary(ctx context.Context, dictionary *model.Dictionary) error {
	dictionary.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, dictionary)
}

func (r *repository) SaveDictionaryItem(ctx context.Context, item *model.DictionaryItem) error {
	item.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, item)
}

func (r *repository) SaveParameter(ctx context.Context, parameter *model.Parameter) error {
	parameter.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, parameter)
}

func alive() database.QueryOption {
	return database.Where("deleted_at IS NULL")
}

func operationRecordOptions(filter model.OperationRecordFilter) []database.QueryOption {
	opts := make([]database.QueryOption, 0, 3)
	method := strings.ToUpper(strings.TrimSpace(filter.Method))
	if method != "" {
		opts = append(opts, database.Where("http_method = ?", method))
	}
	path := strings.TrimSpace(filter.Path)
	if path != "" {
		opts = append(opts, database.Where("path LIKE ?", "%"+path+"%"))
	}
	if filter.Status > 0 {
		opts = append(opts, database.Where("status = ?", filter.Status))
		return opts
	}
	switch strings.ToLower(strings.TrimSpace(filter.StatusClass)) {
	case "4xx":
		opts = append(opts, database.Where("status >= ? AND status < ?", 400, 500))
	case "5xx":
		opts = append(opts, database.Where("status >= ? AND status < ?", 500, 600))
	case "error":
		opts = append(opts, database.Where("status >= ?", 400))
	}
	return opts
}

func parameterOptions(filter model.ParameterFilter) []database.QueryOption {
	opts := []database.QueryOption{alive()}
	name := strings.TrimSpace(filter.Name)
	if name != "" {
		opts = append(opts, database.Where("name LIKE ?", "%"+name+"%"))
	}
	key := strings.TrimSpace(filter.Key)
	if key != "" {
		opts = append(opts, database.Where("param_key LIKE ?", "%"+key+"%"))
	}
	if filter.StartCreatedAt != nil {
		opts = append(opts, database.Where("created_at >= ?", *filter.StartCreatedAt))
	}
	if filter.EndCreatedAt != nil {
		opts = append(opts, database.Where("created_at < ?", *filter.EndCreatedAt))
	}
	return opts
}

func IsStorageUnavailable(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such table") ||
		strings.Contains(text, "doesn't exist") ||
		strings.Contains(text, "undefined_table") ||
		strings.Contains(text, "unknown table")
}
