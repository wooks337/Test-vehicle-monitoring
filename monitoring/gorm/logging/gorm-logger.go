package gorm_logging

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"test-vehcile-monitoring/monitoring/ctxlogger"
	"test-vehcile-monitoring/monitoring/reqid"
)

type GormLogger struct {
	log *logrus.Entry
}

func (l GormLogger) Print(args ...interface{}) {
	if len(args) == 0 {
		l.log.Println(args...)
		return
	}

	switch args[0] {
	case "log", "error":
		// len : 3
		// 0 : "log" , 1 : string(code location), 2 : errorString
		if len(args) == 3 {
			l.log.Errorf("%v %v", args[2], args[1])
		} else {
			l.log.Error(args...)
		}

	case "sql":
		// len 6
		// 0  :"sql" , 1 : string(code location), 2 : time.Duration, 3:string(SQL statement),  4:params, 5: int64(affected rows)
		if len(args) == 6 {
			l.log.Debugf("%s %v %s", args[3], args[4], args[1])
			l.log.Infof("elapsed %v affected on %d rows", args[2], args[5])
		} else {
			l.log.Info(args...)
		}
	default:
		l.log.Print(args...)
	}
}

func WithLogger(ctx context.Context, db *gorm.DB) *gorm.DB {
	reqID, ok := reqid.FromContext(ctx)
	if !ok {
		reqID = reqid.New()
		ctx = reqID.AddToLoggingContext(ctx)
	}

	gormLogger := GormLogger{
		log: ctxlogger.FromContext(ctx, "gorm").
			WithField("type", "gorm").
			WithField("x-request-id", reqID.String()),
	}

	logrus.Debug(gormLogger)
	return db
}
