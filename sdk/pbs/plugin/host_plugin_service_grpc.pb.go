// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package plugin

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

// HostPluginServiceClient is the client API for HostPluginService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HostPluginServiceClient interface {
	// OnCreateCatalog is a hook that runs when a host catalog is
	// created.
	OnCreateCatalog(ctx context.Context, in *OnCreateCatalogRequest, opts ...grpc.CallOption) (*OnCreateCatalogResponse, error)
	// OnUpdateCatalog is a hook that runs when a host catalog is
	// updated.
	OnUpdateCatalog(ctx context.Context, in *OnUpdateCatalogRequest, opts ...grpc.CallOption) (*OnUpdateCatalogResponse, error)
	// OnDeleteCatalog is a hook that runs when a host catalog is
	// deleted.
	OnDeleteCatalog(ctx context.Context, in *OnDeleteCatalogRequest, opts ...grpc.CallOption) (*OnDeleteCatalogResponse, error)
	// OnCreateSet is a hook that runs when a host set is created.
	OnCreateSet(ctx context.Context, in *OnCreateSetRequest, opts ...grpc.CallOption) (*OnCreateSetResponse, error)
	// OnUpdateSet is a hook that runs when a host set is updated.
	OnUpdateSet(ctx context.Context, in *OnUpdateSetRequest, opts ...grpc.CallOption) (*OnUpdateSetResponse, error)
	// OnDeleteSet is a hook that runs when a host set is deleted.
	OnDeleteSet(ctx context.Context, in *OnDeleteSetRequest, opts ...grpc.CallOption) (*OnDeleteSetResponse, error)
	// ListHosts looks up all the hosts in the provided host sets.
	ListHosts(ctx context.Context, in *ListHostsRequest, opts ...grpc.CallOption) (*ListHostsResponse, error)
}

type hostPluginServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHostPluginServiceClient(cc grpc.ClientConnInterface) HostPluginServiceClient {
	return &hostPluginServiceClient{cc}
}

func (c *hostPluginServiceClient) OnCreateCatalog(ctx context.Context, in *OnCreateCatalogRequest, opts ...grpc.CallOption) (*OnCreateCatalogResponse, error) {
	out := new(OnCreateCatalogResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/OnCreateCatalog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostPluginServiceClient) OnUpdateCatalog(ctx context.Context, in *OnUpdateCatalogRequest, opts ...grpc.CallOption) (*OnUpdateCatalogResponse, error) {
	out := new(OnUpdateCatalogResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/OnUpdateCatalog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostPluginServiceClient) OnDeleteCatalog(ctx context.Context, in *OnDeleteCatalogRequest, opts ...grpc.CallOption) (*OnDeleteCatalogResponse, error) {
	out := new(OnDeleteCatalogResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/OnDeleteCatalog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostPluginServiceClient) OnCreateSet(ctx context.Context, in *OnCreateSetRequest, opts ...grpc.CallOption) (*OnCreateSetResponse, error) {
	out := new(OnCreateSetResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/OnCreateSet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostPluginServiceClient) OnUpdateSet(ctx context.Context, in *OnUpdateSetRequest, opts ...grpc.CallOption) (*OnUpdateSetResponse, error) {
	out := new(OnUpdateSetResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/OnUpdateSet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostPluginServiceClient) OnDeleteSet(ctx context.Context, in *OnDeleteSetRequest, opts ...grpc.CallOption) (*OnDeleteSetResponse, error) {
	out := new(OnDeleteSetResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/OnDeleteSet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostPluginServiceClient) ListHosts(ctx context.Context, in *ListHostsRequest, opts ...grpc.CallOption) (*ListHostsResponse, error) {
	out := new(ListHostsResponse)
	err := c.cc.Invoke(ctx, "/plugin.v1.HostPluginService/ListHosts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HostPluginServiceServer is the server API for HostPluginService service.
// All implementations must embed UnimplementedHostPluginServiceServer
// for forward compatibility
type HostPluginServiceServer interface {
	// OnCreateCatalog is a hook that runs when a host catalog is
	// created.
	OnCreateCatalog(context.Context, *OnCreateCatalogRequest) (*OnCreateCatalogResponse, error)
	// OnUpdateCatalog is a hook that runs when a host catalog is
	// updated.
	OnUpdateCatalog(context.Context, *OnUpdateCatalogRequest) (*OnUpdateCatalogResponse, error)
	// OnDeleteCatalog is a hook that runs when a host catalog is
	// deleted.
	OnDeleteCatalog(context.Context, *OnDeleteCatalogRequest) (*OnDeleteCatalogResponse, error)
	// OnCreateSet is a hook that runs when a host set is created.
	OnCreateSet(context.Context, *OnCreateSetRequest) (*OnCreateSetResponse, error)
	// OnUpdateSet is a hook that runs when a host set is updated.
	OnUpdateSet(context.Context, *OnUpdateSetRequest) (*OnUpdateSetResponse, error)
	// OnDeleteSet is a hook that runs when a host set is deleted.
	OnDeleteSet(context.Context, *OnDeleteSetRequest) (*OnDeleteSetResponse, error)
	// ListHosts looks up all the hosts in the provided host sets.
	ListHosts(context.Context, *ListHostsRequest) (*ListHostsResponse, error)
	mustEmbedUnimplementedHostPluginServiceServer()
}

// UnimplementedHostPluginServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHostPluginServiceServer struct {
}

func (UnimplementedHostPluginServiceServer) OnCreateCatalog(context.Context, *OnCreateCatalogRequest) (*OnCreateCatalogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnCreateCatalog not implemented")
}
func (UnimplementedHostPluginServiceServer) OnUpdateCatalog(context.Context, *OnUpdateCatalogRequest) (*OnUpdateCatalogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnUpdateCatalog not implemented")
}
func (UnimplementedHostPluginServiceServer) OnDeleteCatalog(context.Context, *OnDeleteCatalogRequest) (*OnDeleteCatalogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnDeleteCatalog not implemented")
}
func (UnimplementedHostPluginServiceServer) OnCreateSet(context.Context, *OnCreateSetRequest) (*OnCreateSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnCreateSet not implemented")
}
func (UnimplementedHostPluginServiceServer) OnUpdateSet(context.Context, *OnUpdateSetRequest) (*OnUpdateSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnUpdateSet not implemented")
}
func (UnimplementedHostPluginServiceServer) OnDeleteSet(context.Context, *OnDeleteSetRequest) (*OnDeleteSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnDeleteSet not implemented")
}
func (UnimplementedHostPluginServiceServer) ListHosts(context.Context, *ListHostsRequest) (*ListHostsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListHosts not implemented")
}
func (UnimplementedHostPluginServiceServer) mustEmbedUnimplementedHostPluginServiceServer() {}

// UnsafeHostPluginServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HostPluginServiceServer will
// result in compilation errors.
type UnsafeHostPluginServiceServer interface {
	mustEmbedUnimplementedHostPluginServiceServer()
}

func RegisterHostPluginServiceServer(s grpc.ServiceRegistrar, srv HostPluginServiceServer) {
	s.RegisterService(&HostPluginService_ServiceDesc, srv)
}

func _HostPluginService_OnCreateCatalog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnCreateCatalogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).OnCreateCatalog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/OnCreateCatalog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).OnCreateCatalog(ctx, req.(*OnCreateCatalogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostPluginService_OnUpdateCatalog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnUpdateCatalogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).OnUpdateCatalog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/OnUpdateCatalog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).OnUpdateCatalog(ctx, req.(*OnUpdateCatalogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostPluginService_OnDeleteCatalog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnDeleteCatalogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).OnDeleteCatalog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/OnDeleteCatalog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).OnDeleteCatalog(ctx, req.(*OnDeleteCatalogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostPluginService_OnCreateSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnCreateSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).OnCreateSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/OnCreateSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).OnCreateSet(ctx, req.(*OnCreateSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostPluginService_OnUpdateSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnUpdateSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).OnUpdateSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/OnUpdateSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).OnUpdateSet(ctx, req.(*OnUpdateSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostPluginService_OnDeleteSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnDeleteSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).OnDeleteSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/OnDeleteSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).OnDeleteSet(ctx, req.(*OnDeleteSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostPluginService_ListHosts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListHostsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostPluginServiceServer).ListHosts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.HostPluginService/ListHosts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostPluginServiceServer).ListHosts(ctx, req.(*ListHostsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// HostPluginService_ServiceDesc is the grpc.ServiceDesc for HostPluginService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HostPluginService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "plugin.v1.HostPluginService",
	HandlerType: (*HostPluginServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OnCreateCatalog",
			Handler:    _HostPluginService_OnCreateCatalog_Handler,
		},
		{
			MethodName: "OnUpdateCatalog",
			Handler:    _HostPluginService_OnUpdateCatalog_Handler,
		},
		{
			MethodName: "OnDeleteCatalog",
			Handler:    _HostPluginService_OnDeleteCatalog_Handler,
		},
		{
			MethodName: "OnCreateSet",
			Handler:    _HostPluginService_OnCreateSet_Handler,
		},
		{
			MethodName: "OnUpdateSet",
			Handler:    _HostPluginService_OnUpdateSet_Handler,
		},
		{
			MethodName: "OnDeleteSet",
			Handler:    _HostPluginService_OnDeleteSet_Handler,
		},
		{
			MethodName: "ListHosts",
			Handler:    _HostPluginService_ListHosts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plugin/v1/host_plugin_service.proto",
}
