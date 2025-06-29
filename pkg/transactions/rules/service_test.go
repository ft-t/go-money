package rules_test

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestCreateAndDelete(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	mapper.EXPECT().MapRule(gomock.Any()).
		DoAndReturn(func(rule *database.Rule) *gomoneypbv1.Rule {
			return &gomoneypbv1.Rule{
				Id: rule.ID,
			}
		}).Times(2)

	svc := rules.NewService(mapper)

	resp, err := svc.CreateRule(context.TODO(), &rulesv1.CreateRuleRequest{
		Rule: &gomoneypbv1.Rule{
			Title:       "some-title",
			Script:      "some-script",
			Interpreter: gomoneypbv1.RuleInterpreterType_RULE_INTERPRETER_TYPE_LUA,
			SortOrder:   55,
			Enabled:     true,
			IsFinalRule: true,
			GroupName:   "some-group",
		},
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Rule.Id)

	var updated database.Rule
	assert.NoError(t, gormDB.Find(&updated, resp.Rule.Id).Error)

	assert.Equal(t, "some-title", updated.Title)
	assert.Equal(t, "some-script", updated.Script)
	assert.False(t, updated.DeletedAt.Valid)

	deleteResp, err := svc.DeleteRule(context.TODO(), &rulesv1.DeleteRuleRequest{
		Id: resp.Rule.Id,
	})
	assert.NoError(t, err)

	assert.Equal(t, resp.Rule.Id, deleteResp.Rule.Id)

	var deleted database.Rule
	assert.NoError(t, gormDB.Unscoped().First(&deleted, deleteResp.Rule.Id).Error)

	assert.True(t, deleted.DeletedAt.Valid)
}

func TestUpdateRule(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	mapper.EXPECT().MapRule(gomock.Any()).
		DoAndReturn(func(rule *database.Rule) *gomoneypbv1.Rule {
			return &gomoneypbv1.Rule{
				Id: rule.ID,
			}
		})

	existing := &database.Rule{
		Script: "existing-script",
		Title:  "existing-title",
	}
	assert.NoError(t, gormDB.Create(existing).Error)

	svc := rules.NewService(mapper)
	resp, err := svc.UpdateRule(context.TODO(), &rulesv1.UpdateRuleRequest{
		Rule: &gomoneypbv1.Rule{
			Id:     existing.ID,
			Title:  "updated-title",
			Script: "updated-script",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, existing.ID, resp.Rule.Id)

	var updated database.Rule
	assert.NoError(t, gormDB.Find(&updated, existing.ID).Error)
	assert.Equal(t, "updated-title", updated.Title)
	assert.Equal(t, "updated-script", updated.Script)
}

func TestListRule(t *testing.T) {
	t.Run("no filters", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		mapper.EXPECT().MapRule(gomock.Any()).
			DoAndReturn(func(rule *database.Rule) *gomoneypbv1.Rule {
				return &gomoneypbv1.Rule{
					Id: rule.ID,
				}
			}).Times(3)

		svc := rules.NewService(mapper)

		dbRules := []*database.Rule{
			{ID: 1, Title: "Rule 1"},
			{ID: 2, Title: "Rule 2"},
			{ID: 3, Title: "Rule 3"},
		}

		assert.NoError(t, gormDB.Create(dbRules).Error)

		resp, err := svc.ListRules(context.TODO(), &rulesv1.ListRulesRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Rules, 3)
	})

	t.Run("with ids filter and deleted", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		mapper.EXPECT().MapRule(gomock.Any()).
			DoAndReturn(func(rule *database.Rule) *gomoneypbv1.Rule {
				return &gomoneypbv1.Rule{
					Id: rule.ID,
				}
			}).Times(1)

		svc := rules.NewService(mapper)

		dbRules := []*database.Rule{
			{ID: 1, Title: "Rule 1"},
			{ID: 2, Title: "Rule 2", DeletedAt: gorm.DeletedAt{
				Valid: true,
				Time:  time.Now(),
			}},
		}

		assert.NoError(t, gormDB.Create(dbRules).Error)

		resp, err := svc.ListRules(context.TODO(), &rulesv1.ListRulesRequest{
			Ids:            []int32{2},
			IncludeDeleted: true,
		})
		assert.NoError(t, err)
		assert.Len(t, resp.Rules, 1)

		assert.EqualValues(t, 2, resp.Rules[0].Id)
	})
}
