package handler

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"test-vehcile-monitoring/session/sessionstore"
	"testing"
)

func Test_getRequestLogger(t *testing.T) {
	type args struct {
		r             *http.Request
		defaultLogger *logrus.Entry
	}
	tests := []struct {
		name string
		args args
		want *logrus.Entry
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRequestLogger(tt.args.r, tt.args.defaultLogger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRequestLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSessionInfo(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *sessionstore.ServiceSessionInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSessionInfo(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSessionInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSessionInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStringFromMap(t *testing.T) {
	type args struct {
		key        string
		messageKey string
		params     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStringFromMap(tt.args.key, tt.args.messageKey, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("getStringFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getStringFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUUIForMap(t *testing.T) {
	type args struct {
		key        string
		messageKey string
		params     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUUIForMap(tt.args.key, tt.args.messageKey, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUUIForMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUUIForMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUnitFromMap(t *testing.T) {
	type args struct {
		key        string
		messageKey string
		params     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUnitFromMap(tt.args.key, tt.args.messageKey, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUnitFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getUnitFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseHttpRequestParameter(t *testing.T) {
	type args struct {
		r           *http.Request
		requestData httpRequestData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseHttpRequestParameter(tt.args.r, tt.args.requestData); (err != nil) != tt.wantErr {
				t.Errorf("parseHttpRequestParameter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_replyCodeError(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		logger *logrus.Entry
		code   int
		err    string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replyCodeError(tt.args.w, tt.args.logger, tt.args.code, tt.args.err)
		})
	}
}

func Test_replyError(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		logger *logrus.Entry
		code   int
		err    error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replyError(tt.args.w, tt.args.logger, tt.args.code, tt.args.err)
		})
	}
}
