package tracer

import (
	"github.com/opentracing/opentracing-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"io"
)

/*
 분산 시스템 환경에서 트랜젝션 추적용으로 사용
 Jaeger 를 사용하세 구션 하기
*/

type Configuration struct {
	ServiceName string // Service Name
	HostPort    string // Host and port using in Jaeger transport

	SamplerType  string  // sampler type [const, remote, probabilistic]
	SamplerParam float64 // sample parameter [0~1]
}

func Init(env Configuration) (io.Closer, error) {
	jaegerEnv := jaegerconfig.Configuration{
		ServiceName: env.ServiceName,
		Disabled:    false,
		RPCMetrics:  false,
		Tags:        nil,
		Sampler: &jaegerconfig.SamplerConfig{
			Type:  env.SamplerType,
			Param: 1,
		},
		Reporter: &jaegerconfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: env.HostPort,
		},
		Headers:             nil,
		BaggageRestrictions: nil,
		Throttler:           nil,
	}

	if tracer, closer, err := jaegerEnv.NewTracer(); err != nil {
		return nil, err
	} else {
		opentracing.SetGlobalTracer(tracer)
		return closer, nil
	}
}
