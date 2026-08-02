package ports

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/guilhermearpassos/database-monitoring/internal/common/util"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// GrpcIngester is the gRPC port facade that depends on the application layer.
// Transport adapters should use this to invoke use-cases without touching
// command/query directly.
type GrpcIngester struct {
	app app.Application
	ingestorv2.UnimplementedIngestionServiceServer
}

type UpIter struct {
	in grpc.ClientStreamingServer[ingestorv2.SnapshotUploadRequest, ingestorv2.SnapshotUploadResult]
}

func (u UpIter) Next(ctx context.Context) (*domain.SnapUploadMessage, error) {
	msg, err := u.in.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
	}
	if msg == nil {
		return nil, nil
	}
	header := &domain.SnapshotHeader{
		SnapshotID:           msg.GetSnapshotId(),
		TimestampUnix:        msg.GetHeader().GetTimestamp().GetSeconds(),
		ServerHost:           msg.GetHeader().GetServer().GetHost(),
		ServerType:           msg.GetHeader().GetServer().GetType(),
		ExpectedChunks:       msg.GetHeader().GetExpectedChunks(),
		ExpectedTotalSamples: msg.GetHeader().GetExpectedTotalSamples(),
		AgentVersion:         msg.GetHeader().GetAgentVersion(),
		Tags:                 msg.GetHeader().GetTags(),
		MaxChunkBytes:        msg.GetHeader().GetMaxChunkBytes(),
	}
	chunk := &domain.SampleChunk{
		ChunkSeq: msg.GetChunk().GetChunkSeq(),
		Samples:  msg.GetChunk().GetSamples(),
	}
	finalize := &domain.Finalize{TotalSamples: msg.GetFinalize().GetTotalSamples()}
	if msg.GetHeader() == nil {
		header = nil
	}
	if msg.GetChunk() == nil {
		chunk = nil
	}
	if msg.GetFinalize() == nil {
		finalize = nil
	}

	return &domain.SnapUploadMessage{
		SnapshotID: msg.SnapshotId,
		Header:     header,
		Chunk:      chunk,
		Finalize:   finalize,
	}, nil
}

type PlanUpIter struct {
	in grpc.ClientStreamingServer[ingestorv2.ExecutionPlanUploadRequest, ingestorv2.ExecutionPlanUploadResult]
}

func (u PlanUpIter) Next(ctx context.Context) (*domain.PlanUploadMessage, error) {
	msg, err := u.in.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
	}
	if msg == nil {
		return nil, nil
	}
	header := &domain.PlanHeader{
		TimestampUnix:  msg.GetHeader().GetTimestamp().GetSeconds(),
		ServerHost:     msg.GetHeader().GetServer().GetHost(),
		ServerType:     msg.GetHeader().GetServer().GetType(),
		ExpectedChunks: msg.GetHeader().GetExpectedChunks(),
		AgentVersion:   msg.GetHeader().GetAgentVersion(),
		MaxChunkBytes:  msg.GetHeader().GetMaxChunkBytes(),
	}
	domainPlans := make([]*common_domain.ExecutionPlan, len(msg.GetChunk().GetPlans()))
	for i, chunk := range msg.GetChunk().GetPlans() {
		plan, err := converters.ExecutionPlanToDomain(chunk)
		if err != nil {
			return nil, fmt.Errorf("converting plan %d: %w", i, err)
		}
		domainPlans[i] = plan
	}
	chunk := &domain.PlanChunk{
		ChunkSeq:       msg.GetChunk().GetChunkSeq(),
		ExecutionPlans: domainPlans,
	}
	finalize := &domain.Finalize{TotalSamples: msg.GetFinalize().GetTotalSamples()}
	if msg.GetHeader() == nil {
		header = nil
	}
	if msg.GetChunk() == nil {
		chunk = nil
	}
	if msg.GetFinalize() == nil {
		finalize = nil
	}

	return &domain.PlanUploadMessage{
		Header:   header,
		Chunk:    chunk,
		Finalize: finalize,
	}, nil
}

var _ domain.UploadIterator[domain.SnapUploadMessage] = (*UpIter)(nil)

func NewService(a app.Application) *GrpcIngester { return &GrpcIngester{app: a} }

func (s *GrpcIngester) IngestSnapshotStream(in grpc.ClientStreamingServer[ingestorv2.SnapshotUploadRequest, ingestorv2.SnapshotUploadResult]) error {
	_, err := s.app.Command.SaveSnapshot.Handle(in.Context(), &UpIter{in: in})
	var msg string
	status := ingestorv2.SnapshotUploadResult_OK
	if err != nil {
		if errors.As(err, &domain.ErrMissingChunks) {
			status = ingestorv2.SnapshotUploadResult_PARTIAL
		}
	}
	err = in.SendAndClose(&ingestorv2.SnapshotUploadResult{
		Status:  status,
		Message: msg,
	})
	return err
}

func (s *GrpcIngester) GetMissingChunks(ctx context.Context, in *ingestorv2.GetMissingChunksRequest) (*ingestorv2.GetMissingChunksResponse, error) {
	st, miss, err := s.app.Query.GetMissingChunks.Handle(ctx, in.GetSnapshotId())
	if err != nil {
		return nil, err
	}
	var protoStatus ingestorv2.UploadStatus
	switch st {
	case domain.UploadStatusUnknown:
		protoStatus = ingestorv2.UploadStatus_UNKNOWN
	case domain.UploadStatusReceiving:
		protoStatus = ingestorv2.UploadStatus_RECEIVING
	case domain.UploadStatusComplete:
		protoStatus = ingestorv2.UploadStatus_COMPLETE
	case domain.UploadStatusExpired:
		// proto lacks EXPIRED; map to UNKNOWN or COMPLETE as appropriate. We'll use UNKNOWN.
		protoStatus = ingestorv2.UploadStatus_UNKNOWN
	default:
		protoStatus = ingestorv2.UploadStatus_UNKNOWN
	}
	return &ingestorv2.GetMissingChunksResponse{Status: protoStatus, MissingChunks: miss}, nil
}

func (s *GrpcIngester) IngestExecutionPlanStream(in grpc.ClientStreamingServer[ingestorv2.ExecutionPlanUploadRequest, ingestorv2.ExecutionPlanUploadResult]) error {
	err := s.app.Command.SaveExecutionPlans.Handle(in.Context(), &PlanUpIter{in: in})
	if err != nil {
		return err
	}
	err = in.SendAndClose(&ingestorv2.ExecutionPlanUploadResult{})
	return err
}

func (s *GrpcIngester) GetMissingPlans(in *ingestorv2.GetMissingPlansRequest, stream grpc.ServerStreamingServer[ingestorv2.GetMissingPlansResponse]) error {
	if err := in.GetFrom().CheckValid(); err != nil {
		return fmt.Errorf("invalid from timestamp %v: %w", in.GetFrom(), err)
	}
	if err := in.GetTo().CheckValid(); err != nil {
		return fmt.Errorf("invalid to timestamp %v: %w", in.GetTo(), err)
	}
	handles, err := s.app.Query.GetMissingPlans.Handle(stream.Context(), common_domain.ServerMeta{
		Host: in.Server.Host,
		Type: in.Server.Type,
	}, in.GetFrom().AsTime(), in.GetTo().AsTime())
	if err != nil {
		return fmt.Errorf("getting missing plans: %w", err)
	}

	chunks := util.PackByMaxBytes(handles, 100000, func(cur []string, next string) int {
		return proto.Size(&ingestorv2.PlanHandleChunk{PlanHandles: append(cur, next)})
	})
	for _, chunk := range chunks {
		if err := stream.Send(&ingestorv2.GetMissingPlansResponse{
			Payload: &ingestorv2.GetMissingPlansResponse_Chunk{
				Chunk: &ingestorv2.PlanHandleChunk{
					PlanHandles: chunk,
				},
			},
		}); err != nil {
			return fmt.Errorf("sending plan handles: %w", err)
		}
	}
	err = stream.Send(&ingestorv2.GetMissingPlansResponse{
		Payload: &ingestorv2.GetMissingPlansResponse_Finalize{Finalize: &ingestorv2.Finalize{TotalSamples: uint64(len(handles))}},
	})
	if err != nil {
		return fmt.Errorf("sending finalize: %w", err)
	}

	return nil
}
