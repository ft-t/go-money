package handlers

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/import/v1/importv1connect"
	"buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type ImportApi struct {
	importSvc ImportSvc
}

func NewImportApi(
	mux *boilerplate.DefaultGrpcServer,
	importSvc ImportSvc,
) (*ImportApi, error) {
	res := &ImportApi{
		importSvc: importSvc,
	}

	mux.GetMux().Handle(
		importv1connect.NewImportServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}

func (i *ImportApi) ImportTransactions(
	ctx context.Context,
	c *connect.Request[importv1.ImportTransactionsRequest],
) (*connect.Response[importv1.ImportTransactionsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, auth.ErrInvalidToken)
	}

	resp, err := i.importSvc.Import(
		ctx,
		c.Msg,
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}
