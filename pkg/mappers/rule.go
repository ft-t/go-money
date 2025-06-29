package mappers

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *Mapper) MapRule(rule *database.Rule) *gomoneypbv1.Rule {
	mapped := &gomoneypbv1.Rule{
		Id:          rule.ID,
		Title:       rule.Title,
		Script:      rule.Script,
		Interpreter: gomoneypbv1.RuleInterpreterType(rule.InterpreterType),
		SortOrder:   rule.SortOrder,
		CreatedAt:   timestamppb.New(rule.CreatedAt),
		UpdatedAt:   timestamppb.New(rule.UpdatedAt),
		Enabled:     rule.Enabled,
		IsFinalRule: rule.IsFinalRule,
		GroupName:   rule.GroupName,
	}

	if rule.DeletedAt.Valid {
		mapped.DeletedAt = timestamppb.New(rule.DeletedAt.Time)
	}

	return mapped
}
