package mappers_test

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestMapRule(t *testing.T) {
	mapper := mappers.NewMapper(&mappers.MapperConfig{})

	created := time.Now().UTC()
	updated := created.Add(1 * time.Hour)
	deleted := created.Add(2 * time.Hour)

	source := &database.Rule{
		ID:              123,
		Title:           "test",
		Script:          "script1",
		InterpreterType: gomoneypbv1.RuleInterpreterType_RULE_INTERPRETER_TYPE_LUA,
		SortOrder:       555,
		CreatedAt:       created,
		UpdatedAt:       updated,
		Enabled:         true,
		IsFinalRule:     true,
		DeletedAt: gorm.DeletedAt{
			Time:  deleted,
			Valid: true,
		},
		GroupName: "abcd",
	}

	dest := mapper.MapRule(source)

	assert.EqualValues(t, source.ID, dest.Id)
	assert.Equal(t, source.Title, dest.Title)
	assert.Equal(t, source.Script, dest.Script)
	assert.Equal(t, source.InterpreterType, dest.Interpreter)
	assert.EqualValues(t, source.SortOrder, dest.SortOrder)
	assert.EqualValues(t, source.CreatedAt, dest.CreatedAt.AsTime())
	assert.EqualValues(t, source.UpdatedAt, dest.UpdatedAt.AsTime())
	assert.Equal(t, source.Enabled, dest.Enabled)
	assert.Equal(t, source.IsFinalRule, dest.IsFinalRule)
	assert.Equal(t, source.GroupName, dest.GroupName)
	assert.EqualValues(t, source.DeletedAt.Time, source.DeletedAt.Time)
}

func TestMapScheduleRule(t *testing.T) {
	mapper := mappers.NewMapper(&mappers.MapperConfig{})

	rule := &database.ScheduleRule{
		ID:             123,
		Title:          "test",
		Script:         "script1",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Enabled:        true,
		CronExpression: "0 0 * * *",
		LastRunAt:      nil,
		GroupName:      "abcd",
		DeletedAt: gorm.DeletedAt{
			Valid: true,
			Time:  time.Now().UTC().Add(1 * time.Hour),
		},
	}

	resp := mapper.MapScheduleRule(rule)

	assert.EqualValues(t, rule.ID, resp.Id)
	assert.Equal(t, rule.Title, resp.Title)
	assert.Equal(t, rule.Script, resp.Script)
	assert.Equal(t, rule.CreatedAt, resp.CreatedAt.AsTime())
	assert.Equal(t, rule.UpdatedAt, resp.UpdatedAt.AsTime())
	assert.Equal(t, rule.Enabled, resp.Enabled)
	assert.Equal(t, rule.CronExpression, resp.CronExpression)
	assert.Equal(t, rule.GroupName, resp.GroupName)
}
