package rules

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	lua "github.com/yuin/gopher-lua"
)

type LuaTransactionWrapper struct {
	tx *database.Transaction
}

func (w *LuaTransactionWrapper) Title(l *lua.LState) int {
	return w.getSetStringField(l, w.tx.Title, func(val string) {
		w.tx.Title = val
	})
}

func (w *LuaTransactionWrapper) DestinationAmount(l *lua.LState) int {
	return w.getSetNullDecimalField(l, w.tx.DestinationAmount, func(nullDecimal decimal.NullDecimal) {
		w.tx.DestinationAmount = nullDecimal
	})
}

func (w *LuaTransactionWrapper) SourceAmount(l *lua.LState) int {
	return w.getSetNullDecimalField(l, w.tx.SourceAmount, func(nullDecimal decimal.NullDecimal) {
		w.tx.SourceAmount = nullDecimal
	})
}

func (w *LuaTransactionWrapper) SourceCurrency(l *lua.LState) int {
	return w.getSetStringField(l, w.tx.SourceCurrency, func(val string) {
		w.tx.SourceCurrency = val
	})
}

func (w *LuaTransactionWrapper) DestinationCurrency(l *lua.LState) int {
	return w.getSetStringField(l, w.tx.DestinationCurrency, func(val string) {
		w.tx.DestinationCurrency = val
	})
}

func (w *LuaTransactionWrapper) SourceAccountID(l *lua.LState) int {
	return w.getSetNullInt32Field(l, w.tx.SourceAccountID, func(val *int32) {
		w.tx.SourceAccountID = val
	})
}

func (w *LuaTransactionWrapper) DestinationAccountID(l *lua.LState) int {
	return w.getSetNullInt32Field(l, w.tx.DestinationAccountID, func(val *int32) {
		w.tx.DestinationAccountID = val
	})
}

func (w *LuaTransactionWrapper) Notes(l *lua.LState) int {
	return w.getSetStringField(l, w.tx.Notes, func(val string) {
		w.tx.Notes = val
	})
}

func (w *LuaTransactionWrapper) TransactionType(l *lua.LState) int {
	return w.getSetNullInt32Field(l, lo.ToPtr(int32(w.tx.TransactionType)), func(val *int32) {
		if val == nil {
			w.tx.TransactionType = 0 // reset to default
		} else {
			w.tx.TransactionType = gomoneypbv1.TransactionType(*val)
		}
	})
}

func (w *LuaTransactionWrapper) ReferenceNumber(l *lua.LState) int {
	return w.getSetStringField(l, lo.FromPtr(w.tx.ReferenceNumber), func(val string) {
		w.tx.ReferenceNumber = lo.ToPtr(val)
	})
}

func (w *LuaTransactionWrapper) InternalReferenceNumber(l *lua.LState) int {
	return w.getSetStringField(l, lo.FromPtr(w.tx.InternalReferenceNumber), func(val string) {
		w.tx.InternalReferenceNumber = lo.ToPtr(val)
	})
}

func (w *LuaTransactionWrapper) AddTag(l *lua.LState) int {
	if l.GetTop() != 2 {
		l.ArgError(1, "tag ID expected")
		return 0
	}

	tagID := l.CheckInt(2)
	w.tx.TagIDs = append(w.tx.TagIDs, int32(tagID))

	lo.Uniq(w.tx.TagIDs)

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

	return 0
}

func (w *LuaTransactionWrapper) GetTags(l *lua.LState) int {
	if len(w.tx.TagIDs) == 0 {
		l.Push(lua.LNil)
		return 1
	}

	tagsTable := l.NewTable()
	for _, tagID := range w.tx.TagIDs {
		tagsTable.Append(lua.LNumber(tagID))
	}

	l.Push(tagsTable)
	return 1
}

func (w *LuaTransactionWrapper) RemoveAllTags(_ *lua.LState) int {
	w.tx.TagIDs = []int32{}
	return 0
}

func (w *LuaTransactionWrapper) getSetNullInt32Field(l *lua.LState, val *int32, setter func(*int32)) int {
	if l.GetTop() == 2 { // set
		luaVal := l.Get(2)

		if luaVal == lua.LNil {
			setter(nil) // unset
			return 0
		}

		num := l.CheckInt(2)
		setter(lo.ToPtr(int32(num)))

		return 0
	}

	if val != nil {
		l.Push(lua.LNumber(*val))
	} else {
		l.Push(lua.LNil)
	}

	return 1
}

func (w *LuaTransactionWrapper) getSetNullDecimalField(l *lua.LState, val decimal.NullDecimal, setter func(decimal.NullDecimal)) int {
	if l.GetTop() == 2 { // set
		luaVal := l.Get(2)

		if luaVal == lua.LNil {
			setter(decimal.NullDecimal{}) // unset
			return 0
		}

		num := l.CheckNumber(2)
		d := decimal.NewFromFloat(float64(num))
		setter(decimal.NewNullDecimal(d))

		return 0
	}

	if val.Valid {
		l.Push(lua.LNumber(val.Decimal.InexactFloat64()))
	} else {
		l.Push(lua.LNil)
	}

	return 1
}

func (w *LuaTransactionWrapper) getSetStringField(l *lua.LState, val string, setter func(string)) int {
	if l.GetTop() == 2 { // set
		setter(l.CheckString(2))
		return 0
	}

	l.Push(lua.LString(val))
	return 1
}

func (w *LuaTransactionWrapper) check(L *lua.LState) *LuaTransactionWrapper {
	ud := L.CheckUserData(1)

	if v, ok := ud.Value.(*LuaTransactionWrapper); ok {
		return v
	}

	L.ArgError(1, "Transaction expected")

	return nil
}
