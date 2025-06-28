package rules

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/yuin/gopher-lua"
)

type LuaInterpreter struct {
}

func NewLuaInterpreter() *LuaInterpreter {
	return &LuaInterpreter{}
}

const luaTypeName = "transactionType"

func registerType(state *lua.LState, wrapped *LuaTransactionWrapper) {
	mt := state.NewTypeMetatable(luaTypeName)

	state.SetGlobal(luaTypeName, mt)
	state.SetField(mt, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"title": wrapped.Title,

		"destinationAmount":                     wrapped.DestinationAmount,
		"getDestinationAmountWithDecimalPlaces": wrapped.GetDestinationAmountWithDecimalPlaces,
		
		"sourceAmount":                     wrapped.SourceAmount,
		"getSourceAmountWithDecimalPlaces": wrapped.GetSourceAmountWithDecimalPlaces,

		"sourceCurrency":          wrapped.SourceCurrency,
		"destinationCurrency":     wrapped.DestinationCurrency,
		"sourceAccountID":         wrapped.SourceAccountID,
		"destinationAccountID":    wrapped.DestinationAccountID,
		"notes":                   wrapped.Notes,
		"transactionType":         wrapped.TransactionType,
		"referenceNumber":         wrapped.ReferenceNumber,
		"internalReferenceNumber": wrapped.InternalReferenceNumber,
		"addTag":                  wrapped.AddTag,
		"removeTag":               wrapped.RemoveTag,
		"getTags":                 wrapped.GetTags,
		"removeAllTags":           wrapped.RemoveAllTags,
	}))

	ud := state.NewUserData()
	ud.Value = wrapped
	state.SetMetatable(ud, state.GetTypeMetatable(luaTypeName))

	state.SetGlobal("tx", ud)
	state.Push(ud)
}

func (l *LuaInterpreter) Run(
	_ context.Context,
	script string,
	tx *database.Transaction,
) (bool, error) {
	state := lua.NewState()
	defer state.Close()

	wrapped := &LuaTransactionWrapper{
		tx: tx,
	}

	registerType(state, wrapped)

	if err := state.DoString(script); err != nil {
		return false, err
	}

	return wrapped.modified, nil
}
