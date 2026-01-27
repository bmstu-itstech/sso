package middleware

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err)

		log = log.With(slog.String("method", info.FullMethod),
			slog.String("status", code.String()),
			slog.Duration("duration", time.Since(start)),
			slog.Time("time", start),
		)

		if err != nil {
			log = log.With(slog.String("error", err.Error()))
			log.Error("grpc request error")
		} else {
			log.Info("grpc request success")
		}

		return resp, err
	}
}
