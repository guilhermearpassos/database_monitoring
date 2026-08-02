package ingestor

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/common/util"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	domainingestor "github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/ingestor"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client is a thin wrapper around the v2 IngestionService that knows how to
// stream a DBSnapshot using the agreed header -> chunks -> finalize protocol.
// It works with the agent's internal domain models and converts them to protos.
//
// Typical usage:
//   c, _ := ingestor.New(ingestor.Options{Address: "127.0.0.1:50051"})
//   defer c.Close()
//   _ , err := c.SendSnapshot(ctx, snapshot, SendOptions{})
//   ...

type Client struct {
	client ingestorv2.IngestionServiceClient
}

// Ensure this adapter implements the domain interface.
var _ domainingestor.Client = (*Client)(nil)

// SendOptions is the domain contract; reuse it here via type alias.
type SendOptions = domainingestor.SendOptions

func New(client ingestorv2.IngestionServiceClient) (*Client, error) {
	return &Client{client: client}, nil
}

// SendSnapshot streams the given snapshot to the ingester. It attempts a simple
// send; if the final RPC returns an error, it will query GetMissingChunks and
// try a best-effort resume once.
func (c *Client) SendSnapshot(ctx context.Context, snap *common_domain.DataBaseSnapshot, so SendOptions) (*ingestorv2.SnapshotUploadResult, error) {
	if snap == nil {
		return nil, fmt.Errorf("nil snapshot")
	}
	// default to ~1 MiB if not provided, with a small safety margin
	if so.MaxUncompressedBytes <= 0 {
		so.MaxUncompressedBytes = 900 * 1024
	}

	// First attempt: plain streaming upload
	res, err := c.streamUploadSamples(ctx, snap, so, nil)
	if err == nil {
		return res, nil
	}

	// On error, try to resume once using GetMissingChunks if the server knows the snapshot
	status, missing, qErr := c.queryMissing(ctx, snap.SnapInfo.ID)
	if qErr != nil {
		return nil, fmt.Errorf("upload failed and resume query failed: %v (resume query err=%v)", err, qErr)
	}
	if status == ingestorv2.UploadStatus_UNKNOWN {
		// Server has no state — try once more from scratch
		res2, err2 := c.streamUploadSamples(ctx, snap, so, nil)
		if err2 != nil {
			return nil, err2
		}
		return res2, nil
	}
	// Server has partial data; resend only missing sequences
	res2, err2 := c.streamUploadSamples(ctx, snap, so, missing)
	if err2 != nil {
		return nil, err2
	}
	return res2, nil
}

func (c *Client) streamUploadSamples(ctx context.Context, snap *common_domain.DataBaseSnapshot, so SendOptions, resendSeqs []uint32) (*ingestorv2.SnapshotUploadResult, error) {
	stream, err := c.client.IngestSnapshotStream(withAgentMeta(ctx, so.AgentVersion))
	if err != nil {
		return nil, err
	}

	// Pre-build proto samples and pack into chunks under MaxUncompressedBytes
	protoSamples := make([]*dbmv1.QuerySample, 0, len(snap.Samples))
	for _, s := range snap.Samples {
		protoSamples = append(protoSamples, toProtoSample(s))
	}
	// Use a reusable probe to compute vtproto size when tentatively appending the next item
	probe := &ingestorv2.SampleChunk{ChunkSeq: 1}
	chunks := util.PackByMaxBytes(protoSamples, so.MaxUncompressedBytes, func(cur []*dbmv1.QuerySample, next *dbmv1.QuerySample) int {
		probe.Samples = append(cur, next)
		return probe.SizeVT()
	})

	// Header
	header := &ingestorv2.SnapshotHeader{
		Timestamp:            timestamppb.New(snap.SnapInfo.Timestamp),
		Server:               toProtoServer(snap.SnapInfo.Server),
		ExpectedChunks:       uint32(len(chunks)),
		ExpectedTotalSamples: uint64(len(snap.Samples)),
		AgentVersion:         so.AgentVersion,
		Tags:                 so.Tags,
		Compression:          so.Compression,
		MaxChunkBytes:        uint32(so.MaxUncompressedBytes),
	}
	if err := stream.Send(&ingestorv2.SnapshotUploadRequest{
		SnapshotId: snap.SnapInfo.ID,
		Payload:    &ingestorv2.SnapshotUploadRequest_Header{Header: header},
	}); err != nil {
		_ = stream.CloseSend()
		return nil, fmt.Errorf("send header: %w", err)
	}

	// Determine which chunks to send
	var seqSet map[uint32]struct{}
	if len(resendSeqs) > 0 {
		seqSet = make(map[uint32]struct{}, len(resendSeqs))
		for _, s := range resendSeqs {
			seqSet[s] = struct{}{}
		}
	}

	// Stream chunks
	for i, chunkSamples := range chunks {
		seq := uint32(i + 1)
		if seqSet != nil {
			if _, ok := seqSet[seq]; !ok {
				continue
			}
		}
		msg := &ingestorv2.SnapshotUploadRequest{
			SnapshotId: snap.SnapInfo.ID,
			Payload: &ingestorv2.SnapshotUploadRequest_Chunk{
				Chunk: &ingestorv2.SampleChunk{
					ChunkSeq: seq, Samples: chunkSamples,
				},
			},
		}
		if err := stream.Send(msg); err != nil {
			_ = stream.CloseSend()
			return nil, fmt.Errorf("send chunk %d: %w", seq, err)
		}
	}

	// Finalize
	fin := &ingestorv2.SnapshotUploadRequest{
		SnapshotId: snap.SnapInfo.ID,
		Payload: &ingestorv2.SnapshotUploadRequest_Finalize{
			Finalize: &ingestorv2.Finalize{TotalSamples: uint64(len(snap.Samples))},
		},
	}
	if err := stream.Send(fin); err != nil {
		_ = stream.CloseSend()
		return nil, fmt.Errorf("send finalize: %w", err)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) queryMissing(ctx context.Context, snapshotID string) (ingestorv2.UploadStatus, []uint32, error) {
	resp, err := c.client.GetMissingChunks(ctx, &ingestorv2.GetMissingChunksRequest{SnapshotId: snapshotID})
	if err != nil {
		return ingestorv2.UploadStatus_UNKNOWN, nil, err
	}
	return resp.Status, resp.MissingChunks, nil
}

func toProtoServer(s common_domain.ServerMeta) *dbmv1.ServerMetadata {
	return &dbmv1.ServerMetadata{Host: s.Host, Type: s.Type}
}

func toProtoSample(s *common_domain.QuerySample) *dbmv1.QuerySample {
	if s == nil {
		return nil
	}
	qs := &dbmv1.QuerySample{
		Status:            s.Status,
		SqlHandle:         s.SqlHandle,
		Text:              s.Text,
		Blocked:           s.IsBlocked,
		Blocker:           s.IsBlocker,
		TimeElapsedMillis: s.TimeElapsedMs,
		Session: &dbmv1.SessionMetadata{
			SessionId:        s.Session.SessionID,
			LoginTime:        timestamppb.New(s.Session.LoginTime),
			Host:             s.Session.HostName,
			ProgramName:      s.Session.ProgramName,
			LoginName:        s.Session.LoginName,
			Status:           s.Session.Status,
			LastRequestStart: timestamppb.New(s.Session.LastRequestStartTime),
			LastRequestEnd:   timestamppb.New(s.Session.LastRequestEndTime),
			ConnectionId:     s.Session.ConnectionId,
			ClientIp:         s.Session.ClientIP,
		},
		Db:         &dbmv1.DBMetadata{DatabaseId: s.Database.DatabaseID, DatabaseName: s.Database.DatabaseName},
		BlockInfo:  &dbmv1.BlockMetadata{BlockedBy: s.Block.BlockedBy, BlockedSessions: append([]string(nil), s.Block.BlockedSessions...)},
		WaitInfo:   &dbmv1.WaitMetadata{WaitType: valOrEmpty(s.Wait.WaitType), WaitTime: int64(s.Wait.WaitTime), LastWaitType: s.Wait.LastWaitType, WaitResource: s.Wait.WaitResource},
		SnapInfo:   &dbmv1.SnapMetadata{Id: s.Snapshot.ID, Timestamp: timestamppb.New(s.Snapshot.Timestamp)},
		PlanHandle: s.PlanHandle,
		Id:         s.Id,
		Command:    &dbmv1.CommandMetadata{TransactionId: s.CommandMetadata.TransactionId, RequestId: s.CommandMetadata.RequestId, EstimatedCompletionTime: s.CommandMetadata.EstimatedCompletionTime, PercentComplete: s.CommandMetadata.PercentComplete},
		QueryHash:  s.QueryHash,
	}
	return qs
}

func valOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// withAgentMeta attaches simple agent-version metadata to the stream for logs/tracing.
func withAgentMeta(ctx context.Context, agentVersion string) context.Context {
	if agentVersion == "" {
		return ctx
	}
	md := metadata.Pairs("x-agent-version", agentVersion)
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) SendExecutionPlans(ctx context.Context, plans []*common_domain.ExecutionPlan, server common_domain.ServerMeta, so SendOptions) (*ingestorv2.ExecutionPlanUploadResult, error) {

	if plans == nil {
		return &ingestorv2.ExecutionPlanUploadResult{}, nil
	}
	// default to ~1 MiB if not provided, with a small safety margin
	if so.MaxUncompressedBytes <= 0 {
		so.MaxUncompressedBytes = 900 * 1024
	}

	// First attempt: plain streaming upload
	res, err := c.streamUploadPlans(ctx, plans, server, so)
	if err != nil {
		return nil, fmt.Errorf("send execution plans: %w", err)
	}

	return res, nil
}

func (c *Client) streamUploadPlans(ctx context.Context, plans []*common_domain.ExecutionPlan,
	server common_domain.ServerMeta, so SendOptions) (*ingestorv2.ExecutionPlanUploadResult, error) {
	stream, err := c.client.IngestExecutionPlanStream(withAgentMeta(ctx, so.AgentVersion))
	if err != nil {
		return nil, err
	}

	// Pre-build proto samples and pack into chunks under MaxUncompressedBytes
	protoSamples := make([]*dbmv1.ExecutionPlan, 0, len(plans))
	for _, s := range plans {
		proto, err := converters.ExecutionPlanToProto(s)
		if err != nil {
			return nil, fmt.Errorf("convert execution plan to proto: %w", err)
		}
		protoSamples = append(protoSamples, proto)
	}
	// Use a reusable probe to compute vtproto size when tentatively appending the next item
	probe := &ingestorv2.ExecPlanChunk{ChunkSeq: 1}
	chunks := util.PackByMaxBytes(protoSamples, so.MaxUncompressedBytes, func(cur []*dbmv1.ExecutionPlan, next *dbmv1.ExecutionPlan) int {
		probe.Plans = append(cur, next)
		return probe.SizeVT()
	})

	// Header
	header := &ingestorv2.PlanHeader{
		Timestamp:      timestamppb.New(time.Now()),
		Server:         toProtoServer(server),
		ExpectedChunks: uint32(len(chunks)),
		AgentVersion:   so.AgentVersion,
		Compression:    so.Compression,
		MaxChunkBytes:  uint32(so.MaxUncompressedBytes),
	}
	if err := stream.Send(&ingestorv2.ExecutionPlanUploadRequest{
		Payload: &ingestorv2.ExecutionPlanUploadRequest_Header{Header: header},
	}); err != nil {
		_ = stream.CloseSend()
		return nil, fmt.Errorf("send header: %w", err)
	}

	// Stream chunks
	for i, chunkSamples := range chunks {
		seq := uint32(i + 1)
		msg := &ingestorv2.ExecutionPlanUploadRequest{
			Payload: &ingestorv2.ExecutionPlanUploadRequest_Chunk{
				Chunk: &ingestorv2.ExecPlanChunk{
					ChunkSeq: seq, Plans: chunkSamples,
				},
			},
		}
		if err := stream.Send(msg); err != nil {
			_ = stream.CloseSend()
			return nil, fmt.Errorf("send chunk %d: %w", seq, err)
		}
	}

	// Finalize
	fin := &ingestorv2.ExecutionPlanUploadRequest{
		Payload: &ingestorv2.ExecutionPlanUploadRequest_Finalize{
			Finalize: &ingestorv2.Finalize{TotalSamples: uint64(len(plans))},
		},
	}
	if err := stream.Send(fin); err != nil {
		_ = stream.CloseSend()
		return nil, fmt.Errorf("send finalize: %w", err)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetMissingPlansHandles(ctx context.Context, server common_domain.ServerMeta, start time.Time, end time.Time) ([]string, error) {
	stream, err := c.client.GetMissingPlans(ctx, &ingestorv2.GetMissingPlansRequest{
		Server: &dbmv1.ServerMetadata{
			Host: server.Host,
			Type: server.Type,
		},
		From: timestamppb.New(start),
		To:   timestamppb.New(end),
	})
	if err != nil {
		return nil, fmt.Errorf("get missing plan handles: %w", err)
	}
	resp := make([]string, 0)
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("receive missing plan handles: %w", err)
		}
		if r.GetFinalize() != nil {
			break
		}
		if r.GetChunk() != nil {
			resp = append(resp, r.GetChunk().GetPlanHandles()...)
		}
	}
	return resp, nil
}
