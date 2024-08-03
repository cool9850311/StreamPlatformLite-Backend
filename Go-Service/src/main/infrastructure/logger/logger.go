package logger

import (
	domainLogger "Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/util"
	"context"
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

type LoggerImpl struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

func NewLogger(logFile string) (domainLogger.Logger, error) {
	workingdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("%s", err)
	}
	log := logrus.New()

	// Open the log file
	file, err := os.OpenFile(util.TrimPathToBase(workingdir, "Go-Service/")+logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Create a MultiWriter to write logs to both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Set logrus to write to the MultiWriter
	log.SetOutput(multiWriter)

	return &LoggerImpl{
		logger: log,
		entry:  logrus.NewEntry(log),
	}, nil
}

func (l *LoggerImpl) withTraceID(ctx context.Context) *logrus.Entry {
	traceID, ok := ctx.Value("trace_id").(string)
	if !ok {
		return l.entry
	}
	return l.entry.WithField("trace_id", traceID)
}

func (l *LoggerImpl) Panic(ctx context.Context, msg string) {
	l.withTraceID(ctx).Panic(msg)
}

func (l *LoggerImpl) Fatal(ctx context.Context, msg string) {
	l.withTraceID(ctx).Fatal(msg)
}

func (l *LoggerImpl) Error(ctx context.Context, msg string) {
	l.withTraceID(ctx).Error(msg)
}

func (l *LoggerImpl) Warn(ctx context.Context, msg string) {
	l.withTraceID(ctx).Warn(msg)
}

func (l *LoggerImpl) Info(ctx context.Context, msg string) {
	l.withTraceID(ctx).Info(msg)
}

func (l *LoggerImpl) Debug(ctx context.Context, msg string) {
	l.withTraceID(ctx).Debug(msg)
}

func (l *LoggerImpl) Trace(ctx context.Context, msg string) {
	l.withTraceID(ctx).Trace(msg)
}
