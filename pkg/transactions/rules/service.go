package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm"
	"time"
)

type Service struct {
	mapper MapperSvc
}

func NewService(
	mapper MapperSvc,
) *Service {
	return &Service{
		mapper: mapper,
	}
}

func (s *Service) DeleteRule(ctx context.Context, req *rulesv1.DeleteRuleRequest) (*rulesv1.DeleteRuleResponse, error) {
	var rule database.Rule

	if err := database.GetDbWithContext(ctx, database.DbTypeMaster).Delete(&rule).Error; err != nil {
		return nil, err
	}

	return &rulesv1.DeleteRuleResponse{
		Rule: s.mapper.MapRule(&rule),
	}, nil
}

func (s *Service) mapRule(rule *gomoneypbv1.Rule) *database.Rule {
	mapped := &database.Rule{
		Title:           rule.Title,
		Script:          rule.Script,
		InterpreterType: rule.Interpreter,
		SortOrder:       rule.SortOrder,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		Enabled:         rule.Enabled,
		IsFinalRule:     rule.IsFinalRule,
		GroupName:       rule.GroupName,
	}

	if rule.DeletedAt != nil {
		mapped.DeletedAt = gorm.DeletedAt{
			Time:  rule.DeletedAt.AsTime(),
			Valid: true,
		}
	}

	return mapped
}

func (s *Service) CreateRule(ctx context.Context, req *rulesv1.CreateRuleRequest) (*rulesv1.CreateRuleResponse, error) {
	newRule := s.mapRule(req.Rule)

	newRule.ID = 0
	newRule.CreatedAt = time.Now().UTC()
	newRule.UpdatedAt = time.Now().UTC()

	if err := database.GetDbWithContext(ctx, database.DbTypeMaster).Create(newRule).Error; err != nil {
		return nil, err
	}

	return &rulesv1.CreateRuleResponse{
		Rule: s.mapper.MapRule(newRule),
	}, nil
}

func (s *Service) ListRules(ctx context.Context, req *rulesv1.ListRulesRequest) (*rulesv1.ListRulesResponse, error) {
	var rules []*database.Rule

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

	mappedRules := make([]*gomoneypbv1.Rule, 0, len(rules))
	for _, rule := range rules {
		mappedRules = append(mappedRules, s.mapper.MapRule(rule))
	}

	return &rulesv1.ListRulesResponse{
		Rules: mappedRules,
	}, nil
}

func (s *Service) UpdateRule(ctx context.Context, req *rulesv1.UpdateRuleRequest) (*rulesv1.UpdateRuleResponse, error) {
	updatedRule := s.mapRule(req.Rule)

	updatedRule.UpdatedAt = time.Now().UTC()

	if err := database.GetDbWithContext(ctx, database.DbTypeMaster).Save(updatedRule).Error; err != nil {
		return nil, err
	}

	return &rulesv1.UpdateRuleResponse{
		Rule: s.mapper.MapRule(updatedRule),
	}, nil
}
