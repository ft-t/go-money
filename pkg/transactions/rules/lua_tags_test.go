package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTagsApi(t *testing.T) {
	t.Run("add tags", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		tx:addTag(1)
		tx:addTag(2)
		tx:addTag(3)
		tx:addTag(4)
		tx:addTag(5)
	`

		tx := &database.Transaction{
			Title: "abcd",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.True(t, result)

		assert.Equal(t, 5, len(tx.TagIDs))
		assert.ElementsMatch(t, []int32{1, 2, 3, 4, 5}, tx.TagIDs)
	})

	t.Run("remove tags", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		tx:addTag(1)
		tx:addTag(2)
		tx:removeTag(2)
		tx:removeTag(3)
	`

		tx := &database.Transaction{
			Title:  "abcd",
			TagIDs: []int32{1, 2, 3},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.True(t, result)

		assert.Equal(t, 1, len(tx.TagIDs))
		assert.ElementsMatch(t, []int32{1}, tx.TagIDs)
	})

	t.Run("remove all tags", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		tx:addTag(1)
		tx:addTag(2)
		tx:removeAllTags()
	`

		tx := &database.Transaction{
			Title:  "abcd",
			TagIDs: []int32{1, 2},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.True(t, result)

		assert.Equal(t, 0, len(tx.TagIDs))
	})

	t.Run("get tags and remove single tag", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		tx:addTag(1)
		tx:addTag(2)
		tx:addTag(3)

		local tags = tx:getTags()
		print(#tags)
		tx:removeTag(2)
	`

		tx := &database.Transaction{
			Title:  "abcd",
			TagIDs: []int32{1, 2, 3},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.True(t, result)

		assert.Equal(t, 2, len(tx.TagIDs))
		assert.ElementsMatch(t, []int32{1, 3}, tx.TagIDs)
	})

	t.Run("remove tag 5 if it exists", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		local tags = tx:getTags()
		if tags[5] == nil then
			tx:removeTag(5)
		end
	`

		tx := &database.Transaction{
			Title:  "abcd",
			TagIDs: []int32{2, 3, 5},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.True(t, result)

		assert.Equal(t, 2, len(tx.TagIDs))
		assert.ElementsMatch(t, []int32{2, 3}, tx.TagIDs)
	})

	t.Run("missing tag ID (handled in lua)", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		function errHandler( err )
		   print( "ERROR:", err )
		end

		status = xpcall(function() tx:addTag() end, errHandler)
		print( status)
	`

		tx := &database.Transaction{
			Title: "abcd",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.False(t, result)

		assert.Equal(t, 0, len(tx.TagIDs))
	})

	t.Run("missing tag in add tag ID", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		tx:addTag()
	`

		tx := &database.Transaction{
			Title: "abcd",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.Error(t, err, "bad argument #1 to addTag (tag ID expected)")
		assert.False(t, result)

		assert.Equal(t, 0, len(tx.TagIDs))
	})

	t.Run("missing tag in remove tag ID", func(t *testing.T) {
		interpreter := &rules.LuaInterpreter{}

		script := `
		tx:removeTag()
	`

		tx := &database.Transaction{
			Title: "abcd",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.Error(t, err, "bad argument #1 to removeTag (tag ID expected)")
		assert.False(t, result)

		assert.Equal(t, 0, len(tx.TagIDs))
	})
}
