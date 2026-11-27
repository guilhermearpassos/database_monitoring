// Code generated from database_monitoring/ingestor/v2/deadlock_api.proto.

package ingestorv2

import (
	reflect "reflect"
	sync "sync"

	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type IngestDeadlocksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Server    *dbmv1.ServerMetadata   `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
	Deadlocks []*dbmv1.DeadlockReport `protobuf:"bytes,2,rep,name=deadlocks,proto3" json:"deadlocks,omitempty"`
}

func (x *IngestDeadlocksRequest) Reset() {
	*x = IngestDeadlocksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestDeadlocksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestDeadlocksRequest) ProtoMessage() {}

func (x *IngestDeadlocksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*IngestDeadlocksRequest) Descriptor() ([]byte, []int) {
	return file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescGZIP(), []int{0}
}

func (x *IngestDeadlocksRequest) GetServer() *dbmv1.ServerMetadata {
	if x != nil {
		return x.Server
	}
	return nil
}

func (x *IngestDeadlocksRequest) GetDeadlocks() []*dbmv1.DeadlockReport {
	if x != nil {
		return x.Deadlocks
	}
	return nil
}

type IngestDeadlocksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StoredCount int32 `protobuf:"varint,1,opt,name=stored_count,json=storedCount,proto3" json:"stored_count,omitempty"`
}

func (x *IngestDeadlocksResponse) Reset() {
	*x = IngestDeadlocksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestDeadlocksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestDeadlocksResponse) ProtoMessage() {}

func (x *IngestDeadlocksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*IngestDeadlocksResponse) Descriptor() ([]byte, []int) {
	return file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescGZIP(), []int{1}
}

func (x *IngestDeadlocksResponse) GetStoredCount() int32 {
	if x != nil {
		return x.StoredCount
	}
	return 0
}

var File_database_monitoring_ingestor_v2_deadlock_api_proto protoreflect.FileDescriptor

var file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDesc = []byte{
	0x0a, 0x32, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x2f, 0x76,
	0x32, 0x2f, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x32, 0x1a, 0x25, 0x64, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e,
	0x67, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x25, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f,
	0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x61, 0x64,
	0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9e, 0x01, 0x0a, 0x16, 0x49,
	0x6e, 0x67, 0x65, 0x73, 0x74, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3e, 0x0a, 0x06, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x06, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x44, 0x0a, 0x09, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63,
	0x6b, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x62,
	0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76,
	0x31, 0x2e, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x52, 0x09, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x22, 0x3c, 0x0a, 0x17, 0x49,
	0x6e, 0x67, 0x65, 0x73, 0x74, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x64,
	0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x42, 0x63, 0x5a, 0x61, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x75, 0x69, 0x6c, 0x68, 0x65, 0x72, 0x6d,
	0x65, 0x61, 0x72, 0x70, 0x61, 0x73, 0x73, 0x6f, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61,
	0x73, 0x65, 0x2d, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x6d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x6f, 0x72,
	0x2f, 0x76, 0x32, 0x3b, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x76, 0x32, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescOnce sync.Once
	file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescData = file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDesc
)

func file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescGZIP() []byte {
	file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescOnce.Do(func() {
		file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescData)
	})
	return file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDescData
}

var file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_database_monitoring_ingestor_v2_deadlock_api_proto_goTypes = []any{
	(*IngestDeadlocksRequest)(nil),  // 0: database_monitoring.v2.IngestDeadlocksRequest
	(*IngestDeadlocksResponse)(nil), // 1: database_monitoring.v2.IngestDeadlocksResponse
	(*dbmv1.ServerMetadata)(nil),    // 2: database_monitoring.v1.ServerMetadata
	(*dbmv1.DeadlockReport)(nil),    // 3: database_monitoring.v1.DeadlockReport
}
var file_database_monitoring_ingestor_v2_deadlock_api_proto_depIdxs = []int32{
	2, // 0: database_monitoring.v2.IngestDeadlocksRequest.server:type_name -> database_monitoring.v1.ServerMetadata
	3, // 1: database_monitoring.v2.IngestDeadlocksRequest.deadlocks:type_name -> database_monitoring.v1.DeadlockReport
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_database_monitoring_ingestor_v2_deadlock_api_proto_init() }
func file_database_monitoring_ingestor_v2_deadlock_api_proto_init() {
	if File_database_monitoring_ingestor_v2_deadlock_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*IngestDeadlocksRequest); i {
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
		file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*IngestDeadlocksResponse); i {
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
			RawDescriptor: file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_database_monitoring_ingestor_v2_deadlock_api_proto_goTypes,
		DependencyIndexes: file_database_monitoring_ingestor_v2_deadlock_api_proto_depIdxs,
		MessageInfos:      file_database_monitoring_ingestor_v2_deadlock_api_proto_msgTypes,
	}.Build()
	File_database_monitoring_ingestor_v2_deadlock_api_proto = out.File
	file_database_monitoring_ingestor_v2_deadlock_api_proto_rawDesc = nil
	file_database_monitoring_ingestor_v2_deadlock_api_proto_goTypes = nil
	file_database_monitoring_ingestor_v2_deadlock_api_proto_depIdxs = nil
}
