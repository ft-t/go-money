package backup_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/backup"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestService_GetTables(t *testing.T) {
	t.Run("returns tables from database", func(t *testing.T) {
		svc := backup.NewService(gormDB, t.TempDir(), cfg.Db)

		tables, err := svc.GetTables(context.Background())

		require.NoError(t, err)
		assert.Contains(t, tables, "accounts")
		assert.Contains(t, tables, "transactions")
		assert.Contains(t, tables, "categories")
		assert.NotContains(t, tables, "gormigrate")
	})
}

func TestService_Create(t *testing.T) {
	t.Run("creates backup file", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		require.NoError(t, gormDB.Exec("INSERT INTO categories (id, name, created_at) VALUES (1, 'Test Category', NOW())").Error)

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		meta, err := svc.Create(context.Background())

		require.NoError(t, err)
		assert.NotEmpty(t, meta.Filename)
		assert.True(t, strings.HasPrefix(meta.Filename, "gomoney-backup-"))
		assert.True(t, strings.HasSuffix(meta.Filename, ".json.gz"))
		assert.Greater(t, meta.SizeBytes, int64(0))

		_, err = os.Stat(filepath.Join(backupDir, meta.Filename))
		assert.NoError(t, err)
	})
}

func TestService_List(t *testing.T) {
	t.Run("returns list of backups", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		meta1, err := svc.Create(context.Background())
		require.NoError(t, err)

		time.Sleep(time.Second)

		meta2, err := svc.Create(context.Background())
		require.NoError(t, err)

		list, err := svc.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, meta2.Filename, list[0].Filename)
		assert.Equal(t, meta1.Filename, list[1].Filename)
	})

	t.Run("returns empty list when no backups", func(t *testing.T) {
		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		list, err := svc.List(context.Background())

		require.NoError(t, err)
		assert.Empty(t, list)
	})
}

func TestService_GetFile(t *testing.T) {
	t.Run("returns file reader", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		meta, err := svc.Create(context.Background())
		require.NoError(t, err)

		reader, err := svc.GetFile(context.Background(), meta.Filename)

		require.NoError(t, err)
		defer func() { _ = reader.Close() }()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Greater(t, len(data), 0)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		_, err := svc.GetFile(context.Background(), "gomoney-backup-2024-01-01-000000.json.gz")

		assert.Error(t, err)
	})

	t.Run("rejects invalid filename", func(t *testing.T) {
		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		_, err := svc.GetFile(context.Background(), "../../../etc/passwd")

		assert.Error(t, err)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("deletes backup file", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		meta, err := svc.Create(context.Background())
		require.NoError(t, err)

		err = svc.Delete(context.Background(), meta.Filename)

		require.NoError(t, err)

		list, err := svc.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		err := svc.Delete(context.Background(), "gomoney-backup-2024-01-01-000000.json.gz")

		assert.Error(t, err)
	})

	t.Run("rejects invalid filename", func(t *testing.T) {
		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		err := svc.Delete(context.Background(), "../../../etc/passwd")

		assert.Error(t, err)
	})
}

func TestService_RestoreFromReader(t *testing.T) {
	t.Run("restores data from backup", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		require.NoError(t, gormDB.Exec("INSERT INTO categories (id, name, created_at) VALUES (1, 'Original Category', NOW())").Error)

		meta, err := svc.Create(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.Exec("UPDATE categories SET name = 'Modified Category' WHERE id = 1").Error)
		require.NoError(t, gormDB.Exec("INSERT INTO categories (id, name, created_at) VALUES (2, 'New Category', NOW())").Error)

		var count int64
		gormDB.Table("categories").Count(&count)
		assert.Equal(t, int64(2), count)

		file, err := svc.GetFile(context.Background(), meta.Filename)
		require.NoError(t, err)
		defer func() { _ = file.Close() }()

		result, err := svc.RestoreFromReader(context.Background(), file, false)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.AutoBackupFile)
		assert.Greater(t, result.TablesRestored, 0)

		var categories []map[string]any
		gormDB.Table("categories").Find(&categories)
		assert.Len(t, categories, 1)
		assert.Equal(t, "Original Category", categories[0]["name"])
	})
}

func TestService_Restore(t *testing.T) {
	t.Run("restores from filename", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		require.NoError(t, gormDB.Exec("INSERT INTO tags (id, name, created_at) VALUES (1, 'Original Tag', NOW())").Error)
		meta, err := svc.Create(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.Exec("DELETE FROM tags WHERE id = 1").Error)

		result, err := svc.Restore(context.Background(), meta.Filename, false)

		require.NoError(t, err)
		assert.True(t, result.Success)

		var tags []map[string]any
		gormDB.Table("tags").Find(&tags)
		assert.Len(t, tags, 1)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		result, err := svc.Restore(context.Background(), "gomoney-backup-2024-01-01-000000.json.gz", false)

		assert.Error(t, err)
		assert.False(t, result.Success)
	})
}

func TestService_RestoreDryRun(t *testing.T) {
	t.Run("validates backup without modifying data", func(t *testing.T) {
		require.NoError(t, testingutils.FlushAllTables(cfg.Db))

		backupDir := t.TempDir()
		svc := backup.NewService(gormDB, backupDir, cfg.Db)

		require.NoError(t, gormDB.Exec("INSERT INTO categories (id, name, created_at) VALUES (1, 'Test Category', NOW())").Error)
		meta, err := svc.Create(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.Exec("UPDATE categories SET name = 'Modified' WHERE id = 1").Error)

		result, err := svc.Restore(context.Background(), meta.Filename, true)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Empty(t, result.AutoBackupFile)

		var categories []map[string]any
		gormDB.Table("categories").Find(&categories)
		assert.Len(t, categories, 1)
		assert.Equal(t, "Modified", categories[0]["name"])
	})
}
