// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// HandleCliClient is the client API for HandleCli service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HandleCliClient interface {
	HandleCliRequest(ctx context.Context, in *CliRequest, opts ...grpc.CallOption) (*ResponseStatus, error)
}

type handleCliClient struct {
	cc grpc.ClientConnInterface
}

func NewHandleCliClient(cc grpc.ClientConnInterface) HandleCliClient {
	return &handleCliClient{cc}
}

func (c *handleCliClient) HandleCliRequest(ctx context.Context, in *CliRequest, opts ...grpc.CallOption) (*ResponseStatus, error) {
	out := new(ResponseStatus)
	err := c.cc.Invoke(ctx, "/protobuf.HandleCli/HandleCliRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HandleCliServer is the server API for HandleCli service.
// All implementations must embed UnimplementedHandleCliServer
// for forward compatibility
type HandleCliServer interface {
	HandleCliRequest(context.Context, *CliRequest) (*ResponseStatus, error)
	mustEmbedUnimplementedHandleCliServer()
}

// UnimplementedHandleCliServer must be embedded to have forward compatible implementations.
type UnimplementedHandleCliServer struct {
}

func (UnimplementedHandleCliServer) HandleCliRequest(context.Context, *CliRequest) (*ResponseStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleCliRequest not implemented")
}
func (UnimplementedHandleCliServer) mustEmbedUnimplementedHandleCliServer() {}

// UnsafeHandleCliServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HandleCliServer will
// result in compilation errors.
type UnsafeHandleCliServer interface {
	mustEmbedUnimplementedHandleCliServer()
}

func RegisterHandleCliServer(s grpc.ServiceRegistrar, srv HandleCliServer) {
	s.RegisterService(&HandleCli_ServiceDesc, srv)
}

func _HandleCli_HandleCliRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CliRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandleCliServer).HandleCliRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.HandleCli/HandleCliRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandleCliServer).HandleCliRequest(ctx, req.(*CliRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// HandleCli_ServiceDesc is the grpc.ServiceDesc for HandleCli service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HandleCli_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protobuf.HandleCli",
	HandlerType: (*HandleCliServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleCliRequest",
			Handler:    _HandleCli_HandleCliRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "clihandler.proto",
}