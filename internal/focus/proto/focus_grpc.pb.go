// Copyright (c) 2022 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: focus.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Focus_Raise_FullMethodName   = "/focus.Focus/Raise"
	Focus_Version_FullMethodName = "/focus.Focus/Version"
)

// FocusClient is the client API for Focus service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FocusClient interface {
	Raise(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error)
}

type focusClient struct {
	cc grpc.ClientConnInterface
}

func NewFocusClient(cc grpc.ClientConnInterface) FocusClient {
	return &focusClient{cc}
}

func (c *focusClient) Raise(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Focus_Raise_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *focusClient) Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error) {
	out := new(VersionResponse)
	err := c.cc.Invoke(ctx, Focus_Version_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FocusServer is the server API for Focus service.
// All implementations must embed UnimplementedFocusServer
// for forward compatibility
type FocusServer interface {
	Raise(context.Context, *wrapperspb.StringValue) (*emptypb.Empty, error)
	Version(context.Context, *emptypb.Empty) (*VersionResponse, error)
	mustEmbedUnimplementedFocusServer()
}

// UnimplementedFocusServer must be embedded to have forward compatible implementations.
type UnimplementedFocusServer struct {
}

func (UnimplementedFocusServer) Raise(context.Context, *wrapperspb.StringValue) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Raise not implemented")
}
func (UnimplementedFocusServer) Version(context.Context, *emptypb.Empty) (*VersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedFocusServer) mustEmbedUnimplementedFocusServer() {}

// UnsafeFocusServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FocusServer will
// result in compilation errors.
type UnsafeFocusServer interface {
	mustEmbedUnimplementedFocusServer()
}

func RegisterFocusServer(s grpc.ServiceRegistrar, srv FocusServer) {
	s.RegisterService(&Focus_ServiceDesc, srv)
}

func _Focus_Raise_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(wrapperspb.StringValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FocusServer).Raise(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Focus_Raise_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FocusServer).Raise(ctx, req.(*wrapperspb.StringValue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Focus_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FocusServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Focus_Version_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FocusServer).Version(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Focus_ServiceDesc is the grpc.ServiceDesc for Focus service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Focus_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "focus.Focus",
	HandlerType: (*FocusServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Raise",
			Handler:    _Focus_Raise_Handler,
		},
		{
			MethodName: "Version",
			Handler:    _Focus_Version_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "focus.proto",
}
