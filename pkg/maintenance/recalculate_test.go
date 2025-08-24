package maintenance_test

import (
	"context"
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/maintenance"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRecalculate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		rec := maintenance.NewRecalculateService(&maintenance.RecalculateServiceConfig{
			AccountSvc:     accSvc,
			TransactionSvc: txSvc,
		})

		accSvc.EXPECT().GetAllAccounts(gomock.All()).
			Return([]*database.Account{
				{
					ID: 5,
				},
			}, nil)

		txSvc.EXPECT().StoreStat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Any()).
			Return(nil)

		assert.NoError(t, rec.RecalculateAll(context.TODO()))
	})

	t.Run("get all accounts error", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		rec := maintenance.NewRecalculateService(&maintenance.RecalculateServiceConfig{
			AccountSvc:     accSvc,
			TransactionSvc: txSvc,
		})

		accSvc.EXPECT().GetAllAccounts(gomock.Any()).
			Return(nil, assert.AnError)

		assert.ErrorIs(t, rec.RecalculateAll(context.TODO()), assert.AnError)
	})

	t.Run("store stat error", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		rec := maintenance.NewRecalculateService(&maintenance.RecalculateServiceConfig{
			AccountSvc:     accSvc,
			TransactionSvc: txSvc,
		})

		accSvc.EXPECT().GetAllAccounts(gomock.Any()).
			Return([]*database.Account{
				{
					ID: 5,
				},
			}, nil)

		txSvc.EXPECT().StoreStat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Any()).
			Return(assert.AnError)

		assert.ErrorIs(t, rec.RecalculateAll(context.TODO()), assert.AnError)
	})
}
