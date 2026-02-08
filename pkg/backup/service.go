package backup

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

const (
	BackupVersion = 1
	AppVersion    = "1.0.0"
)

var excludedTables = map[string]bool{
	"gormigrate":        true,
	"schema_migrations": true,
	"jti_revocations":   true,
}

type service struct {
	db        *gorm.DB
	backupDir string
	dbConfig  boilerplate.DbConfig
}

func NewService(db *gorm.DB, backupDir string, dbConfig boilerplate.DbConfig) *service {
	return &service{
		db:        db,
		backupDir: backupDir,
		dbConfig:  dbConfig,
	}
}

func (s *service) GetTables(ctx context.Context) ([]string, error) {
	var tables []string

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		if !excludedTables[tableName] {
			tables = append(tables, tableName)
		}
	}

	return tables, nil
}

func (s *service) Create(ctx context.Context) (*BackupMetadata, error) {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	tables, err := s.GetTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	backupData := BackupFile{
		Version:    BackupVersion,
		CreatedAt:  time.Now().UTC(),
		AppVersion: AppVersion,
		Tables:     make(map[string][]any),
	}

	for _, table := range tables {
		var rows []map[string]any
		if err := s.db.WithContext(ctx).Table(table).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("failed to export table %s: %w", table, err)
		}
		backupData.Tables[table] = make([]any, len(rows))
		for i, row := range rows {
			backupData.Tables[table][i] = row
		}
	}

	filename := fmt.Sprintf("gomoney-backup-%s.json.gz", time.Now().Format("2006-01-02-150405"))
	fullPath := filepath.Join(s.backupDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)

	encoder := json.NewEncoder(gzWriter)
	if err := encoder.Encode(backupData); err != nil {
		_ = gzWriter.Close()
		return nil, fmt.Errorf("failed to encode backup: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	return &BackupMetadata{
		Filename:  filename,
		CreatedAt: backupData.CreatedAt,
		SizeBytes: stat.Size(),
	}, nil
}

func (s *service) List(ctx context.Context) ([]BackupMetadata, error) {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "gomoney-backup-") || !strings.HasSuffix(entry.Name(), ".json.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupMetadata{
			Filename:  entry.Name(),
			CreatedAt: info.ModTime(),
			SizeBytes: info.Size(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

func (s *service) GetFile(ctx context.Context, filename string) (io.ReadCloser, error) {
	if !isValidBackupFilename(filename) {
		return nil, fmt.Errorf("invalid backup filename")
	}

	path := filepath.Join(s.backupDir, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}

	return file, nil
}

func isValidBackupFilename(filename string) bool {
	return strings.HasPrefix(filename, "gomoney-backup-") &&
		strings.HasSuffix(filename, ".json.gz") &&
		!strings.Contains(filename, "/") &&
		!strings.Contains(filename, "..")
}

func (s *service) Delete(ctx context.Context, filename string) error {
	if !isValidBackupFilename(filename) {
		return fmt.Errorf("invalid backup filename")
	}

	path := filepath.Join(s.backupDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	return nil
}

func (s *service) Restore(ctx context.Context, filename string, dryRun bool) (*RestoreResult, error) {
	file, err := s.GetFile(ctx, filename)
	if err != nil {
		return &RestoreResult{Success: false, Error: err.Error()}, err
	}
	defer func() { _ = file.Close() }()

	return s.RestoreFromReader(ctx, file, dryRun)
}

func (s *service) RestoreFromReader(ctx context.Context, r io.Reader, dryRun bool) (*RestoreResult, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return &RestoreResult{Success: false, Error: "invalid gzip format"}, err
	}
	defer func() { _ = gzReader.Close() }()

	var backupData BackupFile
	if err := json.NewDecoder(gzReader).Decode(&backupData); err != nil {
		return &RestoreResult{Success: false, Error: "invalid JSON format"}, err
	}

	if backupData.Version > BackupVersion {
		return &RestoreResult{Success: false, Error: "unsupported backup version"}, fmt.Errorf("backup version %d not supported", backupData.Version)
	}

	if dryRun {
		return s.restoreDryRun(ctx, &backupData)
	}

	return s.restoreToDatabase(ctx, &backupData)
}

func (s *service) restoreToDatabase(ctx context.Context, backupData *BackupFile) (*RestoreResult, error) {
	autoBackup, err := s.Create(ctx)
	if err != nil {
		return &RestoreResult{Success: false, Error: "failed to create auto-backup"}, fmt.Errorf("auto-backup failed: %w", err)
	}

	tables, err := s.GetTables(ctx)
	if err != nil {
		return &RestoreResult{Success: false, Error: "failed to get tables"}, err
	}

	var tablesRestored int
	var recordsRestored int64

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET session_replication_role = 'replica'").Error; err != nil {
			return fmt.Errorf("failed to disable constraints: %w", err)
		}

		for _, table := range tables {
			if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %q CASCADE", table)).Error; err != nil {
				return fmt.Errorf("failed to truncate table %s: %w", table, err)
			}
		}

		for tableName, rows := range backupData.Tables {
			for _, row := range rows {
				rowMap, ok := row.(map[string]any)
				if !ok {
					continue
				}
				if err := tx.Table(tableName).Create(rowMap).Error; err != nil {
					return fmt.Errorf("failed to insert into %s: %w", tableName, err)
				}
				recordsRestored++
			}
			tablesRestored++
		}

		if err := tx.Exec("SET session_replication_role = 'origin'").Error; err != nil {
			return fmt.Errorf("failed to re-enable constraints: %w", err)
		}

		return nil
	})

	if err != nil {
		return &RestoreResult{
			Success:        false,
			AutoBackupFile: autoBackup.Filename,
			Error:          err.Error(),
		}, err
	}

	s.resetSequences(ctx, backupData)

	return &RestoreResult{
		Success:         true,
		AutoBackupFile:  autoBackup.Filename,
		TablesRestored:  tablesRestored,
		RecordsRestored: recordsRestored,
	}, nil
}

func (s *service) resetSequences(ctx context.Context, backupData *BackupFile) {
	for tableName, rows := range backupData.Tables {
		if len(rows) == 0 {
			continue
		}

		var maxID int64
		for _, row := range rows {
			rowMap, ok := row.(map[string]any)
			if !ok {
				continue
			}
			if id, ok := rowMap["id"]; ok {
				var idVal int64
				switch v := id.(type) {
				case float64:
					idVal = int64(v)
				case int64:
					idVal = v
				case int:
					idVal = int64(v)
				}
				if idVal > maxID {
					maxID = idVal
				}
			}
		}

		if maxID > 0 {
			seqName := fmt.Sprintf("%s_id_seq", tableName)
			s.db.WithContext(ctx).Exec(fmt.Sprintf("SELECT setval('%s', %d, true)", seqName, maxID))
		}
	}
}

func (s *service) restoreDryRun(ctx context.Context, backupData *BackupFile) (*RestoreResult, error) {
	tempDBName := fmt.Sprintf("gomoney_restore_test_%d", time.Now().UnixNano())

	adminConfig := s.dbConfig
	adminConfig.Db = "postgres"
	adminConnStr, _ := boilerplate.GetDbConnectionString(adminConfig)

	adminConn, err := pgx.Connect(ctx, adminConnStr)
	if err != nil {
		return &RestoreResult{Success: false, Error: "failed to connect to postgres"}, err
	}
	defer func() { _ = adminConn.Close(ctx) }()

	_, err = adminConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", tempDBName))
	if err != nil {
		return &RestoreResult{Success: false, Error: "failed to create temp database"}, err
	}

	defer func() {
		_, _ = adminConn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", tempDBName))
	}()

	tempConfig := s.dbConfig
	tempConfig.Db = tempDBName
	tempDB, err := boilerplate.GetGormConnection(tempConfig)
	if err != nil {
		return &RestoreResult{Success: false, Error: "failed to connect to temp database"}, err
	}

	sqlDB, _ := tempDB.DB()
	defer func() { _ = sqlDB.Close() }()

	if err := database.RunMigrations(tempDB); err != nil {
		return &RestoreResult{Success: false, Error: fmt.Sprintf("migrations failed: %v", err)}, err
	}

	tempService := &service{
		db:        tempDB,
		backupDir: s.backupDir,
		dbConfig:  tempConfig,
	}

	result, err := tempService.restoreToTempDatabase(ctx, backupData)
	if err != nil {
		return &RestoreResult{Success: false, Error: err.Error()}, err
	}

	result.AutoBackupFile = ""
	return result, nil
}

func (s *service) restoreToTempDatabase(ctx context.Context, backupData *BackupFile) (*RestoreResult, error) {
	tables, err := s.GetTables(ctx)
	if err != nil {
		return &RestoreResult{Success: false, Error: "failed to get tables"}, err
	}

	var tablesRestored int
	var recordsRestored int64

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET session_replication_role = 'replica'").Error; err != nil {
			return fmt.Errorf("failed to disable constraints: %w", err)
		}

		for _, table := range tables {
			if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %q CASCADE", table)).Error; err != nil {
				return fmt.Errorf("failed to truncate table %s: %w", table, err)
			}
		}

		for tableName, rows := range backupData.Tables {
			for _, row := range rows {
				rowMap, ok := row.(map[string]any)
				if !ok {
					continue
				}
				if err := tx.Table(tableName).Create(rowMap).Error; err != nil {
					return fmt.Errorf("failed to insert into %s: %w", tableName, err)
				}
				recordsRestored++
			}
			tablesRestored++
		}

		if err := tx.Exec("SET session_replication_role = 'origin'").Error; err != nil {
			return fmt.Errorf("failed to re-enable constraints: %w", err)
		}

		return nil
	})

	if err != nil {
		return &RestoreResult{Success: false, Error: err.Error()}, err
	}

	return &RestoreResult{
		Success:         true,
		TablesRestored:  tablesRestored,
		RecordsRestored: recordsRestored,
	}, nil
}
