package rules

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/yuin/gopher-lua"
)

type LuaInterpreter struct {
	cfg *LuaInterpreterConfig
}

type LuaInterpreterConfig struct {
	AccountsSvc          AccountsSvc
	DecimalSvc           DecimalSvc
	CurrencyConverterSvc CurrencyConverterSvc
}

func NewLuaInterpreter(
	cfg *LuaInterpreterConfig,
) *LuaInterpreter {
	return &LuaInterpreter{
		cfg: cfg,
	}
}

const luaTransactionType = "transactionType"
const luaHelpers = "helpersType"

func (l *LuaInterpreter) registerHelpers(ctx context.Context, state *lua.LState) {
	mt := state.NewTypeMetatable(luaHelpers)

	helpers := NewLuaHelpers(ctx, l.cfg)

	state.SetGlobal(luaHelpers, mt)
	state.SetField(mt, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"getAccountByID":  helpers.GetAccountById,
		"convertCurrency": helpers.Convert,
	}))

	ud := state.NewUserData()
	ud.Value = helpers
	state.SetMetatable(ud, state.GetTypeMetatable(luaHelpers))

	state.SetGlobal("helpers", ud)
	state.Push(ud)
}

func (l *LuaInterpreter) registerTransaction(state *lua.LState, wrapped *LuaTransactionWrapper) {
	mt := state.NewTypeMetatable(luaTransactionType)

	state.SetGlobal(luaTransactionType, mt)
	state.SetField(mt, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"title": wrapped.Title,

		"destinationAmount":                     wrapped.DestinationAmount,
		"getDestinationAmountWithDecimalPlaces": wrapped.GetDestinationAmountWithDecimalPlaces,

		"sourceAmount":                     wrapped.SourceAmount,
		"getSourceAmountWithDecimalPlaces": wrapped.GetSourceAmountWithDecimalPlaces,

		"sourceCurrency":             wrapped.SourceCurrency,
		"destinationCurrency":        wrapped.DestinationCurrency,
		"sourceAccountID":            wrapped.SourceAccountID,
		"categoryID":                 wrapped.CategoryID,
		"destinationAccountID":       wrapped.DestinationAccountID,
		"notes":                      wrapped.Notes,
		"transactionType":            wrapped.TransactionType,
		"referenceNumber":            wrapped.ReferenceNumber,
		"internalReferenceNumber":    wrapped.InternalReferenceNumber,
		"addTag":                     wrapped.AddTag,
		"removeTag":                  wrapped.RemoveTag,
		"getTags":                    wrapped.GetTags,
		"removeAllTags":              wrapped.RemoveAllTags,
		"transactionDateTimeAddDate": wrapped.TransactionDateTimeAddDate,
		"transactionDateTimeSetTime": wrapped.TransactionDateTimeSetTime,
	}))

	ud := state.NewUserData()
	ud.Value = wrapped
	state.SetMetatable(ud, state.GetTypeMetatable(luaTransactionType))

	state.SetGlobal("tx", ud)
	state.Push(ud)
}

func (l *LuaInterpreter) Run(
	ctx context.Context,
	script string,
	tx *database.Transaction,
) (bool, error) {
	state := lua.NewState()
	defer state.Close()

	wrapped := &LuaTransactionWrapper{
		tx: tx,
	}

	l.registerTransaction(state, wrapped)
	l.registerHelpers(ctx, state)

	if err := state.DoString(script); err != nil {
		return false, err
	}

	return wrapped.modified, nil
}
