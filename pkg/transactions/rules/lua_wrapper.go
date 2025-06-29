package rules

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

type LuaTransactionWrapper struct {
	tx       *database.Transaction
	modified bool
}

func (w *LuaTransactionWrapper) Title(l *lua.LState) int {
	return w.getSetStringField(l, w.tx.Title, func(val string) {
		w.tx.Title = val
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

func (w *LuaTransactionWrapper) getSetNullInt32Field(l *lua.LState, val *int32, setter func(*int32)) int {
	if l.GetTop() == 2 { // set
		w.modified = true
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

func (w *LuaTransactionWrapper) getSetStringField(l *lua.LState, val string, setter func(string)) int {
	if l.GetTop() == 2 { // set
		w.modified = true
		setter(l.CheckString(2))
		return 0
	}

	l.Push(lua.LString(val))
	return 1
}
