package ports

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/adapters/state"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeSnapshotStream struct {
	grpc.ServerStream
	ctx    context.Context
	msgs   []*ingestorv2.SnapshotUploadRequest
	idx    int
	recvEr error
	resp   *ingestorv2.SnapshotUploadResult
}

func (f *fakeSnapshotStream) Context() context.Context { return f.ctx }

func (f *fakeSnapshotStream) Recv() (*ingestorv2.SnapshotUploadRequest, error) {
	if f.idx < len(f.msgs) {
		m := f.msgs[f.idx]
		f.idx++
		return m, nil
	}
	if f.recvEr != nil {
		return nil, f.recvEr
	}
	return nil, io.EOF
}

func (f *fakeSnapshotStream) SendAndClose(res *ingestorv2.SnapshotUploadResult) error {
	f.resp = res
	return nil
}

type fakeSnapshotRepo struct {
	ensureErr error
	saveErr   error
}

func (f *fakeSnapshotRepo) EnsureSnapshot(_ context.Context, _ domain.SnapshotHeader) error {
	return f.ensureErr
}
func (f *fakeSnapshotRepo) SaveSamples(_ context.Context, _ string, _ uint32, _ []*dbmv1.QuerySample) error {
	return f.saveErr
}
func (f *fakeSnapshotRepo) FinalizeSnapshot(_ context.Context, _ string) error { return nil }
func (f *fakeSnapshotRepo) SaveExecutionPlans(_ context.Context, _ []*common_domain.ExecutionPlan) error {
	return nil
}
func (f *fakeSnapshotRepo) GetMissingPlans(_ context.Context, _ common_domain.ServerMeta, _, _ time.Time) ([]string, error) {
	return nil, nil
}
func (f *fakeSnapshotRepo) PurgeSnapshots(_ context.Context, _, _ time.Time, _ int) error { return nil }
func (f *fakeSnapshotRepo) PurgeAllSnapshots(_ context.Context) error                    { return nil }
func (f *fakeSnapshotRepo) PurgeUnboundedPlans(_ context.Context, _ int) error           { return nil }
func (f *fakeSnapshotRepo) GetUnanalizedPlans(_ context.Context, _ int) ([]*common_domain.ExecutionPlan, error) {
	return nil, nil
}
func (f *fakeSnapshotRepo) SetPlanAnalisysBatch(_ context.Context, _ []domain.PlanAnalisysBatchResult) error {
	return nil
}

func TestIngestSnapshotStream_ErrorHandling(t *testing.T) {
	t.Run("returns PARTIAL when chunks are missing", func(t *testing.T) {
		svc := NewService(app.NewApplicationWithAdapters(state.NewMemoryStore(), &fakeSnapshotRepo{}, nil))
		stream := &fakeSnapshotStream{
			ctx: context.Background(),
			msgs: []*ingestorv2.SnapshotUploadRequest{
				{
					SnapshotId: "snap-1",
					Payload: &ingestorv2.SnapshotUploadRequest_Header{
						Header: &ingestorv2.SnapshotHeader{
							Timestamp:      timestamppb.Now(),
							Server:         &dbmv1.ServerMetadata{Host: "db1", Type: "mssql"},
							ExpectedChunks: 2,
						},
					},
				},
				{
					SnapshotId: "snap-1",
					Payload: &ingestorv2.SnapshotUploadRequest_Chunk{
						Chunk: &ingestorv2.SampleChunk{ChunkSeq: 1},
					},
				},
				{
					SnapshotId: "snap-1",
					Payload: &ingestorv2.SnapshotUploadRequest_Finalize{
						Finalize: &ingestorv2.Finalize{TotalSamples: 0},
					},
				},
			},
		}

		err := svc.IngestSnapshotStream(stream)
		require.NoError(t, err)
		require.NotNil(t, stream.resp)
		require.Equal(t, ingestorv2.SnapshotUploadResult_PARTIAL, stream.resp.Status)
		require.NotEmpty(t, stream.resp.Message)
	})

	t.Run("propagates non-EOF Recv errors instead of returning OK", func(t *testing.T) {
		svc := NewService(app.NewApplicationWithAdapters(state.NewMemoryStore(), &fakeSnapshotRepo{}, nil))
		recvErr := errors.New("stream reset by peer")
		stream := &fakeSnapshotStream{
			ctx:    context.Background(),
			recvEr: recvErr,
		}

		err := svc.IngestSnapshotStream(stream)
		require.ErrorIs(t, err, recvErr)
		require.Nil(t, stream.resp)
	})

	t.Run("propagates repository errors instead of returning OK", func(t *testing.T) {
		repoErr := errors.New("postgres unavailable")
		svc := NewService(app.NewApplicationWithAdapters(state.NewMemoryStore(), &fakeSnapshotRepo{ensureErr: repoErr}, nil))
		stream := &fakeSnapshotStream{
			ctx: context.Background(),
			msgs: []*ingestorv2.SnapshotUploadRequest{
				{
					SnapshotId: "snap-2",
					Payload: &ingestorv2.SnapshotUploadRequest_Header{
						Header: &ingestorv2.SnapshotHeader{
							Timestamp: timestamppb.Now(),
							Server:    &dbmv1.ServerMetadata{Host: "db1", Type: "mssql"},
						},
					},
				},
			},
		}

		err := svc.IngestSnapshotStream(stream)
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, stream.resp)
	})
}

type fakeDeadlockRepo struct {
	saved []*common_domain.DeadlockReport
}

func (f *fakeDeadlockRepo) SaveDeadlocks(_ context.Context, _ common_domain.ServerMeta, deadlocks []*common_domain.DeadlockReport) (int, error) {
	f.saved = append(f.saved, deadlocks...)
	return len(deadlocks), nil
}

func (f *fakeDeadlockRepo) PurgeDeadlocks(_ context.Context, _, _ time.Time, _ int) error {
	return nil
}

func TestIngestDeadlocks(t *testing.T) {
	dlRepo := &fakeDeadlockRepo{}
	svc := NewService(app.NewApplicationWithAdapters(state.NewMemoryStore(), &fakeSnapshotRepo{}, nil, dlRepo))
	resp, err := svc.IngestDeadlocks(context.Background(), &ingestorv2.IngestDeadlocksRequest{
		Server: &dbmv1.ServerMetadata{Host: "sql-prod", Type: "mssql"},
		Deadlocks: []*dbmv1.DeadlockReport{
			{
				Id:         "dl-1",
				OccurredAt: timestamppb.Now(),
				XmlReport: `<deadlock>
					<victim-list><victimProcess id="p1" /></victim-list>
					<process-list><process id="p1" spid="55" waittime="1200" currentdbname="OrdersDB"><inputbuf>UPDATE t SET x=1</inputbuf></process></process-list>
					<resource-list><keylock objectname="OrdersDB.dbo.T" indexname="PK_T" id="l1" mode="X"><owner-list><owner id="p1" mode="S"/></owner-list><waiter-list><waiter id="p1" mode="X" requestType="convert"/></waiter-list></keylock></resource-list>
				</deadlock>`,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), resp.GetStoredCount())
	require.Len(t, dlRepo.saved, 1)
	require.Equal(t, "lock_upgrade", dlRepo.saved[0].Pattern)
	require.Equal(t, "55", dlRepo.saved[0].VictimSPID)
	require.NotEmpty(t, dlRepo.saved[0].Fingerprint)
}
