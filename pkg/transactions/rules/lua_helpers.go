package rules

import (
	"context"
	"github.com/shopspring/decimal"
	lua "github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
)

type LuaHelpers struct {
	ctx context.Context
	cfg *LuaInterpreterConfig
}

func NewLuaHelpers(
	ctx context.Context,
	cfg *LuaInterpreterConfig,
) *LuaHelpers {
	return &LuaHelpers{
		ctx: ctx,
		cfg: cfg,
	}
}

func (h *LuaHelpers) Convert(l *lua.LState) int {
	if l.GetTop() != 4 {
		l.ArgError(1, "from, to, amount are expected")
		return 0
	}

	from := l.CheckString(2)
	to := l.CheckString(3)
	amount := l.CheckNumber(4)

	converted, err := h.cfg.CurrencyConverterSvc.Convert(h.ctx,
		from,
		to,
		decimal.NewFromFloat(float64(amount)),
	)
	if err != nil {
		l.RaiseError("failed to convert currency: %v", err)
		return 0
	}

	l.Push(lua.LNumber(converted.Round(h.cfg.DecimalSvc.GetCurrencyDecimals(h.ctx, to)).
		InexactFloat64()))

	return 1
}

func (h *LuaHelpers) GetAccountById(l *lua.LState) int {
	if l.GetTop() != 2 {
		l.ArgError(1, "account ID expected")
		return 0
	}

	accountID := l.CheckInt(2)

	account, err := h.cfg.AccountsSvc.GetAccountByID(h.ctx, int32(accountID))
	if err != nil {
		l.RaiseError("failed to get account: %v", err)
		return 0
	}

	l.Push(luar.New(l, account))

	return 1
}
