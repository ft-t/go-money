package rules

import (
	"github.com/shopspring/decimal"
	lua "github.com/yuin/gopher-lua"
)

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

func (w *LuaTransactionWrapper) GetSourceAmountWithDecimalPlaces(l *lua.LState) int {
	return w.getAmountWithDecimalPlaces(l, w.tx.SourceAmount)
}

func (w *LuaTransactionWrapper) GetDestinationAmountWithDecimalPlaces(l *lua.LState) int {
	return w.getAmountWithDecimalPlaces(l, w.tx.DestinationAmount)
}

func (w *LuaTransactionWrapper) getAmountWithDecimalPlaces(
	l *lua.LState,
	amount decimal.NullDecimal,
) int {
	if l.GetTop() != 2 {
		l.ArgError(1, "decimalPlaces expected")
		return 0
	}

	decimalPlaces := l.CheckInt(2)

	if amount.Valid {
		l.Push(lua.LNumber(amount.Decimal.Round(int32(decimalPlaces)).InexactFloat64()))
	} else {
		l.Push(lua.LNil)
	}
	return 1
}

func (w *LuaTransactionWrapper) getSetNullDecimalField(l *lua.LState, val decimal.NullDecimal, setter func(decimal.NullDecimal)) int {
	if l.GetTop() == 2 { // set
		w.modified = true
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
