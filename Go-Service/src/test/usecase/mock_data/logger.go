package mock_data

import (
	"context"
)

type MockLogger struct{}

func (m *MockLogger) Panic(ctx context.Context, msg string) {}
func (m *MockLogger) Fatal(ctx context.Context, msg string) {}
func (m *MockLogger) Error(ctx context.Context, msg string) {}
func (m *MockLogger) Warn(ctx context.Context, msg string)  {}
func (m *MockLogger) Info(ctx context.Context, msg string)  {}
func (m *MockLogger) Debug(ctx context.Context, msg string) {}
func (m *MockLogger) Trace(ctx context.Context, msg string) {}
