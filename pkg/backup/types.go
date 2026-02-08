package backup

import "time"

type BackupMetadata struct {
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
	SizeBytes int64     `json:"size_bytes"`
}

type BackupFile struct {
	Version    int              `json:"version"`
	CreatedAt  time.Time        `json:"created_at"`
	AppVersion string           `json:"app_version"`
	Tables     map[string][]any `json:"tables"`
}

type RestoreResult struct {
	Success         bool   `json:"success"`
	AutoBackupFile  string `json:"auto_backup_file,omitempty"`
	TablesRestored  int    `json:"tables_restored"`
	RecordsRestored int64  `json:"records_restored"`
	Error           string `json:"error,omitempty"`
}
