package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"time"
)

type ScheduledService struct {
	mapper    MapperSvc
	scheduler SchedulerSvc
}

func NewScheduledService(
	mapper MapperSvc,
	scheduler SchedulerSvc,
) *ScheduledService {
	return &ScheduledService{
		mapper:    mapper,
		scheduler: scheduler,
	}
}

func (s *ScheduledService) DeleteRule(
	ctx context.Context,
	req *rulesv1.DeleteScheduledRuleRequest,
) (*rulesv1.DeleteScheduledRuleResponse, error) {
	var rule database.ScheduleRule

	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	if err := db.Where("id = ?", req.Id).First(&rule).Error; err != nil {
		return nil, err
	}

	if err := db.Delete(&rule).Error; err != nil {
		return nil, err
	}

	if err := s.scheduler.Reinit(); err != nil {
		return nil, err
	}

	return &rulesv1.DeleteScheduledRuleResponse{
		Rule: s.mapper.MapScheduleRule(&rule),
	}, nil
}

func (s *ScheduledService) CreateRule(
	ctx context.Context,
	req *rulesv1.CreateScheduledRuleRequest,
) (*rulesv1.CreateScheduledRuleResponse, error) {
	newRule := s.mapRule(req.Rule)

	newRule.ID = 0
	newRule.CreatedAt = time.Now().UTC()
	newRule.UpdatedAt = time.Now().UTC()

	if err := s.scheduler.ValidateCronExpression(newRule.CronExpression); err != nil {
		return nil, err
	}

	if err := database.GetDbWithContext(ctx, database.DbTypeMaster).Create(newRule).Error; err != nil {
		return nil, err
	}

	if err := s.scheduler.Reinit(); err != nil {
		return nil, err
	}

	return &rulesv1.CreateScheduledRuleResponse{
		Rule: s.mapper.MapScheduleRule(newRule),
	}, nil
}

func (s *ScheduledService) ListRules(
	ctx context.Context,
	req *rulesv1.ListRulesRequest,
) (*rulesv1.ListRulesResponse, error) {
	var rules []*database.ScheduleRule

	query := database.GetDbWithContext(ctx, database.DbTypeMaster).Order("sort_order")

	if req.IncludeDeleted {
		query = query.Unscoped()
	}

	if len(req.Ids) > 0 {
		query = query.Where("id IN ?", req.Ids)
	}

	if err := query.Find(&rules).Error; err != nil {
		return nil, err
	}

	mappedRules := make([]*gomoneypbv1.ScheduleRule, 0, len(rules))
	for _, rule := range rules {
		mappedRules = append(mappedRules, s.mapper.MapScheduleRule(rule))
	}

	return &rulesv1.ListRulesResponse{
		Rules: nil,
		//Rules: mappedRules,
	}, nil
}

func (s *ScheduledService) UpdateRule(
	ctx context.Context,
	req *rulesv1.UpdateScheduledRuleRequest,
) (*rulesv1.UpdateScheduledRuleResponse, error) {
	updatedRule := s.mapRule(req.Rule)

	updatedRule.UpdatedAt = time.Now().UTC()

	if err := s.scheduler.ValidateCronExpression(updatedRule.Script); err != nil {
		return nil, err
	}

	if err := database.GetDbWithContext(ctx, database.DbTypeMaster).Save(updatedRule).Error; err != nil {
		return nil, err
	}

	return &rulesv1.UpdateScheduledRuleResponse{
		Rule: s.mapper.MapScheduleRule(updatedRule),
	}, nil
}

func (s *ScheduledService) mapRule(rule *gomoneypbv1.ScheduleRule) *database.ScheduleRule {
	mapped := &database.ScheduleRule{
		ID:              rule.Id,
		Title:           rule.Title,
		Script:          rule.Script,
		InterpreterType: rule.Interpreter,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		Enabled:         rule.Enabled,
		GroupName:       rule.GroupName,
		CronExpression:  rule.CronExpression,
	}

	return mapped
}
