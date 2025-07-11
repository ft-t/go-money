package jobs

import (
	"context"
	"github.com/rs/zerolog"
)

func (j *JobScheduler) FixDailyGap(ctx context.Context) error {
	ctx = zerolog.Ctx(ctx).With().Str("job", "fix_daily_gap").Logger().WithContext(ctx)
	zerolog.Ctx(ctx).Info().Msg("Starting daily gap fix job")

	return j.cfg.MaintenanceSvc.FixDailyGaps(ctx)
}
