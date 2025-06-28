package rules

import "github.com/ft-t/go-money/pkg/database"

type RuleGroup struct {
	Name  string
	Rules []*database.Rule
}
