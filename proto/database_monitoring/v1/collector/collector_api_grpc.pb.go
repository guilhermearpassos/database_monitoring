// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.20.1
// source: database_monitoring/v1/collector/collector_api.proto

package collectorv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	IngestionService_RegisterAgent_FullMethodName  = "/database_monitoring.v1.IngestionService/RegisterAgent"
	IngestionService_IngestMetrics_FullMethodName  = "/database_monitoring.v1.IngestionService/IngestMetrics"
	IngestionService_IngestSnapshot_FullMethodName = "/database_monitoring.v1.IngestionService/IngestSnapshot"
)

// IngestionServiceClient is the client API for IngestionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IngestionServiceClient interface {
	RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error)
	IngestMetrics(ctx context.Context, in *DatabaseMetrics, opts ...grpc.CallOption) (*IngestMetricsResponse, error)
	IngestSnapshot(ctx context.Context, in *IngestSnapshotRequest, opts ...grpc.CallOption) (*IngestSnapshotResponse, error)
}

type ingestionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIngestionServiceClient(cc grpc.ClientConnInterface) IngestionServiceClient {
	return &ingestionServiceClient{cc}
}

func (c *ingestionServiceClient) RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterAgentResponse)
	err := c.cc.Invoke(ctx, IngestionService_RegisterAgent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ingestionServiceClient) IngestMetrics(ctx context.Context, in *DatabaseMetrics, opts ...grpc.CallOption) (*IngestMetricsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IngestMetricsResponse)
	err := c.cc.Invoke(ctx, IngestionService_IngestMetrics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ingestionServiceClient) IngestSnapshot(ctx context.Context, in *IngestSnapshotRequest, opts ...grpc.CallOption) (*IngestSnapshotResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IngestSnapshotResponse)
	err := c.cc.Invoke(ctx, IngestionService_IngestSnapshot_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IngestionServiceServer is the server API for IngestionService service.
// All implementations must embed UnimplementedIngestionServiceServer
// for forward compatibility.
type IngestionServiceServer interface {
	RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error)
	IngestMetrics(context.Context, *DatabaseMetrics) (*IngestMetricsResponse, error)
	IngestSnapshot(context.Context, *IngestSnapshotRequest) (*IngestSnapshotResponse, error)
	mustEmbedUnimplementedIngestionServiceServer()
}

// UnimplementedIngestionServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedIngestionServiceServer struct{}

func (UnimplementedIngestionServiceServer) RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterAgent not implemented")
}
func (UnimplementedIngestionServiceServer) IngestMetrics(context.Context, *DatabaseMetrics) (*IngestMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IngestMetrics not implemented")
}
func (UnimplementedIngestionServiceServer) IngestSnapshot(context.Context, *IngestSnapshotRequest) (*IngestSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IngestSnapshot not implemented")
}
func (UnimplementedIngestionServiceServer) mustEmbedUnimplementedIngestionServiceServer() {}
func (UnimplementedIngestionServiceServer) testEmbeddedByValue()                          {}

// UnsafeIngestionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IngestionServiceServer will
// result in compilation errors.
type UnsafeIngestionServiceServer interface {
	mustEmbedUnimplementedIngestionServiceServer()
}

func RegisterIngestionServiceServer(s grpc.ServiceRegistrar, srv IngestionServiceServer) {
	// If the following call pancis, it indicates UnimplementedIngestionServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&IngestionService_ServiceDesc, srv)
}

func _IngestionService_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IngestionServiceServer).RegisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IngestionService_RegisterAgent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IngestionServiceServer).RegisterAgent(ctx, req.(*RegisterAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IngestionService_IngestMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DatabaseMetrics)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IngestionServiceServer).IngestMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IngestionService_IngestMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IngestionServiceServer).IngestMetrics(ctx, req.(*DatabaseMetrics))
	}
	return interceptor(ctx, in, info, handler)
}

func _IngestionService_IngestSnapshot_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IngestSnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IngestionServiceServer).IngestSnapshot(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IngestionService_IngestSnapshot_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IngestionServiceServer).IngestSnapshot(ctx, req.(*IngestSnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IngestionService_ServiceDesc is the grpc.ServiceDesc for IngestionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IngestionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "database_monitoring.v1.IngestionService",
	HandlerType: (*IngestionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterAgent",
			Handler:    _IngestionService_RegisterAgent_Handler,
		},
		{
			MethodName: "IngestMetrics",
			Handler:    _IngestionService_IngestMetrics_Handler,
		},
		{
			MethodName: "IngestSnapshot",
			Handler:    _IngestionService_IngestSnapshot_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "database_monitoring/v1/collector/collector_api.proto",
}
