package rules

import (
	lua "github.com/yuin/gopher-lua"
	"time"
)

func (w *LuaTransactionWrapper) TransactionDateTimeAddDate(l *lua.LState) int {
	if l.GetTop() != 4 { // set
		l.ArgError(1, "expected 3 arguments")
		return 0
	}

	w.modified = true

	year := l.CheckInt(2)
	month := l.CheckInt(3)
	day := l.CheckInt(4)

	w.tx.TransactionDateTime = w.tx.TransactionDateTime.AddDate(year, month, day)
	w.tx.TransactionDateOnly = w.tx.TransactionDateTime

	time.Date(w.tx.TransactionDateTime.Year(), w.tx.TransactionDateTime.Month(),
		w.tx.TransactionDateTime.Day(), 0, 0, 0, 0, w.tx.TransactionDateTime.Location(),
	)

	return 0
}

func (w *LuaTransactionWrapper) TransactionDateTimeSetTime(l *lua.LState) int {
	if l.GetTop() != 3 { // set
		l.ArgError(1, "expected 2 arguments")
		return 0
	}

	w.modified = true

	hour := l.CheckInt(2)
	minute := l.CheckInt(3)

	w.tx.TransactionDateTime = time.Date(
		w.tx.TransactionDateTime.Year(),
		w.tx.TransactionDateTime.Month(),
		w.tx.TransactionDateTime.Day(),
		hour, minute, 0, 0,
		w.tx.TransactionDateTime.Location(),
	)
	w.tx.TransactionDateOnly = w.tx.TransactionDateTime

	return 0
}
