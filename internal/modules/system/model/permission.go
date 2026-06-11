package model

import "time"

type PermissionEntry struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionSyncItem struct {
	Code        string `json:"code"`
	Created     bool   `json:"created"`
	Description string `json:"description"`
	Exists      bool   `json:"exists"`
	Name        string `json:"name"`
}

type PermissionSyncResult struct {
	Created       int                  `json:"created"`
	Items         []PermissionSyncItem `json:"items"`
	Persisted     bool                 `json:"persisted"`
	Skipped       int                  `json:"skipped"`
	StorageStatus string               `json:"storageStatus"`
	SyncedAt      time.Time            `json:"syncedAt"`
	Total         int                  `json:"total"`
}
