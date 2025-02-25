package appcfg

import "context"

//go:generate mockgen -destination interfaces_mocks_test.go -package appcfg_test -source=interfaces.go

type UserSvc interface {
	ShouldCreateAdmin(ctx context.Context) (bool, error)
}
