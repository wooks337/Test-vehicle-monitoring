package logging

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"test-vehcile-monitoring/monitoring/ctxlogger"
	"test-vehcile-monitoring/monitoring/httpwares"
	"time"
)

// Configuration define the middleware settings.
type Configuration struct {
	EnableRequestBody  bool
	EnableResponseBody bool
}

func Middleware(conf Configuration) httpwares.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			logger := ctxlogger.FromContext(r.Context(), "http_logging.Middleware")

			proto := r.Proto
			url := r.URL.String()
			method := r.Method

			l := logger.WithFields(logrus.Fields{
				"proto":  proto,
				"url":    url,
				"method": method,
				"type":   "HTTP",
			})

			if conf.EnableResponseBody && r.ContentLength > 0 {
				b, _ := ioutil.ReadAll(r.Body)
				l = l.WithField("RequestBody", string(b))

				r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}

			l.Infoln("Requested")

			writer := httpwares.NewWrappedResponseWriter(w, conf.EnableRequestBody)
			next.ServeHTTP(writer, r)

			// Log the response
			// nanoseconds to milliseconds
			elasped := time.Since(startTime).Milliseconds()

			l = logger.WithFields(logrus.Fields{
				"proto":   proto,
				"url":     url,
				"method":  method,
				"status":  writer.StatusCode(),
				"elasped": strconv.FormatInt(elasped, 10) + "ms",
			})

			hasBinaryDataOfImage := strings.Contains(r.URL.String(), "/image")
			if conf.EnableResponseBody && writer.MessageLength() > 0 {
				appendFields := logrus.Fields{
					"length": writer.MessageLength(),
				}

				if !hasBinaryDataOfImage {
					appendFields["ResponseBody"] = string(writer.Body().Bytes())
				}

				l = l.WithFields(appendFields)
			}

			l.Infoln("Responded")
		})
	}
}
