package currency

import "github.com/shopspring/decimal"

type DecimalService struct {
}

func NewDecimalService() *DecimalService {
	return &DecimalService{}
}

func (s *DecimalService) GetCurrencyDecimals(currency string) int32 {
	return 2
}

func (s *DecimalService) ToString(amount decimal.Decimal, currency string) string {
	return amount.StringFixed(s.GetCurrencyDecimals(currency))
}
