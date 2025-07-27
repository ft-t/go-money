package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"time"
)

type ScheduleService struct {
	mapper    MapperSvc
	scheduler SchedulerSvc
}

func NewScheduleService(
	mapper MapperSvc,
	scheduler SchedulerSvc,
) *ScheduleService {
	return &ScheduleService{
		mapper:    mapper,
		scheduler: scheduler,
	}
}

func (s *ScheduleService) DeleteRule(
	ctx context.Context,
	req *rulesv1.DeleteScheduleRuleRequest,
) (*rulesv1.DeleteScheduleRuleResponse, error) {
	var rule database.ScheduleRule

	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))

	if err := db.Where("id = ?", req.Id).First(&rule).Error; err != nil {
		return nil, err
	}

	if err := db.Delete(&rule).Error; err != nil {
		return nil, err
	}

	if err := s.scheduler.Reinit(ctx); err != nil {
		return nil, err
	}

	return &rulesv1.DeleteScheduleRuleResponse{
		Rule: s.mapper.MapScheduleRule(&rule),
	}, nil
}

func (s *ScheduleService) CreateRule(
	ctx context.Context,
	req *rulesv1.CreateScheduleRuleRequest,
) (*rulesv1.CreateScheduleRuleResponse, error) {
	newRule := s.mapRule(req.Rule)

	newRule.ID = 0
	newRule.CreatedAt = time.Now().UTC()
	newRule.UpdatedAt = time.Now().UTC()

	if err := s.scheduler.ValidateCronExpression(newRule.CronExpression); err != nil {
		return nil, err
	}

	if err := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).
		Create(newRule).Error; err != nil {
		return nil, err
	}

	if err := s.scheduler.Reinit(ctx); err != nil {
		return nil, err
	}

	return &rulesv1.CreateScheduleRuleResponse{
		Rule: s.mapper.MapScheduleRule(newRule),
	}, nil
}

func (s *ScheduleService) ListRules(
	ctx context.Context,
	req *rulesv1.ListScheduleRulesRequest,
) (*rulesv1.ListScheduleRulesResponse, error) {
	var rules []*database.ScheduleRule

	query := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).
		Order("id desc")

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

	return &rulesv1.ListScheduleRulesResponse{
		Rules: mappedRules,
	}, nil
}

func (s *ScheduleService) UpdateRule(
	ctx context.Context,
	req *rulesv1.UpdateScheduleRuleRequest,
) (*rulesv1.UpdateScheduleRuleResponse, error) {
	updatedRule := s.mapRule(req.Rule)

	updatedRule.UpdatedAt = time.Now().UTC()

	if err := s.scheduler.ValidateCronExpression(updatedRule.CronExpression); err != nil {
		return nil, err
	}

	if err := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Save(updatedRule).Error; err != nil {
		return nil, err
	}

	if err := s.scheduler.Reinit(ctx); err != nil {
		return nil, err
	}

	return &rulesv1.UpdateScheduleRuleResponse{
		Rule: s.mapper.MapScheduleRule(updatedRule),
	}, nil
}

func (s *ScheduleService) mapRule(rule *gomoneypbv1.ScheduleRule) *database.ScheduleRule {
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
