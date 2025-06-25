package tags_test

import (
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/tags"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestCreateAndGet(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapper(gomock.NewController(t))

	srv := tags.NewService(mapper)

	mapper.EXPECT().MapTag(gomock.Any(), gomock.Any()).
		Return(&gomoneypbv1.Tag{})

	resp, err := srv.CreateTag(context.TODO(), &tagsv1.CreateTagRequest{
		Name:  "some-name",
		Color: "a",
		Icon:  "i",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	allTags, err := srv.GetAllTags(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, allTags, 1)

	assert.Equal(t, "some-name", allTags[0].Name)
	assert.Equal(t, "a", allTags[0].Color)
	assert.Equal(t, "i", allTags[0].Icon)
	assert.NotZero(t, allTags[0].ID)
}

func TestCreateDuplicateTagFail(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapper(gomock.NewController(t))

	srv := tags.NewService(mapper)

	mapper.EXPECT().MapTag(gomock.Any(), gomock.Any()).
		Return(&gomoneypbv1.Tag{})

	_, err := srv.CreateTag(context.TODO(), &tagsv1.CreateTagRequest{
		Name:  "some-name",
		Color: "a",
		Icon:  "i",
	})
	assert.NoError(t, err)

	_, err = srv.CreateTag(context.TODO(), &tagsv1.CreateTagRequest{
		Name:  "some-name",
		Color: "b",
		Icon:  "j",
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "tag with this name already exists")
}

func TestUpdateTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))

		srv := tags.NewService(mapper)

		mapper.EXPECT().MapTag(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, tag *database.Tag) *gomoneypbv1.Tag {
				return &gomoneypbv1.Tag{
					Id: tag.ID,
				}
			}).Times(2)

		resp, err := srv.CreateTag(context.TODO(), &tagsv1.CreateTagRequest{
			Name:  "some-name",
			Color: "a",
			Icon:  "i",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		_, err = srv.UpdateTag(context.TODO(), &tagsv1.UpdateTagRequest{
			Id:    resp.Tag.Id,
			Name:  "some-name-2",
			Color: "b",
			Icon:  "j",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		allTags, err := srv.GetAllTags(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, allTags, 1)

		assert.Equal(t, "some-name-2", allTags[0].Name)
		assert.Equal(t, "b", allTags[0].Color)
		assert.Equal(t, "j", allTags[0].Icon)
	})

	t.Run("tag not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := tags.NewService(nil)

		_, err := srv.UpdateTag(context.TODO(), &tagsv1.UpdateTagRequest{
			Name:  "some-name",
			Color: "a",
			Icon:  "i",
		})
		assert.ErrorContains(t, err, "tag with not found")
	})
}

func TestDeleteTag(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapper(gomock.NewController(t))

	srv := tags.NewService(mapper)

	mapper.EXPECT().MapTag(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, tag *database.Tag) *gomoneypbv1.Tag {
			return &gomoneypbv1.Tag{
				Id: tag.ID,
			}
		})

	resp, err := srv.CreateTag(context.TODO(), &tagsv1.CreateTagRequest{
		Name:  "some-name",
		Color: "a",
		Icon:  "i",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	err = srv.DeleteTag(context.TODO(), &tagsv1.DeleteTagRequest{
		Id: resp.Tag.Id,
	})
	assert.NoError(t, err)

	allTags, err := srv.GetAllTags(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, allTags, 0)
}

func TestImportTags(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapper(gomock.NewController(t))

	srv := tags.NewService(mapper)

	assert.NoError(t, gormDB.Create(&database.Tag{
		Name:  "tag1",
		Color: "red",
		Icon:  "icon1",
	}).Error)

	resp, err := srv.ImportTags(context.TODO(), &tagsv1.ImportTagsRequest{
		Tags: []*tagsv1.CreateTagRequest{
			{
				Name:  "tag1",
				Color: "xx",
				Icon:  "yy",
			},
			{
				Name:  "tag2",
				Color: "white",
				Icon:  "icon2",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	allTags, err := srv.GetAllTags(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, allTags, 2)

	assert.Equal(t, "tag1", allTags[0].Name)
	assert.Equal(t, "xx", allTags[0].Color)
	assert.Equal(t, "yy", allTags[0].Icon)

	assert.Equal(t, "tag2", allTags[1].Name)
	assert.Equal(t, "white", allTags[1].Color)
	assert.Equal(t, "icon2", allTags[1].Icon)
}
