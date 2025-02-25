package currency

import (
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/shopspring/decimal"
)

type DecimalService struct {
	decimalCountCache *expirable.LRU[string, int32]
}

func NewDecimalService() *DecimalService {
	return &DecimalService{
		decimalCountCache: expirable.NewLRU[string, int32](100, nil, configuration.DefaultCacheTTL),
	}
}

func (s *DecimalService) GetCurrencyDecimals(ctx context.Context, currency string) int32 {
	if cached, ok := s.decimalCountCache.Get(currency); ok {
		return cached
	}

	db := database.GetDbWithContext(ctx, database.DbTypeReadonly)

	var currencyDecimal int32
	if err := db.Model(&database.Currency{}).Where("id = ?", currency).
		Select("decimal_places").Find(&currencyDecimal).Error; err != nil {
		return configuration.DefaultDecimalPlaces
	}

	s.decimalCountCache.Add(currency, currencyDecimal)

	return currencyDecimal
}

func (s *DecimalService) ToString(ctx context.Context, amount decimal.Decimal, currency string) string {
	return amount.StringFixed(s.GetCurrencyDecimals(ctx, currency))
}
