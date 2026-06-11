package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type gormExecutor struct {
	db *gorm.DB
}

func (d *database) executor(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	if tx := txFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return d.gormDB().WithContext(ctx)
}

func (d *database) Create(ctx context.Context, value any) error {
	return mapError(d.executor(ctx).Create(value).Error)
}

func (d *database) Save(ctx context.Context, value any) error {
	return mapError(d.executor(ctx).Save(value).Error)
}

func (d *database) First(ctx context.Context, dest any, opts ...QueryOption) error {
	return mapError(applyQuery(d.executor(ctx), opts...).First(dest).Error)
}

func (d *database) Find(ctx context.Context, dest any, opts ...QueryOption) error {
	return mapError(applyQuery(d.executor(ctx), opts...).Find(dest).Error)
}

func (d *database) Update(ctx context.Context, model any, values map[string]any, opts ...QueryOption) (Result, error) {
	result := applyQuery(d.executor(ctx).Model(model), opts...).Updates(values)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (d *database) Delete(ctx context.Context, model any, opts ...QueryOption) (Result, error) {
	result := applyQuery(d.executor(ctx), opts...).Delete(model)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (d *database) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	result := d.executor(ctx).Exec(sql, args...)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (d *database) Raw(ctx context.Context, dest any, sql string, args ...any) (Result, error) {
	result := d.executor(ctx).Raw(sql, args...).Scan(dest)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (d *database) Count(ctx context.Context, model any, opts ...QueryOption) (int64, error) {
	var count int64
	result := applyQuery(d.executor(ctx).Model(model), opts...).Count(&count)
	return count, mapError(result.Error)
}

func (d *database) HasTable(ctx context.Context, model any) (bool, error) {
	return d.executor(ctx).Migrator().HasTable(model), nil
}

func (e *gormExecutor) Create(ctx context.Context, value any) error {
	return mapError(e.withContext(ctx).Create(value).Error)
}

func (e *gormExecutor) Save(ctx context.Context, value any) error {
	return mapError(e.withContext(ctx).Save(value).Error)
}

func (e *gormExecutor) First(ctx context.Context, dest any, opts ...QueryOption) error {
	return mapError(applyQuery(e.withContext(ctx), opts...).First(dest).Error)
}

func (e *gormExecutor) Find(ctx context.Context, dest any, opts ...QueryOption) error {
	return mapError(applyQuery(e.withContext(ctx), opts...).Find(dest).Error)
}

func (e *gormExecutor) Update(ctx context.Context, model any, values map[string]any, opts ...QueryOption) (Result, error) {
	result := applyQuery(e.withContext(ctx).Model(model), opts...).Updates(values)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (e *gormExecutor) Delete(ctx context.Context, model any, opts ...QueryOption) (Result, error) {
	result := applyQuery(e.withContext(ctx), opts...).Delete(model)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (e *gormExecutor) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	result := e.withContext(ctx).Exec(sql, args...)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (e *gormExecutor) Raw(ctx context.Context, dest any, sql string, args ...any) (Result, error) {
	result := e.withContext(ctx).Raw(sql, args...).Scan(dest)
	return Result{RowsAffected: result.RowsAffected}, mapError(result.Error)
}

func (e *gormExecutor) Count(ctx context.Context, model any, opts ...QueryOption) (int64, error) {
	var count int64
	result := applyQuery(e.withContext(ctx).Model(model), opts...).Count(&count)
	return count, mapError(result.Error)
}

func (e *gormExecutor) HasTable(ctx context.Context, model any) (bool, error) {
	return e.withContext(ctx).Migrator().HasTable(model), nil
}

func (e *gormExecutor) withContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	return e.db.WithContext(ctx)
}

func applyQuery(db *gorm.DB, opts ...QueryOption) *gorm.DB {
	q := Query{}
	for _, opt := range opts {
		if opt != nil {
			opt(&q)
		}
	}
	if q.Table != "" {
		db = db.Table(q.Table)
	}
	for _, condition := range q.Where {
		db = db.Where(condition.Expr, condition.Args...)
	}
	if q.Order != "" {
		db = db.Order(q.Order)
	}
	if q.Limit > 0 {
		db = db.Limit(q.Limit)
	}
	if q.Offset > 0 {
		db = db.Offset(q.Offset)
	}
	return db
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
