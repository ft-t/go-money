package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"context"
)

type DryRun struct {
}

func NewDryRun() *DryRun {
	return &DryRun{}
}

func (s *DryRun) DryRunRule(ctx context.Context, req *rulesv1.DryRunRuleRequest) (*rulesv1.DryRunRuleResponse, error) {
	//TODO implement me
	panic("implement me")
}
