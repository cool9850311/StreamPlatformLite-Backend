package logger

import "context"

type Logger interface {
	Panic(ctx context.Context, msg string)
	Fatal(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)
	Debug(ctx context.Context, msg string)
	Trace(ctx context.Context, msg string)
}
