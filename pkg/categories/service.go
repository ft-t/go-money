package categories

import (
	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm"
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

func (s *Service) ListCategories(
	ctx context.Context,
	req *categoriesv1.ListCategoriesRequest,
) (*categoriesv1.ListCategoriesResponse, error) {
	var categories []*database.Category

	query := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	if req.IncludeDeleted {
		query = query.Unscoped()
	}

	if len(req.Ids) > 0 {
		query = query.Where("id IN ?", req.Ids)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	var mapped []*gomoneypbv1.Category
	for _, category := range categories {
		mapped = append(mapped, s.mapper.MapCategory(ctx, category))
	}

	return &categoriesv1.ListCategoriesResponse{
		Categories: mapped,
	}, nil
}

func (s *Service) GetAllCategories(ctx context.Context) ([]*database.Category, error) {
	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly))

	var categories []*database.Category

	if err := db.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Service) DeleteCategory(
	ctx context.Context,
	req *categoriesv1.DeleteCategoryRequest,
) (*categoriesv1.DeleteCategoryResponse, error) {
	var category database.Category

	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	if err := db.Where("id = ?", req.Id).First(&category).Error; err != nil {
		return nil, err
	}

	if err := db.Delete(&category).Error; err != nil {
		return nil, err
	}

	return &categoriesv1.DeleteCategoryResponse{
		Category: s.mapper.MapCategory(ctx, &category),
	}, nil
}

func (s *Service) CreateCategory(
	ctx context.Context,
	req *categoriesv1.CreateCategoryRequest,
) (*categoriesv1.CreateCategoryResponse, error) {
	category := &database.Category{
		Name: req.Category.Name,
	}

	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	if err := s.ensureNotExists(db, category.Name); err != nil {
		return nil, err
	}

	if err := db.Create(category).Error; err != nil {
		return nil, err
	}

	return &categoriesv1.CreateCategoryResponse{
		Category: s.mapper.MapCategory(ctx, category),
	}, nil
}

func (s *Service) ensureNotExists(db *gorm.DB, name string) error {
	var existingCategory database.Category
	if err := db.Debug().Where("name = ?", name).Find(&existingCategory).Error; err != nil {
		return err
	}

	if existingCategory.ID != 0 {
		return gorm.ErrDuplicatedKey
	}

	return nil
}

func (s *Service) UpdateCategory(
	ctx context.Context,
	req *categoriesv1.UpdateCategoryRequest,
) (*categoriesv1.UpdateCategoryResponse, error) {
	var category database.Category

	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	if err := db.Where("id = ?", req.Category.Id).
		First(&category).Error; err != nil {
		return nil, err
	}

	category.Name = req.Category.Name

	if err := s.ensureNotExists(db, category.Name); err != nil {
		return nil, err
	}

	if err := db.Save(&category).Error; err != nil {
		return nil, err
	}

	return &categoriesv1.UpdateCategoryResponse{
		Category: s.mapper.MapCategory(ctx, &category),
	}, nil
}
