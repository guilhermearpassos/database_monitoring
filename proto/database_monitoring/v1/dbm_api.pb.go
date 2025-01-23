// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.20.1
// source: database_monitoring/v1/dbm_api.proto

package dbmv1

import (
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

type GetSnapshotRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetSnapshotRequest) Reset() {
	*x = GetSnapshotRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetSnapshotRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSnapshotRequest) ProtoMessage() {}

func (x *GetSnapshotRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSnapshotRequest.ProtoReflect.Descriptor instead.
func (*GetSnapshotRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{0}
}

func (x *GetSnapshotRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetSnapshotResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Snapshot *DBSnapshot `protobuf:"bytes,1,opt,name=snapshot,proto3" json:"snapshot,omitempty"`
}

func (x *GetSnapshotResponse) Reset() {
	*x = GetSnapshotResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetSnapshotResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSnapshotResponse) ProtoMessage() {}

func (x *GetSnapshotResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSnapshotResponse.ProtoReflect.Descriptor instead.
func (*GetSnapshotResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{1}
}

func (x *GetSnapshotResponse) GetSnapshot() *DBSnapshot {
	if x != nil {
		return x.Snapshot
	}
	return nil
}

type ListSnapshotsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start      *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	End        *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=end,proto3" json:"end,omitempty"`
	Host       string                 `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
	Database   string                 `protobuf:"bytes,4,opt,name=database,proto3" json:"database,omitempty"`
	PageSize   int32                  `protobuf:"varint,5,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	PageNumber int64                  `protobuf:"varint,6,opt,name=page_number,json=pageNumber,proto3" json:"page_number,omitempty"`
}

func (x *ListSnapshotsRequest) Reset() {
	*x = ListSnapshotsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListSnapshotsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListSnapshotsRequest) ProtoMessage() {}

func (x *ListSnapshotsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListSnapshotsRequest.ProtoReflect.Descriptor instead.
func (*ListSnapshotsRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{2}
}

func (x *ListSnapshotsRequest) GetStart() *timestamppb.Timestamp {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *ListSnapshotsRequest) GetEnd() *timestamppb.Timestamp {
	if x != nil {
		return x.End
	}
	return nil
}

func (x *ListSnapshotsRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *ListSnapshotsRequest) GetDatabase() string {
	if x != nil {
		return x.Database
	}
	return ""
}

func (x *ListSnapshotsRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListSnapshotsRequest) GetPageNumber() int64 {
	if x != nil {
		return x.PageNumber
	}
	return 0
}

type ListSnapshotsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Snapshots  []*DBSnapshot `protobuf:"bytes,1,rep,name=snapshots,proto3" json:"snapshots,omitempty"`
	PageNumber int64         `protobuf:"varint,2,opt,name=page_number,json=pageNumber,proto3" json:"page_number,omitempty"`
	TotalCount int64         `protobuf:"varint,3,opt,name=total_count,json=totalCount,proto3" json:"total_count,omitempty"`
}

func (x *ListSnapshotsResponse) Reset() {
	*x = ListSnapshotsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListSnapshotsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListSnapshotsResponse) ProtoMessage() {}

func (x *ListSnapshotsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListSnapshotsResponse.ProtoReflect.Descriptor instead.
func (*ListSnapshotsResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{3}
}

func (x *ListSnapshotsResponse) GetSnapshots() []*DBSnapshot {
	if x != nil {
		return x.Snapshots
	}
	return nil
}

func (x *ListSnapshotsResponse) GetPageNumber() int64 {
	if x != nil {
		return x.PageNumber
	}
	return 0
}

func (x *ListSnapshotsResponse) GetTotalCount() int64 {
	if x != nil {
		return x.TotalCount
	}
	return 0
}

type ListServerSummaryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	End   *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *ListServerSummaryRequest) Reset() {
	*x = ListServerSummaryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListServerSummaryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListServerSummaryRequest) ProtoMessage() {}

func (x *ListServerSummaryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListServerSummaryRequest.ProtoReflect.Descriptor instead.
func (*ListServerSummaryRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{4}
}

func (x *ListServerSummaryRequest) GetStart() *timestamppb.Timestamp {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *ListServerSummaryRequest) GetEnd() *timestamppb.Timestamp {
	if x != nil {
		return x.End
	}
	return nil
}

type ListServerSummaryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Servers []*ServerSummary `protobuf:"bytes,1,rep,name=servers,proto3" json:"servers,omitempty"`
}

func (x *ListServerSummaryResponse) Reset() {
	*x = ListServerSummaryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListServerSummaryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListServerSummaryResponse) ProtoMessage() {}

func (x *ListServerSummaryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListServerSummaryResponse.ProtoReflect.Descriptor instead.
func (*ListServerSummaryResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{5}
}

func (x *ListServerSummaryResponse) GetServers() []*ServerSummary {
	if x != nil {
		return x.Servers
	}
	return nil
}

type ListServersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	End   *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *ListServersRequest) Reset() {
	*x = ListServersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListServersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListServersRequest) ProtoMessage() {}

func (x *ListServersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListServersRequest.ProtoReflect.Descriptor instead.
func (*ListServersRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{6}
}

func (x *ListServersRequest) GetStart() *timestamppb.Timestamp {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *ListServersRequest) GetEnd() *timestamppb.Timestamp {
	if x != nil {
		return x.End
	}
	return nil
}

type ListServersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Servers []*ServerMetadata `protobuf:"bytes,1,rep,name=servers,proto3" json:"servers,omitempty"`
}

func (x *ListServersResponse) Reset() {
	*x = ListServersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListServersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListServersResponse) ProtoMessage() {}

func (x *ListServersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListServersResponse.ProtoReflect.Descriptor instead.
func (*ListServersResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{7}
}

func (x *ListServersResponse) GetServers() []*ServerMetadata {
	if x != nil {
		return x.Servers
	}
	return nil
}

type ServerSummary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                   string           `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type                   string           `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Connections            int32            `protobuf:"varint,3,opt,name=connections,proto3" json:"connections,omitempty"`
	RequestRate            float64          `protobuf:"fixed64,4,opt,name=request_rate,json=requestRate,proto3" json:"request_rate,omitempty"`
	ConnectionsByWaitGroup map[string]int32 `protobuf:"bytes,5,rep,name=connections_by_wait_group,json=connectionsByWaitGroup,proto3" json:"connections_by_wait_group,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (x *ServerSummary) Reset() {
	*x = ServerSummary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerSummary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerSummary) ProtoMessage() {}

func (x *ServerSummary) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_dbm_api_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerSummary.ProtoReflect.Descriptor instead.
func (*ServerSummary) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_dbm_api_proto_rawDescGZIP(), []int{8}
}

func (x *ServerSummary) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ServerSummary) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *ServerSummary) GetConnections() int32 {
	if x != nil {
		return x.Connections
	}
	return 0
}

func (x *ServerSummary) GetRequestRate() float64 {
	if x != nil {
		return x.RequestRate
	}
	return 0
}

func (x *ServerSummary) GetConnectionsByWaitGroup() map[string]int32 {
	if x != nil {
		return x.ConnectionsByWaitGroup
	}
	return nil
}

var File_database_monitoring_v1_dbm_api_proto protoreflect.FileDescriptor

var file_database_monitoring_v1_dbm_api_proto_rawDesc = []byte{
	0x0a, 0x24, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x62, 0x6d, 0x5f, 0x61, 0x70, 0x69,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x25, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x24, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x53, 0x6e, 0x61,
	0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x55, 0x0a, 0x13,
	0x47, 0x65, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3e, 0x0a, 0x08, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x44,
	0x42, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x08, 0x73, 0x6e, 0x61, 0x70, 0x73,
	0x68, 0x6f, 0x74, 0x22, 0xe4, 0x01, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x05,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x2c,
	0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x68, 0x6f, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09,
	0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x61, 0x67,
	0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a,
	0x70, 0x61, 0x67, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x9b, 0x01, 0x0a, 0x15, 0x4c,
	0x69, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x09, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61,
	0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31,
	0x2e, 0x44, 0x42, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x09, 0x73, 0x6e, 0x61,
	0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x6e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x70, 0x61, 0x67,
	0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x6f, 0x74, 0x61, 0x6c,
	0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x74, 0x6f,
	0x74, 0x61, 0x6c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x7a, 0x0a, 0x18, 0x4c, 0x69, 0x73, 0x74,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x2c, 0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x03, 0x65, 0x6e, 0x64, 0x22, 0x5c, 0x0a, 0x19, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x3f, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x25, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f,
	0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x73, 0x22, 0x74, 0x0a, 0x12, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x2c, 0x0a, 0x03, 0x65, 0x6e,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x22, 0x57, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x40, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x26, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x22, 0xc5, 0x02, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d,
	0x61, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x21, 0x0a,
	0x0c, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x0b, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x61, 0x74, 0x65,
	0x12, 0x7c, 0x0a, 0x19, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x5f,
	0x62, 0x79, 0x5f, 0x77, 0x61, 0x69, 0x74, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x05, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x41, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x79, 0x57, 0x61, 0x69, 0x74, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x16, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x42, 0x79, 0x57, 0x61, 0x69, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x1a, 0x49,
	0x0a, 0x1b, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x79, 0x57,
	0x61, 0x69, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32, 0xc0, 0x03, 0x0a, 0x06, 0x44, 0x42,
	0x4d, 0x41, 0x70, 0x69, 0x12, 0x6c, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x73, 0x12, 0x2c, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x66, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f,
	0x74, 0x12, 0x2a, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x6e,
	0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e,
	0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68,
	0x6f, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x78, 0x0a, 0x11, 0x4c, 0x69,
	0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12,
	0x30, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x31, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x66, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x73, 0x12, 0x2a, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2b, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x3b, 0x5a, 0x39,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x75, 0x69, 0x6c, 0x68,
	0x65, 0x72, 0x6d, 0x65, 0x61, 0x72, 0x70, 0x61, 0x73, 0x73, 0x6f, 0x73, 0x2f, 0x64, 0x61, 0x74,
	0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67,
	0x5f, 0x76, 0x31, 0x3b, 0x64, 0x62, 0x6d, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_database_monitoring_v1_dbm_api_proto_rawDescOnce sync.Once
	file_database_monitoring_v1_dbm_api_proto_rawDescData = file_database_monitoring_v1_dbm_api_proto_rawDesc
)

func file_database_monitoring_v1_dbm_api_proto_rawDescGZIP() []byte {
	file_database_monitoring_v1_dbm_api_proto_rawDescOnce.Do(func() {
		file_database_monitoring_v1_dbm_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_database_monitoring_v1_dbm_api_proto_rawDescData)
	})
	return file_database_monitoring_v1_dbm_api_proto_rawDescData
}

var file_database_monitoring_v1_dbm_api_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_database_monitoring_v1_dbm_api_proto_goTypes = []any{
	(*GetSnapshotRequest)(nil),        // 0: database_monitoring.v1.GetSnapshotRequest
	(*GetSnapshotResponse)(nil),       // 1: database_monitoring.v1.GetSnapshotResponse
	(*ListSnapshotsRequest)(nil),      // 2: database_monitoring.v1.ListSnapshotsRequest
	(*ListSnapshotsResponse)(nil),     // 3: database_monitoring.v1.ListSnapshotsResponse
	(*ListServerSummaryRequest)(nil),  // 4: database_monitoring.v1.ListServerSummaryRequest
	(*ListServerSummaryResponse)(nil), // 5: database_monitoring.v1.ListServerSummaryResponse
	(*ListServersRequest)(nil),        // 6: database_monitoring.v1.ListServersRequest
	(*ListServersResponse)(nil),       // 7: database_monitoring.v1.ListServersResponse
	(*ServerSummary)(nil),             // 8: database_monitoring.v1.ServerSummary
	nil,                               // 9: database_monitoring.v1.ServerSummary.ConnectionsByWaitGroupEntry
	(*DBSnapshot)(nil),                // 10: database_monitoring.v1.DBSnapshot
	(*timestamppb.Timestamp)(nil),     // 11: google.protobuf.Timestamp
	(*ServerMetadata)(nil),            // 12: database_monitoring.v1.ServerMetadata
}
var file_database_monitoring_v1_dbm_api_proto_depIdxs = []int32{
	10, // 0: database_monitoring.v1.GetSnapshotResponse.snapshot:type_name -> database_monitoring.v1.DBSnapshot
	11, // 1: database_monitoring.v1.ListSnapshotsRequest.start:type_name -> google.protobuf.Timestamp
	11, // 2: database_monitoring.v1.ListSnapshotsRequest.end:type_name -> google.protobuf.Timestamp
	10, // 3: database_monitoring.v1.ListSnapshotsResponse.snapshots:type_name -> database_monitoring.v1.DBSnapshot
	11, // 4: database_monitoring.v1.ListServerSummaryRequest.start:type_name -> google.protobuf.Timestamp
	11, // 5: database_monitoring.v1.ListServerSummaryRequest.end:type_name -> google.protobuf.Timestamp
	8,  // 6: database_monitoring.v1.ListServerSummaryResponse.servers:type_name -> database_monitoring.v1.ServerSummary
	11, // 7: database_monitoring.v1.ListServersRequest.start:type_name -> google.protobuf.Timestamp
	11, // 8: database_monitoring.v1.ListServersRequest.end:type_name -> google.protobuf.Timestamp
	12, // 9: database_monitoring.v1.ListServersResponse.servers:type_name -> database_monitoring.v1.ServerMetadata
	9,  // 10: database_monitoring.v1.ServerSummary.connections_by_wait_group:type_name -> database_monitoring.v1.ServerSummary.ConnectionsByWaitGroupEntry
	2,  // 11: database_monitoring.v1.DBMApi.ListSnapshots:input_type -> database_monitoring.v1.ListSnapshotsRequest
	0,  // 12: database_monitoring.v1.DBMApi.GetSnapshot:input_type -> database_monitoring.v1.GetSnapshotRequest
	4,  // 13: database_monitoring.v1.DBMApi.ListServerSummary:input_type -> database_monitoring.v1.ListServerSummaryRequest
	6,  // 14: database_monitoring.v1.DBMApi.ListServers:input_type -> database_monitoring.v1.ListServersRequest
	3,  // 15: database_monitoring.v1.DBMApi.ListSnapshots:output_type -> database_monitoring.v1.ListSnapshotsResponse
	1,  // 16: database_monitoring.v1.DBMApi.GetSnapshot:output_type -> database_monitoring.v1.GetSnapshotResponse
	5,  // 17: database_monitoring.v1.DBMApi.ListServerSummary:output_type -> database_monitoring.v1.ListServerSummaryResponse
	7,  // 18: database_monitoring.v1.DBMApi.ListServers:output_type -> database_monitoring.v1.ListServersResponse
	15, // [15:19] is the sub-list for method output_type
	11, // [11:15] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_database_monitoring_v1_dbm_api_proto_init() }
func file_database_monitoring_v1_dbm_api_proto_init() {
	if File_database_monitoring_v1_dbm_api_proto != nil {
		return
	}
	file_database_monitoring_v1_snapshot_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_database_monitoring_v1_dbm_api_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GetSnapshotRequest); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*GetSnapshotResponse); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ListSnapshotsRequest); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*ListSnapshotsResponse); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*ListServerSummaryRequest); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*ListServerSummaryResponse); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ListServersRequest); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*ListServersResponse); i {
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
		file_database_monitoring_v1_dbm_api_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*ServerSummary); i {
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
			RawDescriptor: file_database_monitoring_v1_dbm_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_database_monitoring_v1_dbm_api_proto_goTypes,
		DependencyIndexes: file_database_monitoring_v1_dbm_api_proto_depIdxs,
		MessageInfos:      file_database_monitoring_v1_dbm_api_proto_msgTypes,
	}.Build()
	File_database_monitoring_v1_dbm_api_proto = out.File
	file_database_monitoring_v1_dbm_api_proto_rawDesc = nil
	file_database_monitoring_v1_dbm_api_proto_goTypes = nil
	file_database_monitoring_v1_dbm_api_proto_depIdxs = nil
}
