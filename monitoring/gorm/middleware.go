package gorm_middleware

import (
	"context"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net/http"
	gorm_logging "test-vehcile-monitoring/monitoring/gorm/logging"
)

var serviceDB *gorm.DB

func SetGlobalGorm(src *gorm.DB) {
	serviceDB = src
}

func Handler(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db = gorm_logging.WithLogger(r.Context(), db)

			next.ServeHTTP(w, r)
		})
	}
}

func PropagateGormInterceptor(ctx context.Context, method string, req, resp interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	serviceDB = gorm_logging.WithLogger(ctx, serviceDB)

	err = invoker(ctx, method, req, resp, cc, opts...)
	return
}
