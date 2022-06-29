package utils

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	}

	return http.StatusInternalServerError
}

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

const (
	// AuthorizationKey key for authorization metadata
	AuthorizationKey = contextKey("authorization")

	// RequestIDKey key for request_id metadata
	RequestIDKey = contextKey("request_id")

	// SpanIDKey key for span_id metadata
	SpanIDKey = contextKey("span_id")

	// SourceServiceKey key for source_service metadata
	SourceServiceKey = contextKey("source_service")
)

type SpanID struct {
	// 현재 context 에서 누적된 span id
	Prefix string

	// 현재 context 에서 외부 서비스 호출마다 증가시킬 시퀀스
	Seq int
}

// Increase 현재 context의 span id  시퀀스 증가
func (sid *SpanID) Increase() {
	sid.Seq++
}

// String span id를 문자열 변환
func (sid *SpanID) String() string {
	if sid.Seq == 0 {
		return sid.Prefix
	} else if sid.Prefix == "" {
		return fmt.Sprintf("%d", sid.Seq)
	}

	return fmt.Sprintf("%s.%d", sid.Prefix, sid.Seq)
}

// PropagateMetadata Metadata required to propagate from incoming to outgoing Context
type PropagateMetadata struct {
	Authorization string
	RequestID     string
	SourceService string
	SpanID        SpanID
}

// GetPropagateMetadataFromContext Get PropagateMetadata from Incoming Context
func GetPropagateMetadataFromContext(ctx context.Context) *PropagateMetadata {
	var (
		authorization = ""
		requestid     = ""
		spanid        = SpanID{}
	)

	md, ok := metadata.FromIncomingContext(ctx)

	// metadata empty
	if !ok {
		return &PropagateMetadata{
			Authorization: authorization,
			RequestID:     requestid,
			SpanID:        spanid,
		}
	}

	mdAuthorization := md.Get(AuthorizationKey.String())
	if mdAuthorization != nil && len(mdAuthorization) > 0 {
		authorization = mdAuthorization[0]
	}

	mdRequestID := md.Get(RequestIDKey.String())
	if mdRequestID != nil && len(mdRequestID) > 0 {
		requestid = mdRequestID[0]
	}

	mdSpanID := md.Get(SpanIDKey.String())
	if mdSpanID != nil && len(mdSpanID) > 0 {
		spanid = SpanID{mdSpanID[0], 0}
	}

	return &PropagateMetadata{
		Authorization: authorization,
		RequestID:     requestid,
		SpanID:        spanid,
	}
}

// UpdatePropagatedMetadataToOutgoingContext Update PropagatedMetadata to Outgoing Context
func UpdatePropagatedMetadataToOutgoingContext(ctx context.Context, pmd *PropagateMetadata) context.Context {
	if pmd == nil {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx,
		AuthorizationKey.String(), pmd.Authorization,
		RequestIDKey.String(), pmd.RequestID,
		SpanIDKey.String(), pmd.SpanID.String(),
	)

}

// InitRequest 요청 초기화 시 호촐되어 request ID가 없으면 발급, 메타데이터 추출 등을 수행
func InitRequest(ctx context.Context) (newCtx context.Context, pmd *PropagateMetadata, err error) {
	pmd = GetPropagateMetadataFromContext(ctx)

	if pmd.RequestID == "" {
		reqID, genErr := uuid.NewUUID()

		if genErr != nil {
			err = errors.Wrap(genErr, "Error while generating request ID")
		}

		pmd.RequestID = reqID.String()
	}

	requestID := pmd.RequestID
	header := metadata.Pairs(RequestIDKey.String(), requestID)
	grpc.SetHeader(ctx, header)

	newCtx = UpdatePropagatedMetadataToOutgoingContext(ctx, pmd)

	return newCtx, pmd, err
}
