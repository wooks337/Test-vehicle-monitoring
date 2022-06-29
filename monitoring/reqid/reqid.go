package reqid

import (
	"context"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"net/http"
	"test-vehcile-monitoring/monitoring/ctxlogger"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.DebugLevel)
}

type ctxRequestIDKey struct{}

const (
	mdRequestIDKey          = "x-request-id" // for metadata.MD for gRPC
	hdRequestIDKey          = "x-request-id" // for HTTP request header
	kafkaHeaderRequestIDKey = "x-request-id" // for Kafka message headers
	bgRequestIDKey          = "x-request-id"
	lgRequestIDKey          = "x-request-id"
)

// ReqID defines a request ID and helper functions.
type ReqID struct {
	id string
}

// New randomly gernerates a ReqID
func New() ReqID {
	return ReqID{
		id: uuid.New().String(),
	}
}

//  FromContext creates a ReqID with the request ID retrieved from ctx.
func FromContext(ctx context.Context) (ReqID, bool) {
	rid, ok := ctx.Value(ctxRequestIDKey{}).(ReqID)
	return rid, ok
}

// FromHTTP creates a ReqID with the request ID retrieved from h.
func FromHTTP(r *http.Request) (ReqID, bool) {
	var rid ReqID
	var ok bool

	id := r.Header.Get(hdRequestIDKey)
	if id != "" {
		rid.id = id
		ok = true
	}

	return rid, ok
}

// FromMetadata creates a ReqID with the request ID retrieved
// from the metadata parameter. If not found in md, it returns
// zero-valued ReqID and false.
func FromMetadata(md metadata.MD) (ReqID, bool) {
	var rid = ReqID{}

	rids, ok := md[mdRequestIDKey]
	if !ok || len(rids) == 0 {
		return rid, false
	}

	if rids[0] == "" {
		return rid, false
	}

	rid.id = rids[0]
	if len(rids) > 1 {
		log.Debugln("Found multiple request IDs in the metadata. Return the first one.")
	}

	return rid, ok
}

/*
	FromSapn a request ID sent by a client via the baggage
 	of the span. If not found, zero-valued ReqID and false are returned.
*/
func FromSpan(span opentracing.Span) (ReqID, bool) {
	var reqId ReqID
	var ok bool

	id := span.BaggageItem(bgRequestIDKey)
	if id != "" {
		reqId.id = id
		ok = true
	}

	return reqId, ok

}

/*
	FromSpanContext creates a ReqID with the request ID passed in
   the span's baggage which is injected into the RPC carrier on the client-side.
*/
func FromSpanContext(spanContext opentracing.SpanContext) (ReqID, bool) {
	var reqId ReqID
	var ok bool
	var id string

	spanContext.ForeachBaggageItem(func(k, v string) bool {
		if k == bgRequestIDKey {
			id = v
			return false
		}

		return true
	})

	if id != "" {
		reqId.id = id
		ok = true
	}

	return reqId, ok
}

// FromKafkaMessage creates a ReqID with the kafka headers
func FromKafkaMessage(message kafka.Message) (ReqID, bool) {
	var reqId = ReqID{}

	for _, header := range message.Headers {
		if header.Key == kafkaHeaderRequestIDKey {
			reqId.id = string(header.Value)
			return reqId, true
		}
	}

	return reqId, false
}

func (reqId ReqID) ToHTTP(r *http.Request) {
	r.Header.Set(hdRequestIDKey, reqId.String())
}

func (reqId ReqID) ToSpan(span opentracing.Span) {
	span.SetBaggageItem(bgRequestIDKey, reqId.String())
}

func (reqId ReqID) ToMetadataKV() (string, string) {
	return mdRequestIDKey, reqId.String()
}

// String stringifies the request ID.
func (reqId *ReqID) String() string {
	return reqId.id
}

// NewContext creates a context with rid.
func (reqId ReqID) NewContext(parent context.Context) context.Context {
	return context.WithValue(parent, ctxRequestIDKey{}, reqId)
}

// NewMetadata creates a metadata with reqId
func (reqId ReqID) NewMetadata() metadata.MD {
	return metadata.Pairs(mdRequestIDKey, reqId.String())
}

// AddToLoggingContext adds the request ID to the logging context in the given ctx.
func (reqId *ReqID) AddToLoggingContext(ctx context.Context) context.Context {
	c := ctxlogger.AddField(ctx, lgRequestIDKey, reqId.String())
	return c
}
