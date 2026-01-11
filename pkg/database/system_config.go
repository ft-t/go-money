package database

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	SystemConfigKeyJwtPrivateKey = "jwt_private_key"
)

type SystemConfig struct {
	ID        string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SystemConfig) TableName() string {
	return "system_configurations"
}

type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

func (r *SystemConfigRepository) Get(ctx context.Context, key string) (string, error) {
	var cfg SystemConfig

	err := r.db.WithContext(ctx).Where("id = ?", key).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	return cfg.Value, nil
}

func (r *SystemConfigRepository) Set(ctx context.Context, key, value string) error {
	now := time.Now().UTC()

	cfg := SystemConfig{
		ID:        key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return r.db.WithContext(ctx).
		Where("id = ?", key).
		Assign(map[string]any{
			"value":      value,
			"updated_at": now,
		}).
		FirstOrCreate(&cfg).Error
}
