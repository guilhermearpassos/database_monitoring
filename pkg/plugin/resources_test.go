package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	querierv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/querier/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockCallResourceResponseSender struct {
	response *backend.CallResourceResponse
}

func (s *mockCallResourceResponseSender) Send(response *backend.CallResourceResponse) error {
	s.response = response
	return nil
}

type mockQuerierService struct {
	querierv2.UnimplementedQuerierAPIServer
}

func (m *mockQuerierService) ListServerSummary(_ context.Context, _ *querierv2.ListServerSummaryRequest) (*querierv2.ListServerSummaryResponse, error) {
	return &querierv2.ListServerSummaryResponse{}, nil
}

var testServerAddress string

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	testServerAddress = fmt.Sprintf("127.0.0.1:%d", lis.Addr().(*net.TCPAddr).Port)

	grpcServer := grpc.NewServer()
	querierv2.RegisterQuerierAPIServer(grpcServer, &mockQuerierService{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	code := m.Run()
	grpcServer.Stop()
	os.Exit(code)
}

func TestCallResourceAndHealth(t *testing.T) {
	data := map[string]interface{}{
		"apiUrl": testServerAddress,
	}
	marshal, err := json.Marshal(data)
	require.NoError(t, err)
	inst, err := NewApp(context.Background(), backend.AppInstanceSettings{
		JSONData: marshal,
		Updated:  time.Time{},
	})
	require.NoError(t, err)
	require.NotNil(t, inst)

	app, ok := inst.(*App)
	require.True(t, ok)
	defer app.Dispose()

	health, err := app.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	require.Equal(t, backend.HealthStatusOk, health.Status)

	for _, tc := range []struct {
		name      string
		method    string
		path      string
		body      []byte
		expStatus int
		expBody   []byte
	}{
		{
			name:      "get ping 200",
			method:    http.MethodGet,
			path:      "ping",
			expStatus: http.StatusOK,
		},
		{
			name:      "get echo 405",
			method:    http.MethodGet,
			path:      "echo",
			expStatus: http.StatusMethodNotAllowed,
		},
		{
			name:      "post echo 200",
			method:    http.MethodPost,
			path:      "echo",
			body:      []byte(`{"message":"ok"}`),
			expStatus: http.StatusOK,
			expBody:   []byte(`{"message":"ok"}`),
		},
		{
			name:      "getQueryDetails missing snapId 400",
			method:    http.MethodGet,
			path:      "getQueryDetails?sampleId=abc",
			expStatus: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var r mockCallResourceResponseSender
			err = app.CallResource(context.Background(), &backend.CallResourceRequest{
				Method: tc.method,
				Path:   tc.path,
				Body:   tc.body,
			}, &r)
			require.NoError(t, err)
			require.NotNil(t, r.response)
			if tc.expStatus > 0 {
				require.Equal(t, tc.expStatus, r.response.Status)
			}
			if len(tc.expBody) > 0 {
				require.Equal(t, tc.expBody, bytes.TrimSpace(r.response.Body))
			}
		})
	}
}
