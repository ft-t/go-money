package handlers

import (
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type MaintenanceApi struct {
	recalculateSvc RecalculateSvc
}

func NewMaintenanceApi(
	mux *boilerplate.DefaultGrpcServer,
	recalculateSvc RecalculateSvc,
) *MaintenanceApi {
	res := &MaintenanceApi{
		recalculateSvc: recalculateSvc,
	}

	mux.GetMux().Handle(
		NewMaintenance(res, mux.GetDefaultHandlerOptions()...),
	)

	return res
}
