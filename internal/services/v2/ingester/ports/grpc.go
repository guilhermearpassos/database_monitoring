package ports

import (
	"context"
	"errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/contract"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"google.golang.org/grpc"
	"io"
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

func (u UpIter) Next(ctx context.Context) (*UploadMessage, bool, error) {
	msg, err := u.in.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, false, nil
		}
	}
	return &UploadMessage{
		SnapshotID: msg.SnapshotId,
		Header: &SnapshotHeader{
			SnapshotID:           msg.GetSnapshotId(),
			TimestampUnix:        msg.GetHeader().GetTimestamp().GetSeconds(),
			ServerHost:           msg.GetHeader().GetServer().GetHost(),
			ServerType:           msg.GetHeader().GetServer().GetType(),
			ExpectedChunks:       msg.GetHeader().GetExpectedChunks(),
			ExpectedTotalSamples: msg.GetHeader().GetExpectedTotalSamples(),
			AgentVersion:         msg.GetHeader().GetAgentVersion(),
			Tags:                 msg.GetHeader().GetTags(),
			MaxChunkBytes:        msg.GetHeader().GetMaxChunkBytes(),
			Compression:          Compression(msg.GetHeader().GetCompression().Number()),
		},
		Chunk: &SampleChunk{
			ChunkSeq: msg.GetChunk().GetChunkSeq(),
			Samples:  msg.GetChunk().GetSamples(),
		},
		Finalize: &Finalize{TotalSamples: msg.GetFinalize().GetTotalSamples()},
	}, true, nil
}

var _ UploadIterator = (*UpIter)(nil)

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
	chunks, err := s.getMissingChunks(ctx, in.GetSnapshotId())
	if err != nil {
		return nil, err
	}
	return &ingestorv2.GetMissingChunksResponse{
		Status:        ingestorv2.UploadStatus(chunks.Status),
		MissingChunks: chunks.MissingChunks,
	}, nil
}

// IngestSnapshotStream adapts the ports iterator/DTOs to the app layer and maps the result back.
func (s *GrpcIngester) ingestSnapshotStream(ctx context.Context, it UploadIterator) (string, error) {
	appIter := &iterAdapter{it: it}
	snapId, err := s.app.Command.SaveSnapshot.Handle(ctx, appIter)
	if err != nil {
		return "", err
	}
	return snapId, nil
}

func (s *GrpcIngester) getMissingChunks(ctx context.Context, snapshotID string) (*GetMissingChunksResult, error) {
	st, miss, err := s.app.Query.GetMissingChunks.Handle(ctx, snapshotID)
	if err != nil {
		return nil, err
	}
	return &GetMissingChunksResult{Status: mapUploadStatus(st), MissingChunks: miss}, nil
}

// iterAdapter maps ports.UploadIterator -> app.command.UploadIterator.
type iterAdapter struct{ it UploadIterator }

func (a *iterAdapter) Next(ctx context.Context) (*contract.UploadMessage, bool, error) {
	m, ok, err := a.it.Next(ctx)
	if !ok || err != nil {
		return nil, ok, err
	}
	return &contract.UploadMessage{
		SnapshotID: m.SnapshotID,
		Header: func() *contract.SnapshotHeader {
			if m.Header == nil {
				return nil
			}
			return &contract.SnapshotHeader{
				SnapshotID:           m.Header.SnapshotID,
				TimestampUnix:        m.Header.TimestampUnix,
				ServerHost:           m.Header.ServerHost,
				ServerType:           m.Header.ServerType,
				ExpectedChunks:       m.Header.ExpectedChunks,
				ExpectedTotalSamples: m.Header.ExpectedTotalSamples,
				AgentVersion:         m.Header.AgentVersion,
				Tags:                 append([]string(nil), m.Header.Tags...),
				MaxChunkBytes:        m.Header.MaxChunkBytes,
			}
		}(),
		Chunk: func() *contract.SampleChunk {
			if m.Chunk == nil {
				return nil
			}
			return &contract.SampleChunk{ChunkSeq: m.Chunk.ChunkSeq, Samples: m.Chunk.Samples}
		}(),
		Finalize: func() *contract.Finalize {
			if m.Finalize == nil {
				return nil
			}
			return &contract.Finalize{TotalSamples: m.Finalize.TotalSamples}
		}(),
	}, true, nil
}

func mapUploadStatus(s domain.UploadStatus) UploadStatus {
	switch s {
	case domain.UploadStatusUnknown:
		return UploadStatusUnknown
	case domain.UploadStatusReceiving:
		return UploadStatusReceiving
	case domain.UploadStatusComplete:
		return UploadStatusComplete
	case domain.UploadStatusExpired:
		return UploadStatusExpired
	default:
		return UploadStatusUnknown
	}
}
