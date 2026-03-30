package interceptor

import (
	"context"

	"connectrpc.com/connect"
)

type contextKey string

const ClientIPKey contextKey = "client_ip"

func NewClientIPInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			ctx = context.WithValue(ctx, ClientIPKey, req.Header().Get("X-Real-IP"))
			return next(ctx, req)
		}
	}
}
