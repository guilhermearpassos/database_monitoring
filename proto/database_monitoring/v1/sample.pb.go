// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.20.1
// source: database_monitoring/v1/sample.proto

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

type QuerySample struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status            string           `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	SqlHandle         []byte           `protobuf:"bytes,2,opt,name=sql_handle,json=sqlHandle,proto3" json:"sql_handle,omitempty"`
	Text              string           `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
	Blocked           bool             `protobuf:"varint,4,opt,name=blocked,proto3" json:"blocked,omitempty"`
	Blocker           bool             `protobuf:"varint,5,opt,name=blocker,proto3" json:"blocker,omitempty"`
	TimeElapsedMillis int64            `protobuf:"varint,10,opt,name=time_elapsed_millis,json=timeElapsedMillis,proto3" json:"time_elapsed_millis,omitempty"`
	Session           *SessionMetadata `protobuf:"bytes,6,opt,name=session,proto3" json:"session,omitempty"`
	Db                *DBMetadata      `protobuf:"bytes,7,opt,name=db,proto3" json:"db,omitempty"`
	BlockInfo         *BlockMetadata   `protobuf:"bytes,8,opt,name=block_info,json=blockInfo,proto3" json:"block_info,omitempty"`
	WaitInfo          *WaitMetadata    `protobuf:"bytes,9,opt,name=wait_info,json=waitInfo,proto3" json:"wait_info,omitempty"`
}

func (x *QuerySample) Reset() {
	*x = QuerySample{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_sample_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QuerySample) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuerySample) ProtoMessage() {}

func (x *QuerySample) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_sample_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuerySample.ProtoReflect.Descriptor instead.
func (*QuerySample) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_sample_proto_rawDescGZIP(), []int{0}
}

func (x *QuerySample) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *QuerySample) GetSqlHandle() []byte {
	if x != nil {
		return x.SqlHandle
	}
	return nil
}

func (x *QuerySample) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *QuerySample) GetBlocked() bool {
	if x != nil {
		return x.Blocked
	}
	return false
}

func (x *QuerySample) GetBlocker() bool {
	if x != nil {
		return x.Blocker
	}
	return false
}

func (x *QuerySample) GetTimeElapsedMillis() int64 {
	if x != nil {
		return x.TimeElapsedMillis
	}
	return 0
}

func (x *QuerySample) GetSession() *SessionMetadata {
	if x != nil {
		return x.Session
	}
	return nil
}

func (x *QuerySample) GetDb() *DBMetadata {
	if x != nil {
		return x.Db
	}
	return nil
}

func (x *QuerySample) GetBlockInfo() *BlockMetadata {
	if x != nil {
		return x.BlockInfo
	}
	return nil
}

func (x *QuerySample) GetWaitInfo() *WaitMetadata {
	if x != nil {
		return x.WaitInfo
	}
	return nil
}

type SessionMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SessionId        string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	LoginTime        *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=login_time,json=loginTime,proto3" json:"login_time,omitempty"`
	Host             string                 `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
	ProgramName      string                 `protobuf:"bytes,4,opt,name=program_name,json=programName,proto3" json:"program_name,omitempty"`
	LoginName        string                 `protobuf:"bytes,5,opt,name=login_name,json=loginName,proto3" json:"login_name,omitempty"`
	Status           string                 `protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
	LastRequestStart *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=last_request_start,json=lastRequestStart,proto3" json:"last_request_start,omitempty"`
	LastRequestEnd   *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=last_request_end,json=lastRequestEnd,proto3" json:"last_request_end,omitempty"`
}

func (x *SessionMetadata) Reset() {
	*x = SessionMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_sample_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SessionMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SessionMetadata) ProtoMessage() {}

func (x *SessionMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_sample_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SessionMetadata.ProtoReflect.Descriptor instead.
func (*SessionMetadata) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_sample_proto_rawDescGZIP(), []int{1}
}

func (x *SessionMetadata) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SessionMetadata) GetLoginTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LoginTime
	}
	return nil
}

func (x *SessionMetadata) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *SessionMetadata) GetProgramName() string {
	if x != nil {
		return x.ProgramName
	}
	return ""
}

func (x *SessionMetadata) GetLoginName() string {
	if x != nil {
		return x.LoginName
	}
	return ""
}

func (x *SessionMetadata) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *SessionMetadata) GetLastRequestStart() *timestamppb.Timestamp {
	if x != nil {
		return x.LastRequestStart
	}
	return nil
}

func (x *SessionMetadata) GetLastRequestEnd() *timestamppb.Timestamp {
	if x != nil {
		return x.LastRequestEnd
	}
	return nil
}

type DBMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DatabaseId   string `protobuf:"bytes,1,opt,name=database_id,json=databaseId,proto3" json:"database_id,omitempty"`
	DatabaseName string `protobuf:"bytes,2,opt,name=database_name,json=databaseName,proto3" json:"database_name,omitempty"`
}

func (x *DBMetadata) Reset() {
	*x = DBMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_sample_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DBMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DBMetadata) ProtoMessage() {}

func (x *DBMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_sample_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DBMetadata.ProtoReflect.Descriptor instead.
func (*DBMetadata) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_sample_proto_rawDescGZIP(), []int{2}
}

func (x *DBMetadata) GetDatabaseId() string {
	if x != nil {
		return x.DatabaseId
	}
	return ""
}

func (x *DBMetadata) GetDatabaseName() string {
	if x != nil {
		return x.DatabaseName
	}
	return ""
}

type BlockMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockedBy       string   `protobuf:"bytes,1,opt,name=blocked_by,json=blockedBy,proto3" json:"blocked_by,omitempty"`
	BlockedSessions []string `protobuf:"bytes,2,rep,name=blocked_sessions,json=blockedSessions,proto3" json:"blocked_sessions,omitempty"`
}

func (x *BlockMetadata) Reset() {
	*x = BlockMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_sample_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockMetadata) ProtoMessage() {}

func (x *BlockMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_sample_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockMetadata.ProtoReflect.Descriptor instead.
func (*BlockMetadata) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_sample_proto_rawDescGZIP(), []int{3}
}

func (x *BlockMetadata) GetBlockedBy() string {
	if x != nil {
		return x.BlockedBy
	}
	return ""
}

func (x *BlockMetadata) GetBlockedSessions() []string {
	if x != nil {
		return x.BlockedSessions
	}
	return nil
}

type WaitMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WaitType     string `protobuf:"bytes,1,opt,name=wait_type,json=waitType,proto3" json:"wait_type,omitempty"`
	WaitTime     int64  `protobuf:"varint,2,opt,name=wait_time,json=waitTime,proto3" json:"wait_time,omitempty"`
	LastWaitType string `protobuf:"bytes,3,opt,name=last_wait_type,json=lastWaitType,proto3" json:"last_wait_type,omitempty"`
	WaitResource string `protobuf:"bytes,4,opt,name=wait_resource,json=waitResource,proto3" json:"wait_resource,omitempty"`
}

func (x *WaitMetadata) Reset() {
	*x = WaitMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_v1_sample_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WaitMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WaitMetadata) ProtoMessage() {}

func (x *WaitMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_v1_sample_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WaitMetadata.ProtoReflect.Descriptor instead.
func (*WaitMetadata) Descriptor() ([]byte, []int) {
	return file_database_monitoring_v1_sample_proto_rawDescGZIP(), []int{4}
}

func (x *WaitMetadata) GetWaitType() string {
	if x != nil {
		return x.WaitType
	}
	return ""
}

func (x *WaitMetadata) GetWaitTime() int64 {
	if x != nil {
		return x.WaitTime
	}
	return 0
}

func (x *WaitMetadata) GetLastWaitType() string {
	if x != nil {
		return x.LastWaitType
	}
	return ""
}

func (x *WaitMetadata) GetWaitResource() string {
	if x != nil {
		return x.WaitResource
	}
	return ""
}

var File_database_monitoring_v1_sample_proto protoreflect.FileDescriptor

var file_database_monitoring_v1_sample_proto_rawDesc = []byte{
	0x0a, 0x23, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f,
	0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbc,
	0x03, 0x0a, 0x0b, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x71, 0x6c, 0x5f, 0x68, 0x61,
	0x6e, 0x64, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x71, 0x6c, 0x48,
	0x61, 0x6e, 0x64, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x12, 0x2e, 0x0a,
	0x13, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x65, 0x6c, 0x61, 0x70, 0x73, 0x65, 0x64, 0x5f, 0x6d, 0x69,
	0x6c, 0x6c, 0x69, 0x73, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x74, 0x69, 0x6d, 0x65,
	0x45, 0x6c, 0x61, 0x70, 0x73, 0x65, 0x64, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73, 0x12, 0x41, 0x0a,
	0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27,
	0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x32, 0x0a, 0x02, 0x64, 0x62, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x64,
	0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69,
	0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x52, 0x02, 0x64, 0x62, 0x12, 0x44, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x69, 0x6e,
	0x66, 0x6f, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62,
	0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76,
	0x31, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52,
	0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x41, 0x0a, 0x09, 0x77, 0x61,
	0x69, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e,
	0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x61, 0x69, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x52, 0x08, 0x77, 0x61, 0x69, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0xe9, 0x02,
	0x0a, 0x0f, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x12, 0x39, 0x0a, 0x0a, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x09, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68,
	0x6f, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12,
	0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x61, 0x6d, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x61, 0x6d, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x48, 0x0a, 0x12, 0x6c, 0x61, 0x73,
	0x74, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x10, 0x6c, 0x61, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x12, 0x44, 0x0a, 0x10, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x5f, 0x65, 0x6e, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0e, 0x6c, 0x61, 0x73, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x64, 0x22, 0x52, 0x0a, 0x0a, 0x44, 0x42, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x61, 0x74, 0x61, 0x62,
	0x61, 0x73, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x61, 0x74, 0x61,
	0x62, 0x61, 0x73, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x59, 0x0a,
	0x0d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1d,
	0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x42, 0x79, 0x12, 0x29, 0x0a,
	0x10, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x5f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64,
	0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x93, 0x01, 0x0a, 0x0c, 0x57, 0x61, 0x69,
	0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1b, 0x0a, 0x09, 0x77, 0x61, 0x69,
	0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x77, 0x61,
	0x69, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x77, 0x61, 0x69, 0x74, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x77, 0x61, 0x69, 0x74, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x77, 0x61, 0x69, 0x74,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6c, 0x61, 0x73,
	0x74, 0x57, 0x61, 0x69, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x77, 0x61, 0x69,
	0x74, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x77, 0x61, 0x69, 0x74, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x3b,
	0x5a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x75, 0x69,
	0x6c, 0x68, 0x65, 0x72, 0x6d, 0x65, 0x61, 0x72, 0x70, 0x61, 0x73, 0x73, 0x6f, 0x73, 0x2f, 0x64,
	0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69,
	0x6e, 0x67, 0x5f, 0x76, 0x31, 0x3b, 0x64, 0x62, 0x6d, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_database_monitoring_v1_sample_proto_rawDescOnce sync.Once
	file_database_monitoring_v1_sample_proto_rawDescData = file_database_monitoring_v1_sample_proto_rawDesc
)

func file_database_monitoring_v1_sample_proto_rawDescGZIP() []byte {
	file_database_monitoring_v1_sample_proto_rawDescOnce.Do(func() {
		file_database_monitoring_v1_sample_proto_rawDescData = protoimpl.X.CompressGZIP(file_database_monitoring_v1_sample_proto_rawDescData)
	})
	return file_database_monitoring_v1_sample_proto_rawDescData
}

var file_database_monitoring_v1_sample_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_database_monitoring_v1_sample_proto_goTypes = []any{
	(*QuerySample)(nil),           // 0: database_monitoring.v1.QuerySample
	(*SessionMetadata)(nil),       // 1: database_monitoring.v1.SessionMetadata
	(*DBMetadata)(nil),            // 2: database_monitoring.v1.DBMetadata
	(*BlockMetadata)(nil),         // 3: database_monitoring.v1.BlockMetadata
	(*WaitMetadata)(nil),          // 4: database_monitoring.v1.WaitMetadata
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_database_monitoring_v1_sample_proto_depIdxs = []int32{
	1, // 0: database_monitoring.v1.QuerySample.session:type_name -> database_monitoring.v1.SessionMetadata
	2, // 1: database_monitoring.v1.QuerySample.db:type_name -> database_monitoring.v1.DBMetadata
	3, // 2: database_monitoring.v1.QuerySample.block_info:type_name -> database_monitoring.v1.BlockMetadata
	4, // 3: database_monitoring.v1.QuerySample.wait_info:type_name -> database_monitoring.v1.WaitMetadata
	5, // 4: database_monitoring.v1.SessionMetadata.login_time:type_name -> google.protobuf.Timestamp
	5, // 5: database_monitoring.v1.SessionMetadata.last_request_start:type_name -> google.protobuf.Timestamp
	5, // 6: database_monitoring.v1.SessionMetadata.last_request_end:type_name -> google.protobuf.Timestamp
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_database_monitoring_v1_sample_proto_init() }
func file_database_monitoring_v1_sample_proto_init() {
	if File_database_monitoring_v1_sample_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_database_monitoring_v1_sample_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*QuerySample); i {
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
		file_database_monitoring_v1_sample_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SessionMetadata); i {
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
		file_database_monitoring_v1_sample_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*DBMetadata); i {
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
		file_database_monitoring_v1_sample_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*BlockMetadata); i {
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
		file_database_monitoring_v1_sample_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*WaitMetadata); i {
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
			RawDescriptor: file_database_monitoring_v1_sample_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_database_monitoring_v1_sample_proto_goTypes,
		DependencyIndexes: file_database_monitoring_v1_sample_proto_depIdxs,
		MessageInfos:      file_database_monitoring_v1_sample_proto_msgTypes,
	}.Build()
	File_database_monitoring_v1_sample_proto = out.File
	file_database_monitoring_v1_sample_proto_rawDesc = nil
	file_database_monitoring_v1_sample_proto_goTypes = nil
	file_database_monitoring_v1_sample_proto_depIdxs = nil
}
