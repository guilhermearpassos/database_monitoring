// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.20.1
// source: database_monitoring/v1/collector/collector_api.proto

package collectorv1

import (
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RegisterAgentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TargetHost   string   `protobuf:"bytes,1,opt,name=target_host,json=targetHost,proto3" json:"target_host,omitempty"`
	TargetType   string   `protobuf:"bytes,2,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	AgentVersion string   `protobuf:"bytes,3,opt,name=agent_version,json=agentVersion,proto3" json:"agent_version,omitempty"`
	Tags         []string `protobuf:"bytes,4,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *RegisterAgentRequest) Reset() {
	*x = RegisterAgentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterAgentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterAgentRequest) ProtoMessage() {}

func (x *RegisterAgentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterAgentRequest.ProtoReflect.Descriptor instead.
func (*RegisterAgentRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_collector_api_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterAgentRequest) GetTargetHost() string {
	if x != nil {
		return x.TargetHost
	}
	return ""
}

func (x *RegisterAgentRequest) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *RegisterAgentRequest) GetAgentVersion() string {
	if x != nil {
		return x.AgentVersion
	}
	return ""
}

func (x *RegisterAgentRequest) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

type RegisterAgentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RegisterAgentResponse) Reset() {
	*x = RegisterAgentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterAgentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterAgentResponse) ProtoMessage() {}

func (x *RegisterAgentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterAgentResponse.ProtoReflect.Descriptor instead.
func (*RegisterAgentResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_collector_api_proto_rawDescGZIP(), []int{1}
}

type IngestMetricsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *IngestMetricsResponse) Reset() {
	*x = IngestMetricsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestMetricsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestMetricsResponse) ProtoMessage() {}

func (x *IngestMetricsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IngestMetricsResponse.ProtoReflect.Descriptor instead.
func (*IngestMetricsResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_collector_api_proto_rawDescGZIP(), []int{2}
}

func (x *IngestMetricsResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *IngestMetricsResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type IngestSnapshotRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Snapshot *dbmv1.DBSnapshot `protobuf:"bytes,1,opt,name=snapshot,proto3" json:"snapshot,omitempty"`
}

func (x *IngestSnapshotRequest) Reset() {
	*x = IngestSnapshotRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestSnapshotRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestSnapshotRequest) ProtoMessage() {}

func (x *IngestSnapshotRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IngestSnapshotRequest.ProtoReflect.Descriptor instead.
func (*IngestSnapshotRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_collector_api_proto_rawDescGZIP(), []int{3}
}

func (x *IngestSnapshotRequest) GetSnapshot() *dbmv1.DBSnapshot {
	if x != nil {
		return x.Snapshot
	}
	return nil
}

type IngestSnapshotResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *IngestSnapshotResponse) Reset() {
	*x = IngestSnapshotResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestSnapshotResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestSnapshotResponse) ProtoMessage() {}

func (x *IngestSnapshotResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_collector_collector_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IngestSnapshotResponse.ProtoReflect.Descriptor instead.
func (*IngestSnapshotResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_collector_collector_api_proto_rawDescGZIP(), []int{4}
}

func (x *IngestSnapshotResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *IngestSnapshotResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_database_monitoring_v1_collector_collector_api_proto protoreflect.FileDescriptor

var file_database_monitoring_v1_collector_collector_api_proto_rawDesc = []byte{
	0x0a, 0x34, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x2f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x5f, 0x61, 0x70, 0x69,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x25, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x91, 0x01, 0x0a, 0x14, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x48, 0x6f, 0x73, 0x74,
	0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x23, 0x0a, 0x0d, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x22, 0x17, 0x0a, 0x15, 0x52, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x4b, 0x0a, 0x15, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x22, 0x57, 0x0a, 0x15, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68,
	0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3e, 0x0a, 0x08, 0x73, 0x6e, 0x61,
	0x70, 0x73, 0x68, 0x6f, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x64, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e,
	0x67, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x42, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52,
	0x08, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x22, 0x4c, 0x0a, 0x16, 0x49, 0x6e, 0x67,
	0x65, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0xda, 0x02, 0x0a, 0x10, 0x49, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x6c, 0x0a, 0x0d,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x2c, 0x2e,
	0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41,
	0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x64, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e,
	0x67, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41, 0x67, 0x65,
	0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x67, 0x0a, 0x0d, 0x49, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x27, 0x2e, 0x64, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e,
	0x67, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x1a, 0x2d, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f,
	0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x6f, 0x0a, 0x0e, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x53, 0x6e, 0x61,
	0x70, 0x73, 0x68, 0x6f, 0x74, 0x12, 0x2d, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x49,
	0x6e, 0x67, 0x65, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f,
	0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x4b, 0x5a, 0x49, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x67, 0x75, 0x69, 0x6c, 0x68, 0x65, 0x72, 0x6d, 0x65, 0x61, 0x72, 0x70, 0x61,
	0x73, 0x73, 0x6f, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f,
	0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6c, 0x6c,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x3b, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_database_monitoring_v1_collector_collector_api_proto_rawDescOnce sync.Once
	file_database_monitoring_v1_collector_collector_api_proto_rawDescData = file_database_monitoring_v1_collector_collector_api_proto_rawDesc
)

func file_database_monitoring_v1_collector_collector_api_proto_rawDescGZIP() []byte {
	file_database_monitoring_v1_collector_collector_api_proto_rawDescOnce.Do(func() {
		file_database_monitoring_v1_collector_collector_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_database_monitoring_v1_collector_collector_api_proto_rawDescData)
	})
	return file_database_monitoring_v1_collector_collector_api_proto_rawDescData
}

var file_database_monitoring_v1_collector_collector_api_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_database_monitoring_v1_collector_collector_api_proto_goTypes = []any{
	(*RegisterAgentRequest)(nil),   // 0: database_monitoring.v1.RegisterAgentRequest
	(*RegisterAgentResponse)(nil),  // 1: database_monitoring.v1.RegisterAgentResponse
	(*IngestMetricsResponse)(nil),  // 2: database_monitoring.v1.IngestMetricsResponse
	(*IngestSnapshotRequest)(nil),  // 3: database_monitoring.v1.IngestSnapshotRequest
	(*IngestSnapshotResponse)(nil), // 4: database_monitoring.v1.IngestSnapshotResponse
	(*dbmv1.DBSnapshot)(nil),       // 5: database_monitoring.v1.DBSnapshot
	(*DatabaseMetrics)(nil),        // 6: database_monitoring.v1.DatabaseMetrics
}
var file_database_monitoring_v1_collector_collector_api_proto_depIdxs = []int32{
	5, // 0: database_monitoring.v1.IngestSnapshotRequest.snapshot:type_name -> database_monitoring.v1.DBSnapshot
	0, // 1: database_monitoring.v1.IngestionService.RegisterAgent:input_type -> database_monitoring.v1.RegisterAgentRequest
	6, // 2: database_monitoring.v1.IngestionService.IngestMetrics:input_type -> database_monitoring.v1.DatabaseMetrics
	3, // 3: database_monitoring.v1.IngestionService.IngestSnapshot:input_type -> database_monitoring.v1.IngestSnapshotRequest
	1, // 4: database_monitoring.v1.IngestionService.RegisterAgent:output_type -> database_monitoring.v1.RegisterAgentResponse
	2, // 5: database_monitoring.v1.IngestionService.IngestMetrics:output_type -> database_monitoring.v1.IngestMetricsResponse
	4, // 6: database_monitoring.v1.IngestionService.IngestSnapshot:output_type -> database_monitoring.v1.IngestSnapshotResponse
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_database_monitoring_v1_collector_collector_api_proto_init() }
func file_database_monitoring_v1_collector_collector_api_proto_init() {
	if File_database_monitoring_v1_collector_collector_api_proto != nil {
		return
	}
	file_database_monitoring_v1_collector_metrics_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_database_monitoring_v1_collector_collector_api_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RegisterAgentRequest); i {
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
		file_database_monitoring_v1_collector_collector_api_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*RegisterAgentResponse); i {
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
		file_database_monitoring_v1_collector_collector_api_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*IngestMetricsResponse); i {
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
		file_database_monitoring_v1_collector_collector_api_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*IngestSnapshotRequest); i {
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
		file_database_monitoring_v1_collector_collector_api_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*IngestSnapshotResponse); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_database_monitoring_v1_collector_collector_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_database_monitoring_v1_collector_collector_api_proto_goTypes,
		DependencyIndexes: file_database_monitoring_v1_collector_collector_api_proto_depIdxs,
		MessageInfos:      file_database_monitoring_v1_collector_collector_api_proto_msgTypes,
	}.Build()
	File_database_monitoring_v1_collector_collector_api_proto = out.File
	file_database_monitoring_v1_collector_collector_api_proto_rawDesc = nil
	file_database_monitoring_v1_collector_collector_api_proto_goTypes = nil
	file_database_monitoring_v1_collector_collector_api_proto_depIdxs = nil
}
