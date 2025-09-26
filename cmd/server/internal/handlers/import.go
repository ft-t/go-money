package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/import/v1/importv1connect"
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type ImportApi struct {
	importSvc ImportSvc
	parser    Parser
}

func NewImportApi(
	mux *boilerplate.DefaultGrpcServer,
	importSvc ImportSvc,
	parser Parser,
) (*ImportApi, error) {
	res := &ImportApi{
		importSvc: importSvc,
		parser:    parser,
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
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
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

func (i *ImportApi) ParseTransactions(
	ctx context.Context,
	c *connect.Request[importv1.ParseTransactionsRequest],
) (*connect.Response[importv1.ParseTransactionsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := i.parser.Parse(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}
