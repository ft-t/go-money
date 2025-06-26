package tags

import (
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	"context"
	"fmt"
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

func (s *Service) ListTags(ctx context.Context, req *tagsv1.ListTagsRequest) (*tagsv1.ListTagsResponse, error) {
	var tags []*database.Tag

	query := database.GetDbWithContext(ctx, database.DbTypeReadonly)

	if req.IncludeDeleted {
		query = query.Unscoped()
	}

	if req.Name != nil {
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *req.Name))
	}

	if len(req.Ids) > 0 {
		query = query.Where("id IN ?", req.Ids)
	}

	if err := query.Find(&tags).Error; err != nil {
		return nil, err
	}

	var mapped []*tagsv1.ListTagsResponse_TagItem
	for _, tag := range tags {
		mapped = append(mapped, &tagsv1.ListTagsResponse_TagItem{
			Tag: s.mapper.MapTag(ctx, tag),
		})
	}

	return &tagsv1.ListTagsResponse{
		Tags: mapped,
	}, nil
}

func (s *Service) GetAllTags(ctx context.Context) ([]*database.Tag, error) {
	var tags []*database.Tag

	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (s *Service) CreateTag(
	ctx context.Context,
	req *tagsv1.CreateTagRequest,
) (*tagsv1.CreateTagResponse, error) {
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

func (s *Service) ImportTags(
	ctx context.Context,
	req *tagsv1.ImportTagsRequest,
) (*tagsv1.ImportTagsResponse, error) {
	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()

	finalResp := &tagsv1.ImportTagsResponse{
		Messages:     nil,
		CreatedCount: 0,
		UpdatedCount: 0,
	}

	for _, tag := range req.Tags {
		var existingTag database.Tag

		if err := tx.Where("name = ?", tag.Name).Find(&existingTag).Error; err != nil {
			return nil, errors.Wrap(err, "failed to check existing tag")
		}

		existingTag.Name = tag.Name
		existingTag.Color = tag.Color
		existingTag.Icon = tag.Icon

		if existingTag.ID == 0 {
			existingTag.CreatedAt = time.Now().UTC()
			finalResp.CreatedCount += 1

			finalResp.Messages = append(finalResp.Messages, fmt.Sprintf("created tag: %s", tag.Name))
		} else {
			finalResp.UpdatedCount += 1
			finalResp.Messages = append(finalResp.Messages, fmt.Sprintf("updated tag: %s", tag.Name))
		}

		if err := tx.Save(&existingTag).Error; err != nil {
			return nil, errors.Wrap(err, "failed to save tag")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return finalResp, nil
}
func (s *Service) DeleteTag(ctx context.Context, req *tagsv1.DeleteTagRequest) error {
	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	if err := db.Where("id = ?", req.Id).Delete(&database.Tag{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateTag(
	ctx context.Context,
	req *tagsv1.UpdateTagRequest,
) (*tagsv1.UpdateTagResponse, error) {
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
