package jobs

import "context"

type MaintenanceSvc interface {
	FixDailyGaps(
		ctx context.Context,
	) error
}
