package rules_test

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScheduleService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		mapper.EXPECT().MapScheduleRule(gomock.Any()).
			DoAndReturn(func(rule *database.ScheduleRule) *gomoneypbv1.ScheduleRule {
				return &gomoneypbv1.ScheduleRule{Id: rule.ID}
			})

		scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(nil)
		scheduler.EXPECT().Reinit().Return(nil)

		svc := rules.NewScheduleService(mapper, scheduler)

		resp, err := svc.CreateRule(context.TODO(), &rulesv1.CreateScheduleRuleRequest{
			Rule: &gomoneypbv1.ScheduleRule{
				Title:          "title",
				Script:         "script",
				Interpreter:    gomoneypbv1.RuleInterpreterType_RULE_INTERPRETER_TYPE_LUA,
				Enabled:        true,
				GroupName:      "group",
				CronExpression: "* * * * *",
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Rule.Id)

		var created database.ScheduleRule
		assert.NoError(t, gormDB.Find(&created, resp.Rule.Id).Error)
		assert.Equal(t, "title", created.Title)
		assert.Equal(t, "script", created.Script)
	})

	t.Run("invalid cron expression", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(assert.AnError)

		svc := rules.NewScheduleService(mapper, scheduler)
		_, err := svc.CreateRule(context.TODO(), &rulesv1.CreateScheduleRuleRequest{
			Rule: &gomoneypbv1.ScheduleRule{CronExpression: "bad"},
		})
		assert.Error(t, err)
	})

	t.Run("reinit error", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(nil)
		scheduler.EXPECT().Reinit().Return(assert.AnError)

		svc := rules.NewScheduleService(mapper, scheduler)

		_, err := svc.CreateRule(context.TODO(), &rulesv1.CreateScheduleRuleRequest{
			Rule: &gomoneypbv1.ScheduleRule{
				CronExpression: "* * * * *",
			},
		})
		assert.Error(t, err)
	})
}

func TestScheduleService_CreateRule_Error(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	scheduler := NewMockSchedulerSvc(gomock.NewController(t))

	scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(assert.AnError)

	svc := rules.NewScheduleService(mapper, scheduler)
	_, err := svc.CreateRule(context.TODO(), &rulesv1.CreateScheduleRuleRequest{
		Rule: &gomoneypbv1.ScheduleRule{CronExpression: "bad"},
	})
	assert.Error(t, err)

	scheduler = NewMockSchedulerSvc(gomock.NewController(t))
	mapper = NewMockMapperSvc(gomock.NewController(t))
	scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(nil)
	svc = rules.NewScheduleService(mapper, scheduler)
	mockGorm, _, _ := testingutils.GormMock()

	ctx := database.WithContext(context.TODO(), mockGorm)

	_, err = svc.CreateRule(ctx, &rulesv1.CreateScheduleRuleRequest{
		Rule: &gomoneypbv1.ScheduleRule{CronExpression: "* * * * *"},
	})

	assert.Error(t, err)
}

func TestScheduleService_DeleteRule_Error(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	scheduler := NewMockSchedulerSvc(gomock.NewController(t))
	svc := rules.NewScheduleService(mapper, scheduler)

	_, err := svc.DeleteRule(context.TODO(), &rulesv1.DeleteScheduleRuleRequest{Id: 999})
	assert.Error(t, err)

	rule := &database.ScheduleRule{Title: "t"}
	assert.NoError(t, gormDB.Create(rule).Error)
	scheduler.EXPECT().Reinit().Return(assert.AnError)
	_, err = svc.DeleteRule(context.TODO(), &rulesv1.DeleteScheduleRuleRequest{Id: rule.ID})
	assert.Error(t, err)
}

func TestScheduleService_UpdateRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		mapper.EXPECT().MapScheduleRule(gomock.Any()).
			DoAndReturn(func(rule *database.ScheduleRule) *gomoneypbv1.ScheduleRule {
				return &gomoneypbv1.ScheduleRule{Id: rule.ID}
			})
		scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(nil)

		rule := &database.ScheduleRule{Title: "old", Script: "old"}
		assert.NoError(t, gormDB.Create(rule).Error)

		svc := rules.NewScheduleService(mapper, scheduler)
		resp, err := svc.UpdateRule(context.TODO(), &rulesv1.UpdateScheduleRuleRequest{
			Rule: &gomoneypbv1.ScheduleRule{Id: rule.ID, Title: "new", Script: "new"},
		})
		assert.NoError(t, err)
		assert.Equal(t, rule.ID, resp.Rule.Id)

		var updated database.ScheduleRule
		assert.NoError(t, gormDB.Find(&updated, rule.ID).Error)
		assert.Equal(t, "new", updated.Title)
		assert.Equal(t, "new", updated.Script)
	})

	t.Run("save error", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(nil)

		rule := &database.ScheduleRule{Title: "old", Script: "old"}
		assert.NoError(t, gormDB.Create(rule).Error)

		svc := rules.NewScheduleService(mapper, scheduler)
		mockGorm, _, _ := testingutils.GormMock()
		ctx := database.WithContext(context.TODO(), mockGorm)

		_, err := svc.UpdateRule(ctx, &rulesv1.UpdateScheduleRuleRequest{
			Rule: &gomoneypbv1.ScheduleRule{Id: rule.ID, Title: "new", Script: "new"},
		})
		assert.Error(t, err)
	})
}

func TestScheduleService_UpdateRule_Error(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	scheduler := NewMockSchedulerSvc(gomock.NewController(t))
	scheduler.EXPECT().ValidateCronExpression(gomock.Any()).Return(assert.AnError)

	svc := rules.NewScheduleService(mapper, scheduler)
	_, err := svc.UpdateRule(context.TODO(), &rulesv1.UpdateScheduleRuleRequest{
		Rule: &gomoneypbv1.ScheduleRule{Id: 1, Script: "bad"},
	})
	assert.Error(t, err)
}

func TestScheduleService_ListRules(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	mapper.EXPECT().MapScheduleRule(gomock.Any()).
		DoAndReturn(func(rule *database.ScheduleRule) *gomoneypbv1.ScheduleRule {
			return &gomoneypbv1.ScheduleRule{Id: rule.ID}
		}).Times(2)

	scheduler := NewMockSchedulerSvc(gomock.NewController(t))
	svc := rules.NewScheduleService(mapper, scheduler)

	rulesData := []*database.ScheduleRule{
		{ID: 1, Title: "A"},
		{ID: 2, Title: "B"},
	}
	assert.NoError(t, gormDB.Create(rulesData).Error)

	resp, err := svc.ListRules(context.TODO(), &rulesv1.ListScheduleRulesRequest{
		IncludeDeleted: true,
		Ids:            []int32{1, 2},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Rules, 2)
}

func TestScheduleService_ListRules_Error(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	mapper := NewMockMapperSvc(gomock.NewController(t))
	scheduler := NewMockSchedulerSvc(gomock.NewController(t))
	svc := rules.NewScheduleService(mapper, scheduler)

	mockGorm, _, _ := testingutils.GormMock()
	ctx := database.WithContext(context.TODO(), mockGorm)

	_, err := svc.ListRules(ctx, &rulesv1.ListScheduleRulesRequest{})
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		rule := &database.ScheduleRule{Title: "to delete"}
		assert.NoError(t, gormDB.Create(rule).Error)

		mapper.EXPECT().MapScheduleRule(gomock.Any()).Return(&gomoneypbv1.ScheduleRule{Id: rule.ID})
		scheduler.EXPECT().Reinit().Return(nil)

		svc := rules.NewScheduleService(mapper, scheduler)
		_, err := svc.DeleteRule(context.TODO(), &rulesv1.DeleteScheduleRuleRequest{Id: rule.ID})
		assert.NoError(t, err)

		var deleted database.ScheduleRule
		assert.Error(t, gormDB.First(&deleted, rule.ID).Error)
	})

	t.Run("reinit error", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		rule := &database.ScheduleRule{Title: "to delete"}
		assert.NoError(t, gormDB.Create(rule).Error)

		scheduler.EXPECT().Reinit().Return(assert.AnError)

		svc := rules.NewScheduleService(mapper, scheduler)
		_, err := svc.DeleteRule(context.TODO(), &rulesv1.DeleteScheduleRuleRequest{Id: rule.ID})
		assert.Error(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		scheduler := NewMockSchedulerSvc(gomock.NewController(t))

		rule := &database.ScheduleRule{Title: "to delete"}
		assert.NoError(t, gormDB.Create(rule).Error)

		mockGorm, _, sql := testingutils.GormMock()
		ctx := database.WithContext(context.TODO(), mockGorm)

		sql.ExpectQuery("SELECT .*").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int32(1))) //empty set

		svc := rules.NewScheduleService(mapper, scheduler)
		_, err := svc.DeleteRule(ctx, &rulesv1.DeleteScheduleRuleRequest{Id: rule.ID})
		assert.ErrorContains(t, err, "all expectations were already fulfilled")
	})
}
