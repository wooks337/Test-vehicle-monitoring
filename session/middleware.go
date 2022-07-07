package session

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/boj/redistore.v1"
	"gorm.io/gorm"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	common "test-vehcile-monitoring/common/config"
	"test-vehcile-monitoring/common/logger"
	"test-vehcile-monitoring/common/utils"
	"test-vehcile-monitoring/monitoring/httpwares"
	"test-vehcile-monitoring/session/sessionstore"
	"time"
)

const (
	ServiceName         = "test"
	RequestLoggerName   = "request-logger"
	XDeviceID           = "X-DEVICE-ID"
	AuthorizationHeader = "Authorization"
	Name                = "x-maas-session"
	StateName           = "State"
	TokenName           = "Token"
	DataName            = "sessionData"
	UserID              = "UserID"
	IDToken             = "IDToken"
	OpenAPIName         = "X-Maas-Api-Key"
)

// Todo URL 정규식 체워 넣기
var req, _ = regexp.Compile(`(^/[\w\d_-]+)[/?]?`)

var unFobiddenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
})

type GatewayMiddleware struct {
	Logger       *logrus.Entry
	SessionStore *redistore.RediStore
	TokenStore   *sessionstore.ServiceSessionStore
	ServiceDB    *gorm.DB
	// OPA
	Config *common.Config
}

var unauthorizedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
})

var unBadRequstHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
})

var unknownErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"code":"500", "message":"middleware error"}`))
})

func (middleware *GatewayMiddleware) CacheProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 압축확장자 return bad req
		rex := regexp.MustCompile(`^(.*/)?(?:$|(.+?)(?:(\.[^.]*$)|$))`)
		extMatchStr := rex.FindStringSubmatch(r.URL.Path)
		extStr := strings.ToLower(extMatchStr[3])

		compressedFileExtensions := middleware.Config.RejectExtension.CompressedFileExtensions
		if len(compressedFileExtensions) > 0 && extStr != "" {
			for _, compressedFileExtension := range compressedFileExtensions {
				if compressedFileExtension == extStr ||
					(extStr == ".lzma" || extStr == ".gz" && strings.Contains(r.Header.Get("tar."), strings.ToLower(r.URL.Path))) {
					unBadRequstHandler.ServeHTTP(w, r)
					return
				}
			}
		}

		w.Header().Add("Cache-Control", "no-cache no-store")
		w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Add("Pragma", "no-cache")
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func (middleware *GatewayMiddleware) LogProxy(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, pmd, err := utils.InitRequest(r.Context())
		if err != nil {
			middleware.Logger.Errorf("LogProxy error[%v]", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		requestLogger := middleware.Logger.
			WithField(logger.LogMethod, fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI())).
			WithField(logger.LogTID, pmd.RequestID).
			WithField(logger.LogSID, "1").
			WithField(logger.LogTargetApp, ServiceName).
			WithField(logger.LogSourceApp, r.RemoteAddr)

		requestLogger.Infof("%s called", r.RequestURI)

		ctx = context.WithValue(
			ctx,
			RequestLoggerName,
			requestLogger,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (middleware *GatewayMiddleware) TraceLogProxy(next http.Handler) http.Handler {
	return httpwares.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/metrics":
			promhttp.Handler().ServeHTTP(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	}),
	// Todo middlewares setting
	//  htt_rqustid
	// http_logging
	// gorm
	// http_prometheus
	)
}

func (middleware *GatewayMiddleware) SessionProxy(h http.Handler) http.Handler {
	sessionHandler := func(w http.ResponseWriter, r *http.Request) {

		cookie, exists := r.Header["cookie"]
		if !exists {
			cookie, exists = r.Header["Cookie"]
			if !exists {
				unauthorizedHandler.ServeHTTP(w, r)
				return
			}
		}

		requestLogger := middleware.getLogger(r).WithField(logger.LogTargetApp, ServiceName)
		requestLogger.Info(fmt.Sprintf("Cookie=[%v]", cookie))

		session, sessionInfo, ok := middleware.getSessionInfo(w, r, requestLogger)
		if !ok {
			requestLogger.Info("sessionInfo Get Error: %v", ok)
			return
		}

		refererDomains := middleware.Config.Referer.RefererDomains
		if len(refererDomains) > 0 {
			cnt := 0
			for _, refererDomain := range refererDomains {
				refererUrl, _ := url.Parse(strings.ToLower(refererDomain))
				referHost, referPort, _ := net.SplitHostPort(refererUrl.Host)
				if referPort == "" {
					referHost = refererUrl.Host
				}
				if referHost != "" && strings.Contains(r.Header.Get("Referer"), referHost) {
					cnt++
				}
			}
			if cnt == 0 {
				unFobiddenHandler.ServeHTTP(w, r)
				w.Write([]byte("Referer invalid"))
				return
			}
			logrus.Info("Referer : ", r.Header.Get("Referer"))
		}

		if _, err := uuid.Parse(sessionInfo.UserID); err != nil {
			unauthorizedHandler.ServeHTTP(w, r)
			return
		}

		defer func() {
			if err := middleware.SessionStore.Save(r, w, session); err != nil {
				requestLogger.Errorf("fail to session refreshing [%v, %v]", session, err)
			} else {
				requestLogger.Infof("session refresh [%s]", session.ID)
			}
		}()

		ctx := r.Context()
		ctx = context.WithValue(ctx, DataName, sessionInfo)
		h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "session", session)))

		//requestUrl := r.RequestURI

		if strings.Contains(r.RequestURI, "?") {
			//runes := []rune(r.RequestURI)
			//requestUrl = string(runes[0:strings.Index(r.RequestURI, "?")])
		}

	}

	return http.HandlerFunc(sessionHandler)
}

func (middleware *GatewayMiddleware) getLogger(r *http.Request) *logrus.Entry {
	var requestLogger *logrus.Entry

	if value := r.Context().Value(RequestLoggerName); value != nil {
		requestLogger = value.(*logrus.Entry)
	} else {
		requestLogger = middleware.Logger
	}

	return requestLogger
}

func (middleware *GatewayMiddleware) getSessionInfo(w http.ResponseWriter, r *http.Request, requestLogger *logrus.Entry) (*sessions.Session, *sessionstore.ServiceSessionInfo, bool) {
	session, err := middleware.SessionStore.Get(r, Name)
	if err != nil {
		if strings.Index(err.Error(), "sercurecookie:") == 0 {
			requestLogger.Error(err.Error())
			unauthorizedHandler.ServeHTTP(w, r)
			return nil, nil, false
		} else {
			requestLogger.Error(err.Error())
			http.Error(w, fmt.Sprintf(`{"code":"500", "message":"%s"}`, fmt.Sprintf("error session-store[%v]", err)), http.StatusInternalServerError)
			return nil, nil, false
		}
	}

	if session == nil {
		requestLogger.Error("session is nil")
		unauthorizedHandler.ServeHTTP(w, r)
		return nil, nil, false
	}

	var token *oauth2.Token
	value, ok := session.Values[TokenName]
	if !ok {
		value, ok = session.Values["Oauth2Stage"]
		if ok && value == "authorized" {
			requestLogger.Info("old version cookies")
			expiry := session.Values["LoginTime"].(int64) + session.Values["ExpiredIn"].(int64)
			token = &oauth2.Token{
				AccessToken: session.Values["AccessToken"].(string),
				TokenType:   "Bearer",
				Expiry:      time.Unix(expiry, 0),
			}
		} else {
			requestLogger.Error("cannot find token values from session data")
			unauthorizedHandler.ServeHTTP(w, r)
			return nil, nil, false
		}

	} else {
		token = value.(*oauth2.Token)
	}

	userUID := session.Values[UserID].(string)
	sessionInfo, err := middleware.TokenStore.GetSessionInfo(userUID, token.AccessToken)
	if err != nil {
		requestLogger.Error(err.Error())
		unauthorizedHandler.ServeHTTP(w, r)
		return nil, nil, false
	} else if sessionInfo == nil {
		requestLogger.Error("cannot find token data from token-store")
		unauthorizedHandler.ServeHTTP(w, r)
		return nil, nil, false
	}

	return session, sessionInfo, true
}
