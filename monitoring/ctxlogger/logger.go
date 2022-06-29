package ctxlogger

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
)

type ctxLoggerKy struct{}

type ctxLogger struct {
	logger *logrus.Entry // context logger
	fields logrus.Fields // added files
}

var defaultCtxLogger *logrus.Entry

const LogLevel = logrus.DebugLevel

func Init(level logrus.Level) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyTime:  "@timestamp",
		},
	})

	defaultCtxLogger = logrus.NewEntry(logger)
}

func FromContext(ctx context.Context, src string) *logrus.Entry {
	var logger *logrus.Entry

	l, ok := ctx.Value(ctxLoggerKy{}).(*ctxLogger)

	if ok {
		logger = l.logger.WithFields(l.fields)
	} else {
		logger = defaultCtxLogger
	}

	if src != "" {
		logger = logger.WithField("src", src)
	}

	return logger
}

func AddField(ctx context.Context, key string, value interface{}) context.Context {
	l, ok := ctx.Value(ctxLoggerKy{}).(*ctxLogger)
	if !ok {
		l = &ctxLogger{
			logger: defaultCtxLogger,
			fields: logrus.Fields{},
		}
	}

	l.fields[key] = value
	return context.WithValue(ctx, ctxLogger{}, l)
}

func AddFields(ctx context.Context, files logrus.Fields) context.Context {
	l, ok := ctx.Value(ctxLoggerKy{}).(*ctxLogger)
	if !ok {
		l = &ctxLogger{
			logger: defaultCtxLogger,
			fields: logrus.Fields{},
		}
	}

	for k, v := range files {
		l.fields[k] = v
	}

	return context.WithValue(ctx, ctxLoggerKy{}, l)
}
