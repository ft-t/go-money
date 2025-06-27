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
        print(current:destinationAmount())
		if current:destinationAmount() == nil then
			current:destinationAmount(123.45)
		end
        print(current:destinationAmount())

		current:destinationAmount(nil)
		print(current:destinationAmount())
	`

	result, err := interpreter.Run(context.TODO(), script, &database.Transaction{
		Title: "abcd",
	})
	if err != nil {
		t.Fatalf("Failed to run Lua script: %v", err)
	}

	assert.Empty(t, result)
}
