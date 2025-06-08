package transactions_test

import (
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNoGapDaily(t *testing.T) {
	t.Run("single wallet with gap and date in past", func(t *testing.T) {
		s := transactions.NewStatService()

		dateNow := time.Now().UTC()

		gamMeta, err := s.CheckDailyGapForAccount(
			gormDB,
			123,
			dateNow.AddDate(0, 0, -3),
			dateNow.AddDate(0, 0, -3),
		)

		assert.NoError(t, err)
		assert.NotEmpty(t, gamMeta)
	})
}
