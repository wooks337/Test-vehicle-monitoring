package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"
	"gopkg.in/boj/redistore.v1"
	"gorm.io/gorm"
	_ "gorm.io/gorm/logger"
	"net/http"
	"strconv"
	common "test-vehcile-monitoring/common/config"
	"test-vehcile-monitoring/message"
	"test-vehcile-monitoring/session/sessionstore"
)

const (
	ERROR_MESSAGE_EMPTY            = "EMPTY_%s"
	ERROR_MESSAGE_NOT_A_UUID       = "NOT_A_UUID_%s"
	ERROR_MESSAGE_INVALID_FORMAT   = "INVALID_FORMAT_%s_%s"
	ERROR_MESSAGE_INVALID_PARAMS   = "INVALID_PARAM_%s"
	ERROR_CANNOT_PARSE_FORM_PARAMS = "CANNOT_PARSE_FORM_PARAMS"
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

type BaseHandler struct {
	Logger              *logrus.Entry
	Config              *common.Config
	ServiceDB           *gorm.DB
	SessionStore        *redistore.RediStore
	ServiceSessionStore *sessionstore.ServiceSessionStore
}

func replyError(w http.ResponseWriter, logger *logrus.Entry, code int, err error) {
	logger.Errorf("[%d]%v", code, err)
	msg := &message.ErrorMessage{
		ErrorCode: code,
		Message:   err.Error(),
		Error:     err,
	}

	w.WriteHeader(msg.ErrorCode)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		logger.Error(err)
	}
}

func replyCodeError(w http.ResponseWriter, logger *logrus.Entry, code int, err string) {
	logger.Errorf("[%d]%v", code, err)
	msg := &message.ErrorMessage{
		ErrorCode: code,
		Message:   err,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		logger.Error(err)
	}
}

func getRequestLogger(r *http.Request, defaultLogger *logrus.Entry) *logrus.Entry {
	if value := r.Context().Value(RequestLoggerName); value != nil {
		return value.(*logrus.Entry)
	}

	return defaultLogger
}

/*
func GetContextLogger(r *http.Request) *logrus.Entry {
}*/

func getSessionInfo(r *http.Request) (*sessionstore.ServiceSessionInfo, error) {
	if value := r.Context().Value(DataName); value == nil {
		return nil, errors.New("EMPTY SESSION DATA")
	} else {
		sessionStore := value.(*sessionstore.ServiceSessionInfo)
		return sessionStore, nil
	}
}

func getUUIForMap(key, messageKey string, params map[string]string) (uuid.UUID, error) {
	if param, ok := params[key]; !ok {
		return uuid.Nil, errors.New(fmt.Sprintf(ERROR_MESSAGE_EMPTY, messageKey))
	} else if uuidValue, err := uuid.Parse(param); err != nil {
		return uuid.Nil, errors.New(fmt.Sprintf(ERROR_MESSAGE_NOT_A_UUID, messageKey))
	} else {
		return uuidValue, nil
	}
}

func getStringFromMap(key, messageKey string, params map[string]string) (string, error) {
	if param, ok := params[key]; !ok {
		return "", errors.New(fmt.Sprintf(ERROR_MESSAGE_EMPTY, messageKey))
	} else {
		return param, nil
	}
}

func getUnitFromMap(key, messageKey string, params map[string]string) (uint, error) {
	if strValue, err := getStringFromMap(key, messageKey, params); err == nil {
		rtValue, err := strconv.Atoi(strValue)
		return uint(rtValue), err
	} else {
		return 0, err
	}
}

type httpRequestData interface {
	Validate() error
}

func parseHttpRequestParameter(r *http.Request, requestData httpRequestData) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := schema.NewDecoder().Decode(requestData, r.Form); err != nil {
		return err
	}

	return requestData.Validate()
}

/*
func getUUIDFromRequestVar(r *http.Request) (string, error) {
	query := mux.Vars(r)
	uuidStr, ok := query["ID"]
	if !ok {
		return "", errors.New("id not found")
	}

}*/
