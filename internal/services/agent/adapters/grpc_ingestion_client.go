package adapters

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCIngestionClient struct {
	client collectorv1.IngestionServiceClient
	trace  trace.Tracer
}

func NewGRPCIngestionClient(client collectorv1.IngestionServiceClient) *GRPCIngestionClient {
	return &GRPCIngestionClient{client: client, trace: otel.Tracer("GRPCIngestionClient")}
}

var _ domain.IngestionClient = (*GRPCIngestionClient)(nil)

func (c GRPCIngestionClient) IngestMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, server common_domain.ServerMeta, timestamp time.Time) (err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.IngestMetrics")
	span.SetAttributes(
		attribute.String("dbm.server.host", server.Host),
		attribute.String("dbm.server.type", server.Type),
		attribute.Int("dbm.metrics.count", len(metrics)),
	)
	protoMetrics := make([]*dbmv1.QueryMetric, len(metrics))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	for i, m := range metrics {
		protoMetrics[i], err = converters.QueryMetricToProto(m)
		if err != nil {
			telemetry.Error(ctx, err, "convert metric proto failed", "server", server.Host)
			return fmt.Errorf("convert metric proto: %w - %v", err, m)
		}
	}
	span.AddEvent("sending_metrics", trace.WithAttributes(attribute.Int("dbm.metrics.count", len(protoMetrics))))
	_, err = c.client.IngestMetrics(ctx, &collectorv1.DatabaseMetrics{
		Server:    &dbmv1.ServerMetadata{Host: server.Host, Type: server.Type},
		Timestamp: timestamppb.New(timestamp),
		Metrics:   &collectorv1.DatabaseMetrics_QueryMetrics{QueryMetrics: &collectorv1.DatabaseMetrics_QueryMetricSample{QueryMetrics: protoMetrics}},
	})
	if err != nil {
		telemetry.Error(ctx, err, "ingest metrics RPC failed", "server", server.Host, "count", len(metrics))
		return fmt.Errorf("ingest: %w", err)
	}
	telemetry.Info(ctx, "ingested metrics", "server", server.Host, "count", len(metrics))
	return nil
}

func (c GRPCIngestionClient) IngestSnapshot(ctx context.Context, snapshot *common_domain.DataBaseSnapshot) (err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.IngestSnapshot")
	span.SetAttributes(
		attribute.String("dbm.server.host", snapshot.SnapInfo.Server.Host),
		attribute.String("dbm.server.type", snapshot.SnapInfo.Server.Type),
		attribute.String("dbm.snapshot.id", snapshot.SnapInfo.ID),
		attribute.Int("dbm.snapshot.samples.total", len(snapshot.Samples)),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	firstChunk := true
	chunkIndex := 0
	for i := 0; i < len(snapshot.Samples); i += 50 {
		o := i + 50
		if o > len(snapshot.Samples) {
			o = len(snapshot.Samples)
		}
		samples := snapshot.Samples[i:o]
		chunkIndex++
		if firstChunk {
			firstChunk = false
			snapshot.Samples = samples
			span.AddEvent("send_snapshot_first_chunk", trace.WithAttributes(
				attribute.Int("dbm.samples.count", len(samples)),
				attribute.Int("dbm.chunk.index", chunkIndex),
			))
			_, err = c.client.IngestSnapshot(ctx, &collectorv1.IngestSnapshotRequest{
				Snapshot: converters.DatabaseSnapshotToProto(snapshot),
			})
			if err != nil {
				telemetry.Error(ctx, err, "ingest snapshot RPC failed", "snapshot_id", snapshot.SnapInfo.ID)
				return fmt.Errorf("ingest snapshot: %w", err)
			}
			telemetry.Info(ctx, "ingested snapshot first chunk", "snapshot_id", snapshot.SnapInfo.ID, "samples", len(samples))
		} else {
			protoSamples := make([]*dbmv1.QuerySample, len(samples))
			for i, sample := range samples {
				protoSamples[i] = converters.SampleToProto(sample)
			}
			span.AddEvent("send_snapshot_next_chunk", trace.WithAttributes(
				attribute.Int("dbm.samples.count", len(samples)),
				attribute.Int("dbm.chunk.index", chunkIndex),
			))
			_, err = c.client.IngestSnapshotSamples(ctx, &collectorv1.IngestSnapshotSamplesRequest{
				Id:      snapshot.SnapInfo.ID,
				Samples: protoSamples,
			})
			if err != nil {
				telemetry.Error(ctx, err, "ingest snapshot samples RPC failed", "snapshot_id", snapshot.SnapInfo.ID, "chunk", chunkIndex)
				return fmt.Errorf("ingest snapshot samples: %w", err)
			}
		}
	}
	telemetry.Info(ctx, "ingested snapshot completed", "snapshot_id", snapshot.SnapInfo.ID)
	return nil
}

func (c GRPCIngestionClient) IngestExecPlans(ctx context.Context, executionPlans map[string]*common_domain.ExecutionPlan, server common_domain.ServerMeta) (err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.IngestExecPlans")
	span.SetAttributes(
		attribute.String("dbm.server.host", server.Host),
		attribute.String("dbm.server.type", server.Type),
		attribute.Int("dbm.execplans.count", len(executionPlans)),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	protoPlans := make([]*dbmv1.ExecutionPlan, 0, len(executionPlans))
	for handle, p := range executionPlans {
		protoPlan, err2 := converters.ExecutionPlanToProto(p)
		if err2 != nil {
			telemetry.Error(ctx, err2, "convert exec plan proto failed", "handle", handle)
			return fmt.Errorf("convert plan proto: %w", err2)
		}
		protoPlans = append(protoPlans, protoPlan)
	}
	if len(protoPlans) != 0 {
		chunkIdx := 0
  for chunk := range slices.Chunk(protoPlans, 10) {
			chunkIdx++
			span.AddEvent("sending_exec_plan_chunk", trace.WithAttributes(
				attribute.Int("dbm.chunk.index", chunkIdx),
				attribute.Int("dbm.chunk.size", len(chunk)),
			))
			_, err = c.client.IngestExecutionPlans(ctx, &collectorv1.IngestExecutionPlansRequest{Plans: chunk})
			if err != nil {
				telemetry.Error(ctx, err, "ingest exec plans RPC failed", "server", server.Host, "chunk", chunkIdx)
				return fmt.Errorf("ingest execution plans: %w", err)
			}
		}
	}
	telemetry.Info(ctx, "ingested exec plans", "server", server.Host, "plans", len(protoPlans))
	return nil
}

func (c GRPCIngestionClient) GetKnownPlanHandles(ctx context.Context, server common_domain.ServerMeta) (_ map[string]struct{}, err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.GetKnownPlanHandles")
	span.SetAttributes(
		attribute.String("dbm.server.host", server.Host),
		attribute.String("dbm.server.type", server.Type),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	currPage := int32(1)
	serverMetadata := &dbmv1.ServerMetadata{
		Host: server.Host,
		Type: server.Type,
	}
	span.AddEvent("request_known_plan_handles", trace.WithAttributes(
		attribute.Int("dbm.page", int(currPage)),
		attribute.Int("dbm.page_size", 100),
	))
	knownHandles, err := c.client.GetKnownPlanHandles(ctx,
		&collectorv1.GetKnownPlanHandlesRequest{
			Server:     serverMetadata,
			PageSize:   100,
			PageNumber: currPage,
		})
	if err != nil {
		if grpcErr, ok := status.FromError(err); ok {
			if grpcErr.Code() == codes.NotFound {
				telemetry.Info(ctx, "no known plan handles found", "server", server.Host)
				return nil, nil
			}
			telemetry.Error(ctx, err, "get known plan handles RPC failed", "server", server.Host, "page", currPage)
			return nil, fmt.Errorf("error getting known plan handles for %s: %w", server.Host, err)

		}
		telemetry.Error(ctx, err, "get known plan handles failed", "server", server.Host, "page", currPage)
		return nil, fmt.Errorf("error getting known plan handles for %s: %w", server.Host, err)
	}
	span.SetAttributes(attribute.Int("dbm.total_pages", int(knownHandles.TotalPages)))
	knownHandlesSlice := make(map[string]struct{}, len(knownHandles.Handles))
	for _, data := range knownHandles.Handles {
		knownHandlesSlice[data] = struct{}{}
	}
	currPage++
	for currPage <= knownHandles.TotalPages {
		span.AddEvent("request_known_plan_handles_page", trace.WithAttributes(
			attribute.Int("dbm.page", int(currPage)),
			attribute.Int("dbm.page_size", 100),
		))
		knownHandles, err = c.client.GetKnownPlanHandles(ctx,
			&collectorv1.GetKnownPlanHandlesRequest{
				Server:     serverMetadata,
				PageSize:   100,
				PageNumber: currPage,
			})
		if err != nil {
			telemetry.Error(ctx, err, "get known plan handles page failed", "server", server.Host, "page", currPage)
			return nil, fmt.Errorf("error getting known plan handles for %s page %d: %w", server.Host, currPage, err)
		}
		for _, data := range knownHandles.Handles {
			knownHandlesSlice[data] = struct{}{}
		}
		currPage++
	}
	telemetry.Info(ctx, "fetched known plan handles", "server", server.Host, "count", len(knownHandlesSlice))
	return knownHandlesSlice, nil
}
