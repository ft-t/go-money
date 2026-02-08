package backup

import (
	"context"
	"io"
)

type Service interface {
	Create(ctx context.Context) (*BackupMetadata, error)
	List(ctx context.Context) ([]BackupMetadata, error)
	GetFile(ctx context.Context, filename string) (io.ReadCloser, error)
	Delete(ctx context.Context, filename string) error
	Restore(ctx context.Context, filename string, dryRun bool) (*RestoreResult, error)
	RestoreFromReader(ctx context.Context, r io.Reader, dryRun bool) (*RestoreResult, error)
}
