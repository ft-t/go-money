package currency

import (
	currencyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/currency/v1"
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"time"
)

type Service struct {
	configuration configuration.CurrencyConfig
}

func NewService(cfg configuration.CurrencyConfig) *Service {
	return &Service{
		configuration: cfg,
	}
}

func (s *Service) DeleteCurrency(
	ctx context.Context,
	req *currencyv1.DeleteCurrencyRequest,
) (*currencyv1.DeleteCurrencyResponse, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	var currency database.Currency
	if err := db.Where("id = ?", req.Id).First(&currency).Error; err != nil {
		return nil, err
	}

	if err := db.Delete(&currency).Error; err != nil {
		return nil, err
	}

	return &currencyv1.DeleteCurrencyResponse{
		Currency: s.mapCurrency(&currency),
	}, nil
}

func (s *Service) GetCurrencies(
	ctx context.Context,
	req *currencyv1.GetCurrenciesRequest,
) (*currencyv1.GetCurrenciesResponse, error) {
	var currencies []*database.Currency

	db := database.FromContext(ctx, database.GetDb(database.DbTypeReadonly))

	query := db.Unscoped().Order("is_active desc, id desc")

	if len(req.Ids) > 0 {
		query = query.Where("id in ?", req.Ids)
	}

	if !req.IncludeDisabled {
		query = query.Where("is_active = true")
	}

	if err := query.Find(&currencies).Error; err != nil {
		return nil, err
	}

	var mapped []*v1.Currency
	for _, currency := range currencies {
		mapped = append(mapped, s.mapCurrency(currency))
	}

	return &currencyv1.GetCurrenciesResponse{
		Currencies: mapped,
	}, nil
}

func (s *Service) CreateCurrency(
	ctx context.Context,
	req *currencyv1.CreateCurrencyRequest,
) (*currencyv1.CreateCurrencyResponse, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	currency := &database.Currency{
		ID:            req.Currency.Id,
		Rate:          decimal.Decimal{},
		IsActive:      req.Currency.IsActive,
		DecimalPlaces: req.Currency.DecimalPlaces,
		UpdatedAt:     time.Now().UTC(),
	}

	rate, err := decimal.NewFromString(req.Currency.Rate)
	if err != nil {
		return nil, err
	}

	currency.Rate = rate

	if err = db.Create(currency).Error; err != nil {
		return nil, err
	}

	return &currencyv1.CreateCurrencyResponse{
		Currency: s.mapCurrency(currency),
	}, nil
}

func (s *Service) UpdateCurrency(
	ctx context.Context,
	req *currencyv1.UpdateCurrencyRequest,
) (*currencyv1.UpdateCurrencyResponse, error) {
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	var currency database.Currency
	if err := db.Where("id = ?", req.Id).First(&currency).Error; err != nil {
		return nil, err
	}

	rate, err := decimal.NewFromString(req.Rate)
	if err != nil {
		return nil, err
	}

	currency.Rate = rate
	currency.IsActive = req.IsActive
	currency.DecimalPlaces = req.DecimalPlaces
	currency.UpdatedAt = time.Now().UTC()
	currency.DeletedAt = gorm.DeletedAt{
		Valid: false,
	}

	if currency.ID == s.configuration.BaseCurrency {
		currency.Rate = decimal.NewFromInt(1) // always 1
	}

	if err = db.Save(&currency).Error; err != nil {
		return nil, err
	}

	return &currencyv1.UpdateCurrencyResponse{
		Currency: s.mapCurrency(&currency),
	}, nil
}

func (s *Service) mapCurrency(currency *database.Currency) *v1.Currency {
	cr := &v1.Currency{
		Id:            currency.ID,
		Rate:          currency.Rate.StringFixed(currency.DecimalPlaces),
		IsActive:      currency.IsActive,
		DecimalPlaces: currency.DecimalPlaces,
		UpdatedAt:     timestamppb.New(currency.UpdatedAt),
	}

	if currency.DeletedAt.Valid {
		cr.DeletedAt = timestamppb.New(currency.DeletedAt.Time)
	}

	return cr
}
