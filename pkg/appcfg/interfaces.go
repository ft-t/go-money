package appcfg

import "context"

type UserSvc interface {
	ShouldCreateAdmin(ctx context.Context) (bool, error)
}
