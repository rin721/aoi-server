package model

import "time"

const (
	MediaSourceUpload = "upload"
	MediaSourceURL    = "url"
)

type MediaCategory struct {
	ID        int64           `gorm:"column:id;primaryKey" json:"id,string"`
	ParentID  int64           `gorm:"column:parent_id;not null;index" json:"parentId,string"`
	Name      string          `gorm:"column:name;size:128;not null" json:"name"`
	Sort      int             `gorm:"column:sort_order;not null" json:"sort"`
	CreatedAt time.Time       `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time       `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt *time.Time      `gorm:"column:deleted_at" json:"-"`
	Children  []MediaCategory `gorm:"-" json:"children,omitempty"`
}

func (MediaCategory) TableName() string { return "system_media_categories" }

type MediaCategoryCatalog struct {
	Items         []MediaCategory `json:"items"`
	StorageStatus string          `json:"storageStatus"`
	Total         int             `json:"total"`
}

type MediaAsset struct {
	ID                 int64      `gorm:"column:id;primaryKey" json:"id,string"`
	CategoryID         int64      `gorm:"column:category_id;not null;index" json:"categoryId,string"`
	DisplayName        string     `gorm:"column:display_name;size:255;not null" json:"displayName"`
	OriginalName       string     `gorm:"column:original_name;size:255;not null" json:"originalName"`
	StorageKey         string     `gorm:"column:storage_key;size:512;not null" json:"storageKey"`
	URL                string     `gorm:"column:url;type:text;not null" json:"url"`
	MIMEType           string     `gorm:"column:mime_type;size:128;not null" json:"mimeType"`
	Extension          string     `gorm:"column:extension;size:32;not null" json:"extension"`
	SizeBytes          int64      `gorm:"column:size_bytes;not null" json:"sizeBytes"`
	Source             string     `gorm:"column:source;size:32;not null" json:"source"`
	External           bool       `gorm:"column:external;not null" json:"external"`
	UploadedBy         int64      `gorm:"column:uploaded_by;not null" json:"uploadedBy,string"`
	UploadedByUsername string     `gorm:"column:uploaded_by_username;size:128;not null" json:"uploadedByUsername"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;index" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
	DeletedAt          *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (MediaAsset) TableName() string { return "system_media_assets" }

type MediaAssetFilter struct {
	CategoryID int64
	Keyword    string
	Page       int
	PageSize   int
}

type MediaAssetPage struct {
	Items             []MediaAsset `json:"items"`
	ObjectStorage     string       `json:"objectStorage"`
	Page              int          `json:"page"`
	PageSize          int          `json:"pageSize"`
	StorageStatus     string       `json:"storageStatus"`
	Total             int64        `json:"total"`
	UploadMaxBytes    int64        `json:"uploadMaxBytes"`
	UploadMaxMB       int64        `json:"uploadMaxMb"`
	UploadUnavailable bool         `json:"uploadUnavailable"`
}

type MediaURLImportResult struct {
	Items         []MediaAsset `json:"items"`
	Imported      int          `json:"imported"`
	StorageStatus string       `json:"storageStatus"`
}
