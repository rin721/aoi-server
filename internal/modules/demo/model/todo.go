package model

// 本文件定义 Demo Todo 的持久化模型，作为 GORM 与 SQL 生成器共同识别的表结构契约。

import "time"

// Todo 表示 Demo 模块的持久化实体，同时作为 GORM 模型和 sqlgen 建表输入。
type Todo struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"column:title;size:200;not null" json:"title"`
	Description string     `gorm:"column:description;size:1000" json:"description,omitempty"`
	Completed   bool       `gorm:"column:completed;not null;default:false" json:"completed"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"-"`
}

// TableName 固定 Demo Todo 表名，避免结构体重命名影响数据库兼容性。
func (Todo) TableName() string {
	return "demo_todos"
}
