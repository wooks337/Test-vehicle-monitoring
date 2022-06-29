package http_opentracing

import (
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
	"test-vehcile-monitoring/monitoring/ctxlogger"
	"test-vehcile-monitoring/monitoring/httpwares"
	"test-vehcile-monitoring/monitoring/reqid"
)

func Middleware(tracer opentracing.Tracer, operationFinder func(r *http.Request) string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := ctxlogger.FromContext(r.Context(), "htpp_opentracing.Middleware")

			// Retrieve a client span
			clientSpanCtx, err := tracer.Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.Header),
			)

			// Start a server span
			operation := operationFinder(r)
			var serverSpan opentracing.Span

			if err == opentracing.ErrSpanContextNotFound {
				log.Infoln("Start a root span.")
				serverSpan = tracer.StartSpan(operation)
				ext.SpanKind.Set(serverSpan, "server")
			} else {
				log.Infoln("Start a child span")
				serverSpan = tracer.StartSpan(operation, ext.RPCServerOption(clientSpanCtx))
			}
			defer serverSpan.Finish()

			ext.HTTPMethod.Set(serverSpan, r.Method)
			ext.HTTPUrl.Set(serverSpan, r.URL.String())

			// Associate a request ID with the span if the request ID is found.
			reqID, ok := reqid.FromContext(r.Context())
			if ok {
				log.Infoln("Found a request ID in the request context, %v", reqID.String())
				reqID.ToSpan(serverSpan)
			}

			ctx := opentracing.ContextWithSpan(r.Context(), serverSpan)
			req := r.WithContext(ctx)

			writer := httpwares.NewWrappedResponseWriter(w, false)

			next.ServeHTTP(writer, req)

			ext.HTTPStatusCode.Set(serverSpan, uint16(writer.StatusCode()))
		})
	}
}

func GorillaMuxOperationFinder(r *http.Request) string {
	if info := mux.CurrentRoute(r); info != nil {
		if path, err := info.GetHostTemplate(); err == nil {
			return r.Method + " " + path
		}
	}
	return "unknown"
}
