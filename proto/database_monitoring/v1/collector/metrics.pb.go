// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.20.1
// source: database_monitoring/v1/collector/metrics.proto

package collectorv1

import (
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DatabaseMetrics struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerId  string                 `protobuf:"bytes,1,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Types that are assignable to Metrics:
	//
	//	*DatabaseMetrics_QueryMetrics
	//	*DatabaseMetrics_SystemMetrics
	Metrics isDatabaseMetrics_Metrics `protobuf_oneof:"metrics"`
}

func (x *DatabaseMetrics) Reset() {
	*x = DatabaseMetrics{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DatabaseMetrics) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DatabaseMetrics) ProtoMessage() {}

func (x *DatabaseMetrics) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DatabaseMetrics.ProtoReflect.Descriptor instead.
func (*DatabaseMetrics) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_metrics_proto_rawDescGZIP(), []int{0}
}

func (x *DatabaseMetrics) GetServerId() string {
	if x != nil {
		return x.ServerId
	}
	return ""
}

func (x *DatabaseMetrics) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (m *DatabaseMetrics) GetMetrics() isDatabaseMetrics_Metrics {
	if m != nil {
		return m.Metrics
	}
	return nil
}

func (x *DatabaseMetrics) GetQueryMetrics() *DatabaseMetrics_QueryMetricSample {
	if x, ok := x.GetMetrics().(*DatabaseMetrics_QueryMetrics); ok {
		return x.QueryMetrics
	}
	return nil
}

func (x *DatabaseMetrics) GetSystemMetrics() *SystemMetrics {
	if x, ok := x.GetMetrics().(*DatabaseMetrics_SystemMetrics); ok {
		return x.SystemMetrics
	}
	return nil
}

type isDatabaseMetrics_Metrics interface {
	isDatabaseMetrics_Metrics()
}

type DatabaseMetrics_QueryMetrics struct {
	QueryMetrics *DatabaseMetrics_QueryMetricSample `protobuf:"bytes,3,opt,name=query_metrics,json=queryMetrics,proto3,oneof"`
}

type DatabaseMetrics_SystemMetrics struct {
	SystemMetrics *SystemMetrics `protobuf:"bytes,4,opt,name=system_metrics,json=systemMetrics,proto3,oneof"`
}

func (*DatabaseMetrics_QueryMetrics) isDatabaseMetrics_Metrics() {}

func (*DatabaseMetrics_SystemMetrics) isDatabaseMetrics_Metrics() {}

type SystemMetrics struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CpuUsage          float64                `protobuf:"fixed64,1,opt,name=cpu_usage,json=cpuUsage,proto3" json:"cpu_usage,omitempty"`
	MemoryUsage       float64                `protobuf:"fixed64,2,opt,name=memory_usage,json=memoryUsage,proto3" json:"memory_usage,omitempty"`
	ActiveConnections int64                  `protobuf:"varint,3,opt,name=active_connections,json=activeConnections,proto3" json:"active_connections,omitempty"`
	DiskIoRate        float64                `protobuf:"fixed64,4,opt,name=disk_io_rate,json=diskIoRate,proto3" json:"disk_io_rate,omitempty"`
	NetworkIoRate     float64                `protobuf:"fixed64,5,opt,name=network_io_rate,json=networkIoRate,proto3" json:"network_io_rate,omitempty"`
	Counters          []*PerformanceCounters `protobuf:"bytes,6,rep,name=counters,proto3" json:"counters,omitempty"`
}

func (x *SystemMetrics) Reset() {
	*x = SystemMetrics{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMetrics) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMetrics) ProtoMessage() {}

func (x *SystemMetrics) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMetrics.ProtoReflect.Descriptor instead.
func (*SystemMetrics) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_metrics_proto_rawDescGZIP(), []int{1}
}

func (x *SystemMetrics) GetCpuUsage() float64 {
	if x != nil {
		return x.CpuUsage
	}
	return 0
}

func (x *SystemMetrics) GetMemoryUsage() float64 {
	if x != nil {
		return x.MemoryUsage
	}
	return 0
}

func (x *SystemMetrics) GetActiveConnections() int64 {
	if x != nil {
		return x.ActiveConnections
	}
	return 0
}

func (x *SystemMetrics) GetDiskIoRate() float64 {
	if x != nil {
		return x.DiskIoRate
	}
	return 0
}

func (x *SystemMetrics) GetNetworkIoRate() float64 {
	if x != nil {
		return x.NetworkIoRate
	}
	return 0
}

func (x *SystemMetrics) GetCounters() []*PerformanceCounters {
	if x != nil {
		return x.Counters
	}
	return nil
}

type PerformanceCounters struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CounterName  string  `protobuf:"bytes,1,opt,name=counter_name,json=counterName,proto3" json:"counter_name,omitempty"`
	CounterValue int64   `protobuf:"varint,2,opt,name=counter_value,json=counterValue,proto3" json:"counter_value,omitempty"`
	CounterRate  float64 `protobuf:"fixed64,3,opt,name=counter_rate,json=counterRate,proto3" json:"counter_rate,omitempty"`
}

func (x *PerformanceCounters) Reset() {
	*x = PerformanceCounters{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PerformanceCounters) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PerformanceCounters) ProtoMessage() {}

func (x *PerformanceCounters) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PerformanceCounters.ProtoReflect.Descriptor instead.
func (*PerformanceCounters) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_metrics_proto_rawDescGZIP(), []int{2}
}

func (x *PerformanceCounters) GetCounterName() string {
	if x != nil {
		return x.CounterName
	}
	return ""
}

func (x *PerformanceCounters) GetCounterValue() int64 {
	if x != nil {
		return x.CounterValue
	}
	return 0
}

func (x *PerformanceCounters) GetCounterRate() float64 {
	if x != nil {
		return x.CounterRate
	}
	return 0
}

type DatabaseMetrics_QueryMetricSample struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QueryMetrics []*dbmv1.QueryMetric `protobuf:"bytes,1,rep,name=query_metrics,json=queryMetrics,proto3" json:"query_metrics,omitempty"`
}

func (x *DatabaseMetrics_QueryMetricSample) Reset() {
	*x = DatabaseMetrics_QueryMetricSample{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DatabaseMetrics_QueryMetricSample) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DatabaseMetrics_QueryMetricSample) ProtoMessage() {}

func (x *DatabaseMetrics_QueryMetricSample) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_metrics_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DatabaseMetrics_QueryMetricSample.ProtoReflect.Descriptor instead.
func (*DatabaseMetrics_QueryMetricSample) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_metrics_proto_rawDescGZIP(), []int{0, 0}
}

func (x *DatabaseMetrics_QueryMetricSample) GetQueryMetrics() []*dbmv1.QueryMetric {
	if x != nil {
		return x.QueryMetrics
	}
	return nil
}

var File_database_monitoring_v1_collector_metrics_proto protoreflect.FileDescriptor

var file_database_monitoring_v1_collector_metrics_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x16, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x64, 0x61, 0x74, 0x61, 0x62,
	0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76,
	0x31, 0x2f, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x84,
	0x03, 0x0a, 0x0f, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12,
	0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x60, 0x0a, 0x0d, 0x71, 0x75, 0x65,
	0x72, 0x79, 0x5f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x39, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61,
	0x73, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x0c, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x4e, 0x0a, 0x0e, 0x73,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x48, 0x00, 0x52, 0x0d, 0x73, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x1a, 0x5d, 0x0a, 0x11, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x12, 0x48, 0x0a, 0x0d, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61,
	0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31,
	0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x0c, 0x71, 0x75,
	0x65, 0x72, 0x79, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x42, 0x09, 0x0a, 0x07, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x22, 0x91, 0x02, 0x0a, 0x0d, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x70, 0x75, 0x5f, 0x75,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x63, 0x70, 0x75, 0x55,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x5f, 0x75,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0b, 0x6d, 0x65, 0x6d, 0x6f,
	0x72, 0x79, 0x55, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2d, 0x0a, 0x12, 0x61, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x11, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x20, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x69,
	0x6f, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0a, 0x64, 0x69,
	0x73, 0x6b, 0x49, 0x6f, 0x52, 0x61, 0x74, 0x65, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x5f, 0x69, 0x6f, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x01, 0x52, 0x0d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x6f, 0x52, 0x61, 0x74, 0x65,
	0x12, 0x47, 0x0a, 0x08, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x73, 0x18, 0x06, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f,
	0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x65, 0x72, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x73, 0x52,
	0x08, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x73, 0x22, 0x80, 0x01, 0x0a, 0x13, 0x50, 0x65,
	0x72, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72,
	0x73, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x5f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x65, 0x72, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x65, 0x72, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x0b, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x52, 0x61, 0x74, 0x65, 0x42, 0x4b, 0x5a, 0x49,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x75, 0x69, 0x6c, 0x68,
	0x65, 0x72, 0x6d, 0x65, 0x61, 0x72, 0x70, 0x61, 0x73, 0x73, 0x6f, 0x73, 0x2f, 0x64, 0x61, 0x74,
	0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67,
	0x5f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x3b, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_database_monitoring_v1_collector_metrics_proto_rawDescOnce sync.Once
	file_database_monitoring_v1_collector_metrics_proto_rawDescData = file_database_monitoring_v1_collector_metrics_proto_rawDesc
)

func file_database_monitoring_v1_collector_metrics_proto_rawDescGZIP() []byte {
	file_database_monitoring_v1_collector_metrics_proto_rawDescOnce.Do(func() {
		file_database_monitoring_v1_collector_metrics_proto_rawDescData = protoimpl.X.CompressGZIP(file_database_monitoring_v1_collector_metrics_proto_rawDescData)
	})
	return file_database_monitoring_v1_collector_metrics_proto_rawDescData
}

var file_database_monitoring_v1_collector_metrics_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_database_monitoring_v1_collector_metrics_proto_goTypes = []any{
	(*DatabaseMetrics)(nil),                   // 0: database_monitoring.v1.DatabaseMetrics
	(*SystemMetrics)(nil),                     // 1: database_monitoring.v1.SystemMetrics
	(*PerformanceCounters)(nil),               // 2: database_monitoring.v1.PerformanceCounters
	(*DatabaseMetrics_QueryMetricSample)(nil), // 3: database_monitoring.v1.DatabaseMetrics.QueryMetricSample
	(*timestamppb.Timestamp)(nil),             // 4: google.protobuf.Timestamp
	(*dbmv1.QueryMetric)(nil),                 // 5: database_monitoring.v1.QueryMetric
}
var file_database_monitoring_v1_collector_metrics_proto_depIdxs = []int32{
	4, // 0: database_monitoring.v1.DatabaseMetrics.timestamp:type_name -> google.protobuf.Timestamp
	3, // 1: database_monitoring.v1.DatabaseMetrics.query_metrics:type_name -> database_monitoring.v1.DatabaseMetrics.QueryMetricSample
	1, // 2: database_monitoring.v1.DatabaseMetrics.system_metrics:type_name -> database_monitoring.v1.SystemMetrics
	2, // 3: database_monitoring.v1.SystemMetrics.counters:type_name -> database_monitoring.v1.PerformanceCounters
	5, // 4: database_monitoring.v1.DatabaseMetrics.QueryMetricSample.query_metrics:type_name -> database_monitoring.v1.QueryMetric
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_database_monitoring_v1_collector_metrics_proto_init() }
func file_database_monitoring_v1_collector_metrics_proto_init() {
	if File_database_monitoring_v1_collector_metrics_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_database_monitoring_v1_collector_metrics_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*DatabaseMetrics); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_database_monitoring_v1_collector_metrics_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SystemMetrics); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_database_monitoring_v1_collector_metrics_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*PerformanceCounters); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_database_monitoring_v1_collector_metrics_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*DatabaseMetrics_QueryMetricSample); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_database_monitoring_v1_collector_metrics_proto_msgTypes[0].OneofWrappers = []any{
		(*DatabaseMetrics_QueryMetrics)(nil),
		(*DatabaseMetrics_SystemMetrics)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_database_monitoring_v1_collector_metrics_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_database_monitoring_v1_collector_metrics_proto_goTypes,
		DependencyIndexes: file_database_monitoring_v1_collector_metrics_proto_depIdxs,
		MessageInfos:      file_database_monitoring_v1_collector_metrics_proto_msgTypes,
	}.Build()
	File_database_monitoring_v1_collector_metrics_proto = out.File
	file_database_monitoring_v1_collector_metrics_proto_rawDesc = nil
	file_database_monitoring_v1_collector_metrics_proto_goTypes = nil
	file_database_monitoring_v1_collector_metrics_proto_depIdxs = nil
}
