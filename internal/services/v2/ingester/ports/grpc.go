package ports

import (
	"context"
	"errors"
	"io"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"google.golang.org/grpc"
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

func (u UpIter) Next(ctx context.Context) (*domain.UploadMessage, bool, error) {
	msg, err := u.in.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, false, nil
		}
	}
	if msg == nil {
		return nil, true, nil
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

	return &domain.UploadMessage{
		SnapshotID: msg.SnapshotId,
		Header:     header,
		Chunk:      chunk,
		Finalize:   finalize,
	}, true, nil
}

var _ domain.UploadIterator = (*UpIter)(nil)

func NewService(a app.Application) *GrpcIngester { return &GrpcIngester{app: a} }

func (s *GrpcIngester) IngestSnapshotStream(in grpc.ClientStreamingServer[ingestorv2.SnapshotUploadRequest, ingestorv2.SnapshotUploadResult]) error {

	_, err := s.ingestSnapshotStream(in.Context(), &UpIter{in: in})
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

// IngestSnapshotStream adapts the ports iterator/DTOs to the app layer and maps the result back.
func (s *GrpcIngester) ingestSnapshotStream(ctx context.Context, it domain.UploadIterator) (string, error) {
	snapId, err := s.app.Command.SaveSnapshot.Handle(ctx, it)
	if err != nil {
		return "", err
	}
	return snapId, nil
}
