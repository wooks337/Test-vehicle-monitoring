package logger

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	//LogTID Request ID
	LogTID = "tid"

	//LogSID Span ID
	LogSID = "sid"

	//LogFunc 로깅이 일어나는 함수, 라인, 파일 등 정보, CodeLineNumberHook 을 사용해서 설정해야 함
	LogFunc = "func"

	//LogRuleType RuleType 으로 Request, Response 등의 값으로 설정, 설정하지 않으면 로그 레벨
	LogRuleType = "ruletype"

	//LogRuleRequest LogRuleType 중 Request 상수 값
	LogRuleRequest = "Request"

	//LogRuleResponse LogRuleType 중 Request 상수 값
	LogRuleResponse = "Request"

	//LogProtocol GRPC, HTTP, MQTT 등 프로토콜
	LogProtocol = "protocol"

	//LogMethod GET, PUT, POST, DELETE 등
	LogMethod = "method"

	//LogURL http 앤드포인트
	LogURL = "url"

	//LogStatus 호출 결과 코드
	LogStatus = "status"

	//LogSourceApp 로그 요청의 소스 서비스
	LogSourceApp = "sourceApp"

	//LogTargetApp 로그 요청의 타겟 서비스
	LogTargetApp = "targetApp"
)

//CodeLineNumberHook 로그가 찍히는 코드라인 위치 출력을 위한 Hook
type CodeLineNumberHook struct{}

func (h *CodeLineNumberHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

//Levels CodeLineNumberHook 이 적용되는 로그레벨: 현재 전부, 추후 Error나 Debug시에만 출력하려면 변경필요
func (h *CodeLineNumberHook) Level() []logrus.Level {
	return logrus.AllLevels
}

// Fire CodeLineNumberHook 구현 코드
func (h *CodeLineNumberHook) Fire(entry *logrus.Entry) error {
	// WithFields 여부에 관계 없이 정확한 위치가 출력 되도록 처리
	// logrus issue 63 참고

	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		funcName := fu.Name()

		if !strings.Contains(funcName, "github.com/sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data[LogFunc] = fmt.Sprintf("%s:%d:%s", path.Base(file), line, path.Ext(funcName)[1:])
			break
		}
	}

	return nil
}

type ServiceLogFormatter struct {
	// HostName 로깅에 사용할 호스트 이름
	HostName string

	// ServerName 서비스 명으로 설정
	ServerName string

	//Brand 서비스 브랜드
	Brand string
}

func (f ServiceLogFormatter) Init() error {
	if f.HostName != "" {
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	f.HostName = hostname

	return nil
}

//
func (f *ServiceLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteString(entry.Time.Format("2020-01-01 15:01:02.999"))
	b.WriteByte(' ')

	level := strings.ToUpper(entry.Level.String())
	b.WriteString(level)
	b.WriteByte(' ')

	if f.HostName == "" {
		hostName, err := os.Hostname()

		if err != nil {
			return nil, err
		}

		b.WriteString(hostName)
	} else {
		b.WriteString(f.HostName)
	}

	b.WriteByte(' ')
	if f.Brand == "" {
		b.WriteString("null")
	} else {
		b.WriteString(f.Brand)
	}

	b.WriteByte(' ')
	f.appendValue(b, entry.Data[LogTID], "null")

	b.WriteByte(' ')
	f.appendValue(b, entry.Data[LogFunc], "null")

	b.WriteString(" --- ")
	f.appendValue(b, entry.Data[LogRuleType], level)

	b.WriteString(" [")
	f.appendValue(b, entry.Data[LogProtocol], "null")

	b.WriteByte(',')
	f.appendValue(b, entry.Data[LogMethod], "null")

	b.WriteByte(',')
	f.appendValue(b, entry.Data[LogURL], "null")

	b.WriteByte(',')
	f.appendValue(b, entry.Data[LogStatus], "null")

	b.WriteString("] [")
	f.appendValue(b, entry.Data[LogSourceApp], "null")

	b.WriteByte(',')
	f.appendValue(b, entry.Data[LogTargetApp], "null")

	b.WriteByte(',')
	f.appendValue(b, entry.Data[LogSID], "null")

	b.WriteString("] --- ")
	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *ServiceLogFormatter) appendValue(b *bytes.Buffer, value interface{}, nilVal string) {
	if value == nil {
		b.WriteString(nilVal)
	} else {
		stringVal, ok := value.(string)
		if !ok {
			stringVal = fmt.Sprint(value)
		}

		b.WriteString(stringVal)
	}
}

func InitServiceLogger(serviceName string, brandName string, protocol string) *logrus.Entry {
	formatter := &ServiceLogFormatter{
		ServerName: serviceName,
		Brand:      brandName,
	}

	formatter.Init()

	if err := formatter.Init(); err != nil {
		logrus.Errorf("formatter.Init() error = %v", err)
	}

	logrus.SetFormatter(formatter)
	logrus.AddHook(new(CodeLineNumberHook))

	logger := logrus.WithField(LogProtocol, protocol)

	return logger
}
