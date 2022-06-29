package request_id

import (
	"net/http"
	"test-vehcile-monitoring/monitoring/ctxlogger"
	"test-vehcile-monitoring/monitoring/httpwares"
	"test-vehcile-monitoring/monitoring/reqid"
)

func Middleware() httpwares.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := ctxlogger.FromContext(r.Context(), "http_requestid.middleware")

			reqID, ok := reqid.FromHTTP(r)
			if ok {
				log.Infof("Use the request ID, %s", reqID.String())
			} else {
				reqID = reqid.New()
				reqID.ToHTTP(r)
				log.Infof("Created a request ID, %s", reqID.String())
			}
			ctx := reqID.NewContext(r.Context()) // Add the request ID to the request context
			ctx = reqID.AddToLoggingContext(ctx) // Add the request ID to logging context

			req := r.WithContext(ctx)

			next.ServeHTTP(w, req)
		})
	}
}
