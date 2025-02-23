package boilerplate

import "context"

type MethodExecutionData struct {
	JwtInfo  JwtClaims
	InnerJwt string `json:"-"`
	Context  context.Context
}

type executionKey struct{}

func WithContext(ctx context.Context, execution *MethodExecutionData) context.Context {
	return context.WithValue(ctx, executionKey{}, execution)
}

func FromContext(ctx context.Context) MethodExecutionData {
	val := ctx.Value(executionKey{})

	if val == nil {
		return MethodExecutionData{} // todo !?
	}

	return *val.(*MethodExecutionData)
}
