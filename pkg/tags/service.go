package tags

import (
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"time"
)

type Service struct {
	mapper Mapper
}

func NewService(
	mapper Mapper,
) *Service {
	return &Service{
		mapper: mapper,
	}
}

func (s *Service) GetAllTags(ctx context.Context) ([]*database.Tag, error) {
	var tags []*database.Tag

	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (s *Service) CreateTag(ctx context.Context, req *tagsv1.CreateTagRequest) (*tagsv1.CreateTagResponse, error) {
	var existingTag database.Tag

	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	if err := db.Where("name = ?", req.Name).Find(&existingTag).Error; err != nil {
		return nil, err
	}

	if existingTag.ID != 0 {
		return nil, errors.New("tag with this name already exists")
	}

	existingTag.Name = req.Name
	existingTag.Color = req.Color
	existingTag.Icon = req.Icon
	existingTag.CreatedAt = time.Now().UTC()

	if err := db.Create(&existingTag).Error; err != nil {
		return nil, err
	}

	return &tagsv1.CreateTagResponse{
		Tag: s.mapper.MapTag(ctx, &existingTag),
	}, nil
}

func (s *Service) Import(ctx context.Context) {
	// todo
}

func (s *Service) DeleteTag(ctx context.Context, req *tagsv1.DeleteTagRequest) error {
	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	if err := db.Where("id = ?", req.Id).Delete(&database.Tag{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateTag(ctx context.Context, req *tagsv1.UpdateTagRequest) (*tagsv1.UpdateTagResponse, error) {
	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	var existingTag database.Tag
	if err := db.Where("id = ?", req.Id).First(&existingTag).Error; err != nil {
		return nil, errors.Wrap(err, "tag with not found")
	}

	existingTag.Name = req.Name
	existingTag.Color = req.Color
	existingTag.Icon = req.Icon

	if err := db.Save(&existingTag).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update tag")
	}

	return &tagsv1.UpdateTagResponse{
		Tag: s.mapper.MapTag(ctx, &existingTag),
	}, nil
}
