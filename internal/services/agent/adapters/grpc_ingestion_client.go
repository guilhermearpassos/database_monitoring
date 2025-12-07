package adapters

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"slices"
	"time"
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
	protoMetrics := make([]*dbmv1.QueryMetric, len(metrics))
	defer func() {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.End()
	}()
	for i, m := range metrics {
		protoMetrics[i], err = converters.QueryMetricToProto(m)
		if err != nil {
			return fmt.Errorf("convert metric proto: %w - %v", err, m)
		}
	}
	_, err = c.client.IngestMetrics(ctx, &collectorv1.DatabaseMetrics{
		Server:    &dbmv1.ServerMetadata{Host: server.Host, Type: server.Type},
		Timestamp: timestamppb.New(timestamp),
		Metrics:   &collectorv1.DatabaseMetrics_QueryMetrics{QueryMetrics: &collectorv1.DatabaseMetrics_QueryMetricSample{QueryMetrics: protoMetrics}},
	})
	if err != nil {
		return fmt.Errorf("ingest: %w", err)
	}
	return nil
}

func (c GRPCIngestionClient) IngestSnapshot(ctx context.Context, snapshot *common_domain.DataBaseSnapshot) (err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.IngestSnapshot")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	sampleChunks := slices.Chunk(snapshot.Samples, 50)
	firstChunk := true
	for samples := range sampleChunks {
		if firstChunk {
			firstChunk = false
			snapshot.Samples = samples
			_, err = c.client.IngestSnapshot(ctx, &collectorv1.IngestSnapshotRequest{
				Snapshot: converters.DatabaseSnapshotToProto(snapshot),
			})
			if err != nil {
				return fmt.Errorf("ingest snapshot: %w", err)
			}
		} else {

			protoSamples := make([]*dbmv1.QuerySample, len(samples))
			for i, sample := range samples {
				protoSamples[i] = converters.SampleToProto(sample)
			}
			_, err = c.client.IngestSnapshotSamples(ctx, &collectorv1.IngestSnapshotSamplesRequest{
				Id:      snapshot.SnapInfo.ID,
				Samples: protoSamples,
			})
			if err != nil {
				return fmt.Errorf("ingest snapshot samples: %w", err)
			}
		}
	}
	return nil
}

func (c GRPCIngestionClient) IngestExecPlans(ctx context.Context, executionPlans map[string]*common_domain.ExecutionPlan, server common_domain.ServerMeta) (err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.IngestExecPlans")
	defer func() {
		if err != nil {

			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	protoPlans := make([]*dbmv1.ExecutionPlan, 0, len(executionPlans))
	for _, p := range executionPlans {
		protoPlan, err2 := converters.ExecutionPlanToProto(p)
		if err2 != nil {
			return fmt.Errorf("convert plan proto: %w", err2)
		}
		protoPlans = append(protoPlans, protoPlan)
	}
	if len(protoPlans) != 0 {
		for chunk := range slices.Chunk(protoPlans, 10) {
			_, err = c.client.IngestExecutionPlans(ctx, &collectorv1.IngestExecutionPlansRequest{Plans: chunk})
			if err != nil {
				return fmt.Errorf("ingest execution plans: %w", err)
			}

		}
	}
	return nil
}

func (c GRPCIngestionClient) GetKnownPlanHandles(ctx context.Context, server common_domain.ServerMeta) (_ map[string]struct{}, err error) {
	ctx, span := c.trace.Start(ctx, "GRPCIngestionClient.GetKnownPlanHandles")
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
	knownHandles, err := c.client.GetKnownPlanHandles(context.Background(),
		&collectorv1.GetKnownPlanHandlesRequest{
			Server:     serverMetadata,
			PageSize:   100,
			PageNumber: currPage,
		})
	if err != nil {
		if grpcErr, ok := status.FromError(err); ok {
			if grpcErr.Code() == codes.NotFound {
				return nil, nil
			}
			return nil, fmt.Errorf("error getting known plan handles for %s: %w", server.Host, err)

		}
		return nil, fmt.Errorf("error getting known plan handles for %s: %w", server.Host, err)
	}
	knownHandlesSlice := make(map[string]struct{}, len(knownHandles.Handles))
	for _, data := range knownHandles.Handles {
		knownHandlesSlice[data] = struct{}{}
	}
	currPage++
	for currPage <= knownHandles.TotalPages {
		knownHandles, err = c.client.GetKnownPlanHandles(context.Background(),
			&collectorv1.GetKnownPlanHandlesRequest{
				Server:     serverMetadata,
				PageSize:   100,
				PageNumber: currPage,
			})
		if err != nil {
			return nil, fmt.Errorf("error getting known plan handles for %s page %d: %w", server.Host, currPage, err)
		}
		for _, data := range knownHandles.Handles {
			knownHandlesSlice[data] = struct{}{}
		}
		currPage++
	}

	return knownHandlesSlice, nil
}
