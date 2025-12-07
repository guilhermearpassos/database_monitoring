package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"testing"
	"time"
)

// mockCallResourceResponseSender implements backend.CallResourceResponseSender
// for use in tests.
type mockCallResourceResponseSender struct {
	response *backend.CallResourceResponse
}

// Send sets the received *backend.CallResourceResponse to s.response
func (s *mockCallResourceResponseSender) Send(response *backend.CallResourceResponse) error {
	s.response = response
	return nil
}

type MockGrpcService struct {
	*dbmv1.UnimplementedDBMApiServer
	*dbmv1.UnimplementedDBMSupportApiServer
}

var testServerAddress string

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	testServerAddress = fmt.Sprintf("localhost:%d", lis.Addr().(*net.TCPAddr).Port)

	grpcServer := grpc.NewServer()
	mockService := &MockGrpcService{}
	dbmv1.RegisterDBMApiServer(grpcServer, mockService)
	dbmv1.RegisterDBMSupportApiServer(grpcServer, mockService)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer grpcServer.Stop()

	m.Run()
}

// TestCallResource tests CallResource calls, using backend.CallResourceRequest and backend.CallResourceResponse.
// This ensures the httpadapter for CallResource works correctly.
func TestCallResource(t *testing.T) {
	// Initialize app
	data := map[string]interface{}{
		"APIURL": testServerAddress,
	}
	marshal, err2 := json.Marshal(data)
	require.NoError(t, err2)
	inst, err := NewApp(context.Background(), backend.AppInstanceSettings{
		JSONData:                marshal,
		DecryptedSecureJSONData: nil,
		Updated:                 time.Time{},
		APIVersion:              "",
	})
	if err != nil {
		t.Fatalf("new app: %s", err)
	}
	if inst == nil {
		t.Fatal("inst must not be nil")
	}
	app, ok := inst.(*App)
	if !ok {
		t.Fatal("inst must be of type *App")
	}

	// Set up and run test cases
	for _, tc := range []struct {
		name string

		method string
		path   string
		body   []byte

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
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Request by calling CallResource. This tests the httpadapter.
			var r mockCallResourceResponseSender
			err = app.CallResource(context.Background(), &backend.CallResourceRequest{
				Method: tc.method,
				Path:   tc.path,
				Body:   tc.body,
			}, &r)
			if err != nil {
				t.Fatalf("CallResource error: %s", err)
			}
			if r.response == nil {
				t.Fatal("no response received from CallResource")
			}
			if tc.expStatus > 0 && tc.expStatus != r.response.Status {
				t.Errorf("response status should be %d, got %d", tc.expStatus, r.response.Status)
			}
			if len(tc.expBody) > 0 {
				if tb := bytes.TrimSpace(r.response.Body); !bytes.Equal(tb, tc.expBody) {
					t.Errorf("response body should be %s, got %s", tc.expBody, tb)
				}
			}
		})
	}
}
