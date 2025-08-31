package main

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

func unaryLogging(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			start := time.Now()

			res, err := next(ctx, req)
			level := slog.LevelInfo
			if err != nil {
				level = slog.LevelError
			}

			logger.Log(ctx, level, "rpc",
				"procedure", req.Spec().Procedure,
				"lat_ms", time.Since(start).Milliseconds(),
			)

			return res, err
		}
	}
}
