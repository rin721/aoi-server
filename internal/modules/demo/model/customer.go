package model

import "time"

// Customer 表示受权限保护的客户资源示例，用于演示后台资源归属和可见范围。
type Customer struct {
	ID                uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CustomerName      string     `gorm:"column:customer_name;size:120;not null" json:"customerName"`
	CustomerPhoneData string     `gorm:"column:customer_phone_data;size:64;not null" json:"customerPhoneData"`
	OwnerUserID       int64      `gorm:"column:owner_user_id;not null;index" json:"ownerUserId,string"`
	OwnerUsername     string     `gorm:"column:owner_username;size:120;not null" json:"ownerUsername"`
	OwnerRoleCode     string     `gorm:"column:owner_role_code;size:64;index" json:"ownerRoleCode"`
	OrgID             int64      `gorm:"column:org_id;not null;index" json:"orgId,string"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt         *time.Time `gorm:"column:deleted_at" json:"-"`
}

// TableName 固定客户示例表名，避免模型命名变化影响演示数据兼容性。
func (Customer) TableName() string {
	return "demo_customers"
}
