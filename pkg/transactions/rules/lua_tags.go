package rules

import (
	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

func (w *LuaTransactionWrapper) AddTag(l *lua.LState) int {
	if l.GetTop() != 2 {
		l.ArgError(1, "tag ID expected")
		return 0
	}

	tagID := l.CheckInt(2)
	w.tx.TagIDs = lo.Uniq(append(w.tx.TagIDs, int32(tagID)))

	w.modified = true

	return 0
}

func (w *LuaTransactionWrapper) RemoveTag(l *lua.LState) int {
	if l.GetTop() != 2 {
		l.ArgError(1, "tag ID expected")
		return 0
	}

	tagID := l.CheckInt(2)
	w.tx.TagIDs = lo.Filter(w.tx.TagIDs, func(id int32, _ int) bool {
		return id != int32(tagID)
	})

	w.modified = true

	return 0
}

func (w *LuaTransactionWrapper) GetTags(l *lua.LState) int {
	tagsTable := l.NewTable()
	for _, tagID := range w.tx.TagIDs {
		tagsTable.Append(lua.LNumber(tagID))
	}

	l.Push(tagsTable)
	return 1
}

func (w *LuaTransactionWrapper) RemoveAllTags(_ *lua.LState) int {
	w.modified = true
	w.tx.TagIDs = []int32{}
	return 0
}
