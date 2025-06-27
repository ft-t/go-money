package rules

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"github.com/yuin/gopher-lua"
)

type LuaInterpreter struct {
}

type LuaTransactionWrapper struct {
	Tx *database.Transaction
}

func (w *LuaTransactionWrapper) Title(l *lua.LState) int {
	return w.getSetStringField(l, w.Tx.Title, func(val string) {
		w.Tx.Title = val
	})
}

func (w *LuaTransactionWrapper) DestinationAmount(l *lua.LState) int {
	//return w.getSetNullDecimalField(l, w.check(l).Tx.DestinationAmount, func(nullDecimal decimal.NullDecimal) {
	return w.getSetNullDecimalField(l, w.Tx.DestinationAmount, func(nullDecimal decimal.NullDecimal) {
		w.Tx.DestinationAmount = nullDecimal
	})
}

func (w *LuaTransactionWrapper) SourceAmount(l *lua.LState) int {
	return w.getSetNullDecimalField(l, w.Tx.SourceAmount, func(nullDecimal decimal.NullDecimal) {
		w.Tx.SourceAmount = nullDecimal
	})
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

const luaTypeName = "transaction"

func registerType(state *lua.LState, wrapped *LuaTransactionWrapper) {
	mt := state.NewTypeMetatable(luaTypeName)

	state.SetGlobal(luaTypeName, mt)
	state.SetField(mt, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"title":             wrapped.Title,
		"destinationAmount": wrapped.DestinationAmount,
		"sourceAmount":      wrapped.SourceAmount,
	}))

	ud := state.NewUserData()
	ud.Value = wrapped
	state.SetMetatable(ud, state.GetTypeMetatable(luaTypeName))

	state.SetGlobal("current", ud)
	state.Push(ud)
}

func (l *LuaInterpreter) Run(
	_ context.Context,
	script string,
	tx *database.Transaction,
) (lua.LValue, error) {
	state := lua.NewState()
	defer state.Close()

	wrapped := &LuaTransactionWrapper{
		Tx: tx,
	}

	registerType(state, wrapped)

	if err := state.DoString(script); err != nil {
		return nil, err
	}

	return nil, nil
}
