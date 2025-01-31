// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package tpb

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

// TunnelClient is the client API for Tunnel service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TunnelClient interface {
	RemoteIP(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RemoteIPResponse, error)
	CreateTunnel(ctx context.Context, opts ...grpc.CallOption) (Tunnel_CreateTunnelClient, error)
	// Register a tunnel
	RegisterTunnel(ctx context.Context, in *RegisterTunnelRequest, opts ...grpc.CallOption) (*RegisterTunnelResponse, error)
	// Get brief description information of all tunnels
	GetAllRegisteredTunnel(ctx context.Context, in *GetAllRegisteredTunnelRequest, opts ...grpc.CallOption) (*GetAllRegisteredTunnelResponse, error)
	GetRegisteredTunnelDescriptionByID(ctx context.Context, in *GetRegisteredTunnelDescriptionByIDRequest, opts ...grpc.CallOption) (*RegisteredTunnel, error)
	// Random port trigger
	RequireRandomPortTrigger(ctx context.Context, in *RequireRandomPortTriggerParams, opts ...grpc.CallOption) (*RequireRandomPortTriggerResponse, error)
	QueryExistedRandomPortTrigger(ctx context.Context, in *QueryExistedRandomPortTriggerRequest, opts ...grpc.CallOption) (*QueryExistedRandomPortTriggerResponse, error)
	// Random ICMP length trigger
	QuerySpecificICMPLengthTrigger(ctx context.Context, in *QuerySpecificICMPLengthTriggerParams, opts ...grpc.CallOption) (*QuerySpecificICMPLengthTriggerResponse, error)
}

type tunnelClient struct {
	cc grpc.ClientConnInterface
}

func NewTunnelClient(cc grpc.ClientConnInterface) TunnelClient {
	return &tunnelClient{cc}
}

func (c *tunnelClient) RemoteIP(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RemoteIPResponse, error) {
	out := new(RemoteIPResponse)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/RemoteIP", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) CreateTunnel(ctx context.Context, opts ...grpc.CallOption) (Tunnel_CreateTunnelClient, error) {
	stream, err := c.cc.NewStream(ctx, &Tunnel_ServiceDesc.Streams[0], "/tpb.Tunnel/CreateTunnel", opts...)
	if err != nil {
		return nil, err
	}
	x := &tunnelCreateTunnelClient{stream}
	return x, nil
}

type Tunnel_CreateTunnelClient interface {
	Send(*TunnelInput) error
	Recv() (*TunnelOutput, error)
	grpc.ClientStream
}

type tunnelCreateTunnelClient struct {
	grpc.ClientStream
}

func (x *tunnelCreateTunnelClient) Send(m *TunnelInput) error {
	return x.ClientStream.SendMsg(m)
}

func (x *tunnelCreateTunnelClient) Recv() (*TunnelOutput, error) {
	m := new(TunnelOutput)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *tunnelClient) RegisterTunnel(ctx context.Context, in *RegisterTunnelRequest, opts ...grpc.CallOption) (*RegisterTunnelResponse, error) {
	out := new(RegisterTunnelResponse)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/RegisterTunnel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) GetAllRegisteredTunnel(ctx context.Context, in *GetAllRegisteredTunnelRequest, opts ...grpc.CallOption) (*GetAllRegisteredTunnelResponse, error) {
	out := new(GetAllRegisteredTunnelResponse)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/GetAllRegisteredTunnel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) GetRegisteredTunnelDescriptionByID(ctx context.Context, in *GetRegisteredTunnelDescriptionByIDRequest, opts ...grpc.CallOption) (*RegisteredTunnel, error) {
	out := new(RegisteredTunnel)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/GetRegisteredTunnelDescriptionByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) RequireRandomPortTrigger(ctx context.Context, in *RequireRandomPortTriggerParams, opts ...grpc.CallOption) (*RequireRandomPortTriggerResponse, error) {
	out := new(RequireRandomPortTriggerResponse)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/RequireRandomPortTrigger", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) QueryExistedRandomPortTrigger(ctx context.Context, in *QueryExistedRandomPortTriggerRequest, opts ...grpc.CallOption) (*QueryExistedRandomPortTriggerResponse, error) {
	out := new(QueryExistedRandomPortTriggerResponse)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/QueryExistedRandomPortTrigger", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) QuerySpecificICMPLengthTrigger(ctx context.Context, in *QuerySpecificICMPLengthTriggerParams, opts ...grpc.CallOption) (*QuerySpecificICMPLengthTriggerResponse, error) {
	out := new(QuerySpecificICMPLengthTriggerResponse)
	err := c.cc.Invoke(ctx, "/tpb.Tunnel/QuerySpecificICMPLengthTrigger", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TunnelServer is the server API for Tunnel service.
// All implementations must embed UnimplementedTunnelServer
// for forward compatibility
type TunnelServer interface {
	RemoteIP(context.Context, *Empty) (*RemoteIPResponse, error)
	CreateTunnel(Tunnel_CreateTunnelServer) error
	// Register a tunnel
	RegisterTunnel(context.Context, *RegisterTunnelRequest) (*RegisterTunnelResponse, error)
	// Get brief description information of all tunnels
	GetAllRegisteredTunnel(context.Context, *GetAllRegisteredTunnelRequest) (*GetAllRegisteredTunnelResponse, error)
	GetRegisteredTunnelDescriptionByID(context.Context, *GetRegisteredTunnelDescriptionByIDRequest) (*RegisteredTunnel, error)
	// Random port trigger
	RequireRandomPortTrigger(context.Context, *RequireRandomPortTriggerParams) (*RequireRandomPortTriggerResponse, error)
	QueryExistedRandomPortTrigger(context.Context, *QueryExistedRandomPortTriggerRequest) (*QueryExistedRandomPortTriggerResponse, error)
	// Random ICMP length trigger
	QuerySpecificICMPLengthTrigger(context.Context, *QuerySpecificICMPLengthTriggerParams) (*QuerySpecificICMPLengthTriggerResponse, error)
	mustEmbedUnimplementedTunnelServer()
}

// UnimplementedTunnelServer must be embedded to have forward compatible implementations.
type UnimplementedTunnelServer struct {
}

func (UnimplementedTunnelServer) RemoteIP(context.Context, *Empty) (*RemoteIPResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoteIP not implemented")
}
func (UnimplementedTunnelServer) CreateTunnel(Tunnel_CreateTunnelServer) error {
	return status.Errorf(codes.Unimplemented, "method CreateTunnel not implemented")
}
func (UnimplementedTunnelServer) RegisterTunnel(context.Context, *RegisterTunnelRequest) (*RegisterTunnelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterTunnel not implemented")
}
func (UnimplementedTunnelServer) GetAllRegisteredTunnel(context.Context, *GetAllRegisteredTunnelRequest) (*GetAllRegisteredTunnelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllRegisteredTunnel not implemented")
}
func (UnimplementedTunnelServer) GetRegisteredTunnelDescriptionByID(context.Context, *GetRegisteredTunnelDescriptionByIDRequest) (*RegisteredTunnel, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRegisteredTunnelDescriptionByID not implemented")
}
func (UnimplementedTunnelServer) RequireRandomPortTrigger(context.Context, *RequireRandomPortTriggerParams) (*RequireRandomPortTriggerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequireRandomPortTrigger not implemented")
}
func (UnimplementedTunnelServer) QueryExistedRandomPortTrigger(context.Context, *QueryExistedRandomPortTriggerRequest) (*QueryExistedRandomPortTriggerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryExistedRandomPortTrigger not implemented")
}
func (UnimplementedTunnelServer) QuerySpecificICMPLengthTrigger(context.Context, *QuerySpecificICMPLengthTriggerParams) (*QuerySpecificICMPLengthTriggerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QuerySpecificICMPLengthTrigger not implemented")
}
func (UnimplementedTunnelServer) mustEmbedUnimplementedTunnelServer() {}

// UnsafeTunnelServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TunnelServer will
// result in compilation errors.
type UnsafeTunnelServer interface {
	mustEmbedUnimplementedTunnelServer()
}

func RegisterTunnelServer(s grpc.ServiceRegistrar, srv TunnelServer) {
	s.RegisterService(&Tunnel_ServiceDesc, srv)
}

func _Tunnel_RemoteIP_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).RemoteIP(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/RemoteIP",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).RemoteIP(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_CreateTunnel_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TunnelServer).CreateTunnel(&tunnelCreateTunnelServer{stream})
}

type Tunnel_CreateTunnelServer interface {
	Send(*TunnelOutput) error
	Recv() (*TunnelInput, error)
	grpc.ServerStream
}

type tunnelCreateTunnelServer struct {
	grpc.ServerStream
}

func (x *tunnelCreateTunnelServer) Send(m *TunnelOutput) error {
	return x.ServerStream.SendMsg(m)
}

func (x *tunnelCreateTunnelServer) Recv() (*TunnelInput, error) {
	m := new(TunnelInput)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Tunnel_RegisterTunnel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterTunnelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).RegisterTunnel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/RegisterTunnel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).RegisterTunnel(ctx, req.(*RegisterTunnelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_GetAllRegisteredTunnel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllRegisteredTunnelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).GetAllRegisteredTunnel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/GetAllRegisteredTunnel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).GetAllRegisteredTunnel(ctx, req.(*GetAllRegisteredTunnelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_GetRegisteredTunnelDescriptionByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRegisteredTunnelDescriptionByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).GetRegisteredTunnelDescriptionByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/GetRegisteredTunnelDescriptionByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).GetRegisteredTunnelDescriptionByID(ctx, req.(*GetRegisteredTunnelDescriptionByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_RequireRandomPortTrigger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequireRandomPortTriggerParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).RequireRandomPortTrigger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/RequireRandomPortTrigger",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).RequireRandomPortTrigger(ctx, req.(*RequireRandomPortTriggerParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_QueryExistedRandomPortTrigger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryExistedRandomPortTriggerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).QueryExistedRandomPortTrigger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/QueryExistedRandomPortTrigger",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).QueryExistedRandomPortTrigger(ctx, req.(*QueryExistedRandomPortTriggerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_QuerySpecificICMPLengthTrigger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySpecificICMPLengthTriggerParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).QuerySpecificICMPLengthTrigger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.Tunnel/QuerySpecificICMPLengthTrigger",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).QuerySpecificICMPLengthTrigger(ctx, req.(*QuerySpecificICMPLengthTriggerParams))
	}
	return interceptor(ctx, in, info, handler)
}

// Tunnel_ServiceDesc is the grpc.ServiceDesc for Tunnel service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tunnel_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tpb.Tunnel",
	HandlerType: (*TunnelServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RemoteIP",
			Handler:    _Tunnel_RemoteIP_Handler,
		},
		{
			MethodName: "RegisterTunnel",
			Handler:    _Tunnel_RegisterTunnel_Handler,
		},
		{
			MethodName: "GetAllRegisteredTunnel",
			Handler:    _Tunnel_GetAllRegisteredTunnel_Handler,
		},
		{
			MethodName: "GetRegisteredTunnelDescriptionByID",
			Handler:    _Tunnel_GetRegisteredTunnelDescriptionByID_Handler,
		},
		{
			MethodName: "RequireRandomPortTrigger",
			Handler:    _Tunnel_RequireRandomPortTrigger_Handler,
		},
		{
			MethodName: "QueryExistedRandomPortTrigger",
			Handler:    _Tunnel_QueryExistedRandomPortTrigger_Handler,
		},
		{
			MethodName: "QuerySpecificICMPLengthTrigger",
			Handler:    _Tunnel_QuerySpecificICMPLengthTrigger_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CreateTunnel",
			Handler:       _Tunnel_CreateTunnel_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "tunnel.proto",
}

// DNSLogClient is the client API for DNSLog service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DNSLogClient interface {
	RequireDomain(ctx context.Context, in *RequireDomainParams, opts ...grpc.CallOption) (*RequireDomainResponse, error)
	QueryExistedDNSLog(ctx context.Context, in *QueryExistedDNSLogParams, opts ...grpc.CallOption) (*QueryExistedDNSLogResponse, error)
}

type dNSLogClient struct {
	cc grpc.ClientConnInterface
}

func NewDNSLogClient(cc grpc.ClientConnInterface) DNSLogClient {
	return &dNSLogClient{cc}
}

func (c *dNSLogClient) RequireDomain(ctx context.Context, in *RequireDomainParams, opts ...grpc.CallOption) (*RequireDomainResponse, error) {
	out := new(RequireDomainResponse)
	err := c.cc.Invoke(ctx, "/tpb.DNSLog/RequireDomain", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dNSLogClient) QueryExistedDNSLog(ctx context.Context, in *QueryExistedDNSLogParams, opts ...grpc.CallOption) (*QueryExistedDNSLogResponse, error) {
	out := new(QueryExistedDNSLogResponse)
	err := c.cc.Invoke(ctx, "/tpb.DNSLog/QueryExistedDNSLog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DNSLogServer is the server API for DNSLog service.
// All implementations must embed UnimplementedDNSLogServer
// for forward compatibility
type DNSLogServer interface {
	RequireDomain(context.Context, *RequireDomainParams) (*RequireDomainResponse, error)
	QueryExistedDNSLog(context.Context, *QueryExistedDNSLogParams) (*QueryExistedDNSLogResponse, error)
	mustEmbedUnimplementedDNSLogServer()
}

// UnimplementedDNSLogServer must be embedded to have forward compatible implementations.
type UnimplementedDNSLogServer struct {
}

func (UnimplementedDNSLogServer) RequireDomain(context.Context, *RequireDomainParams) (*RequireDomainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequireDomain not implemented")
}
func (UnimplementedDNSLogServer) QueryExistedDNSLog(context.Context, *QueryExistedDNSLogParams) (*QueryExistedDNSLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryExistedDNSLog not implemented")
}
func (UnimplementedDNSLogServer) mustEmbedUnimplementedDNSLogServer() {}

// UnsafeDNSLogServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DNSLogServer will
// result in compilation errors.
type UnsafeDNSLogServer interface {
	mustEmbedUnimplementedDNSLogServer()
}

func RegisterDNSLogServer(s grpc.ServiceRegistrar, srv DNSLogServer) {
	s.RegisterService(&DNSLog_ServiceDesc, srv)
}

func _DNSLog_RequireDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequireDomainParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DNSLogServer).RequireDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.DNSLog/RequireDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DNSLogServer).RequireDomain(ctx, req.(*RequireDomainParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _DNSLog_QueryExistedDNSLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryExistedDNSLogParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DNSLogServer).QueryExistedDNSLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tpb.DNSLog/QueryExistedDNSLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DNSLogServer).QueryExistedDNSLog(ctx, req.(*QueryExistedDNSLogParams))
	}
	return interceptor(ctx, in, info, handler)
}

// DNSLog_ServiceDesc is the grpc.ServiceDesc for DNSLog service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DNSLog_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tpb.DNSLog",
	HandlerType: (*DNSLogServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequireDomain",
			Handler:    _DNSLog_RequireDomain_Handler,
		},
		{
			MethodName: "QueryExistedDNSLog",
			Handler:    _DNSLog_QueryExistedDNSLog_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tunnel.proto",
}
