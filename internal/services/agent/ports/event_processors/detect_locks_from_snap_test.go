package event_processors

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Mock implementations
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordLockDuration(server, database, waitType, table string, duration float64) {
	m.Called(server, database, waitType, table, duration)
}

func (m *MockMetricsCollector) IncrementTotalLocks(server, database string) {
	m.Called(server, database)
}

type MockSQLParser struct {
	mock.Mock
}

func (m *MockSQLParser) ExtractTablesFromQuery(query string) ([]string, error) {
	args := m.Called(query)
	return args.Get(0).([]string), args.Error(1)
}

func TestMetricsDetector_processSnapshot(t *testing.T) {
	tests := []struct {
		name         string
		snapshot     *common_domain.DataBaseSnapshot
		expectations func(*MockMetricsCollector, *MockSQLParser)
	}{
		{
			name: "processes blocked sample correctly",
			snapshot: &common_domain.DataBaseSnapshot{
				SnapInfo: common_domain.SnapInfo{
					Server: common_domain.ServerMeta{Host: "test-server"},
				},
				Samples: []*common_domain.QuerySample{
					{
						IsBlocked: true,
						Database:  common_domain.DataBaseMetadata{DatabaseName: "test-db"},
						Session:   common_domain.SessionMetadata{SessionID: "session-1"},
						Text:      "SELECT * FROM users",
						Wait: common_domain.WaitMetadata{
							WaitTime: 5000, // 5 seconds
							WaitType: stringPtr("LCK_M_S"),
						},
					},
				},
			},
			expectations: func(mc *MockMetricsCollector, sp *MockSQLParser) {
				sp.On("ExtractTablesFromQuery", "SELECT * FROM users").Return([]string{"users"}, nil)
				mc.On("RecordLockDuration", "test-server", "test-db", "LCK_M_S", "users", 5.0).Once()
				mc.On("IncrementTotalLocks", "test-server", "test-db").Once()
			},
		},
		{
			name: "skips non-blocked samples",
			snapshot: &common_domain.DataBaseSnapshot{
				SnapInfo: common_domain.SnapInfo{
					Server: common_domain.ServerMeta{Host: "test-server"},
				},
				Samples: []*common_domain.QuerySample{
					{
						IsBlocked: false,
						Database:  common_domain.DataBaseMetadata{DatabaseName: "test-db"},
						Session:   common_domain.SessionMetadata{SessionID: "session-1"},
						Text:      "SELECT * FROM users",
					},
				},
			},
			expectations: func(mc *MockMetricsCollector, sp *MockSQLParser) {
				// No expectations - nothing should be called
			},
		},
		{
			name: "handles multiple tables in query",
			snapshot: &common_domain.DataBaseSnapshot{
				SnapInfo: common_domain.SnapInfo{
					Server: common_domain.ServerMeta{Host: "test-server"},
				},
				Samples: []*common_domain.QuerySample{
					{
						IsBlocked: true,
						Database:  common_domain.DataBaseMetadata{DatabaseName: "test-db"},
						Session:   common_domain.SessionMetadata{SessionID: "session-1"},
						Text:      "SELECT * FROM users JOIN orders ON users.id = orders.user_id",
						Wait: common_domain.WaitMetadata{
							WaitTime: 3000,
							WaitType: stringPtr("LCK_M_X"),
						},
					},
				},
			},
			expectations: func(mc *MockMetricsCollector, sp *MockSQLParser) {
				sp.On("ExtractTablesFromQuery", mock.AnythingOfType("string")).Return([]string{"users", "orders"}, nil)
				mc.On("RecordLockDuration", "test-server", "test-db", "LCK_M_X", "users", 3.0).Once()
				mc.On("RecordLockDuration", "test-server", "test-db", "LCK_M_X", "orders", 3.0).Once()
				mc.On("IncrementTotalLocks", "test-server", "test-db").Once()
			},
		},
		{
			name: "handles nil wait type",
			snapshot: &common_domain.DataBaseSnapshot{
				SnapInfo: common_domain.SnapInfo{
					Server: common_domain.ServerMeta{Host: "test-server"},
				},
				Samples: []*common_domain.QuerySample{
					{
						IsBlocked: true,
						Database:  common_domain.DataBaseMetadata{DatabaseName: "test-db"},
						Session:   common_domain.SessionMetadata{SessionID: "session-1"},
						Text:      "SELECT * FROM users",
						Wait: common_domain.WaitMetadata{
							WaitTime: 1000,
							WaitType: nil,
						},
					},
				},
			},
			expectations: func(mc *MockMetricsCollector, sp *MockSQLParser) {
				sp.On("ExtractTablesFromQuery", "SELECT * FROM users").Return([]string{"users"}, nil)
				mc.On("RecordLockDuration", "test-server", "test-db", "unknown", "users", 1.0).Once()
				mc.On("IncrementTotalLocks", "test-server", "test-db").Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollector := new(MockMetricsCollector)
			mockParser := new(MockSQLParser)

			detector := NewMetricsDetector(nil, mockCollector, mockParser)

			tt.expectations(mockCollector, mockParser)

			detector.processSnapshot(tt.snapshot)

			mockCollector.AssertExpectations(t)
			mockParser.AssertExpectations(t)
		})
	}
}

func TestGenerateLockKey(t *testing.T) {
	tests := []struct {
		name      string
		server    string
		database  string
		sessionID string
		expected  string
	}{
		{
			name:      "generates correct key",
			server:    "server1",
			database:  "db1",
			sessionID: "session1",
			expected:  "server1:db1:session1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateLockKey(tt.server, tt.database, tt.sessionID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
