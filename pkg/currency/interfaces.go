package currency

import (
	"context"
	"gorm.io/gorm"
	"net/http"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package currency_test -source=interfaces.go

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type BaseAmountSvc interface {
	RecalculateAmountInBaseCurrencyForAll(
		ctx context.Context,
		tx *gorm.DB,
	) error
}
