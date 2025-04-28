package logger

import (
	"context"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

type contextKey string

const (
	RequestIdKey contextKey = "request_id_ctx"
	RequestId    string     = "request_id"
)

// WithRequestId Create a copy of context with requestid added
func WithRequestId(ctx context.Context, requestId types.UID) context.Context {
	return context.WithValue(ctx, RequestIdKey, Logger(ctx).WithFields(logrus.Fields{RequestId: requestId}))
}

// Logger Return a reference of logrus.Entry with request_id set field
func Logger(ctx context.Context) *logrus.Entry {
	if ctxLogger, ok := ctx.Value(RequestIdKey).(*logrus.Entry); ok {
		return ctxLogger
	}

	log := logrus.StandardLogger()
	logger := logrus.NewEntry(log)
	return logger
}

// AddValueToContextLogger adds new key-value in the existing logger present in context
func AddValueToContextLogger(ctx context.Context, key string, value interface{}) context.Context {
	log := Logger(ctx)
	return context.WithValue(ctx, RequestIdKey, log.WithField(key, value))
}

// Init initializes logrus
func Init() {
	log := logrus.StandardLogger()
	updateLog(log)
}

func updateLog(log *logrus.Logger) {
	log.Formatter = &logrus.JSONFormatter{}
	log.Out = os.Stdout
	log.SetLevel(getLevel())
}

func getLevel() logrus.Level {
	debugMode, _ := strconv.ParseBool(os.Getenv("DEBUG_MODE"))
	if debugMode {
		return logrus.DebugLevel
	} else {
		return logrus.InfoLevel
	}
}
