package ports

import (
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"testing"
)

func TestIngestionSvc_IngestSnapshot(t *testing.T) {
	testCases := []struct {
		Name          string
		requestsSetup func(t *testing.T) []*collectorv1.IngestSnapshotRequest
	}{}
	for _, tt := range testCases {
		tc := tt
		t.Run(tc.Name, func(t *testing.T) {

		})
	}
}
