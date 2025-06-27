package rules

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLuaInterpreter_Run(t *testing.T) {
	interpreter := &LuaInterpreter{}

	script := `
        print(tx:destinationAmount())
		if tx:destinationAmount() == nil then
			tx:destinationAmount(123.45)
		end
        print(tx:destinationAmount())

		tx:destinationAmount(nil)
		print(tx:destinationAmount())

		tx:addTag(1)
	`

	tx := &database.Transaction{
		Title: "abcd",
	}

	result, err := interpreter.Run(context.TODO(), script, tx)
	if err != nil {
		t.Fatalf("Failed to run Lua script: %v", err)
	}

	assert.Empty(t, result)
}
