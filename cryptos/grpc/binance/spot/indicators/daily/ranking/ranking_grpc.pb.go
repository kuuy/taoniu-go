// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: binance/spot/indicators/daily/ranking/ranking.proto

package ranking

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

// RankingClient is the client API for Ranking service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RankingClient interface {
	Pagenate(ctx context.Context, in *PagenateRequest, opts ...grpc.CallOption) (*PagenateReply, error)
}

type rankingClient struct {
	cc grpc.ClientConnInterface
}

func NewRankingClient(cc grpc.ClientConnInterface) RankingClient {
	return &rankingClient{cc}
}

func (c *rankingClient) Pagenate(ctx context.Context, in *PagenateRequest, opts ...grpc.CallOption) (*PagenateReply, error) {
	out := new(PagenateReply)
	err := c.cc.Invoke(ctx, "/taoniu.local.cryptos.grpc.binance.spot.indicators.daily.ranking.Ranking/Pagenate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RankingServer is the server API for Ranking service.
// All implementations must embed UnimplementedRankingServer
// for forward compatibility
type RankingServer interface {
	Pagenate(context.Context, *PagenateRequest) (*PagenateReply, error)
	mustEmbedUnimplementedRankingServer()
}

// UnimplementedRankingServer must be embedded to have forward compatible implementations.
type UnimplementedRankingServer struct {
}

func (UnimplementedRankingServer) Pagenate(context.Context, *PagenateRequest) (*PagenateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Pagenate not implemented")
}
func (UnimplementedRankingServer) mustEmbedUnimplementedRankingServer() {}

// UnsafeRankingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RankingServer will
// result in compilation errors.
type UnsafeRankingServer interface {
	mustEmbedUnimplementedRankingServer()
}

func RegisterRankingServer(s grpc.ServiceRegistrar, srv RankingServer) {
	s.RegisterService(&Ranking_ServiceDesc, srv)
}

func _Ranking_Pagenate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PagenateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RankingServer).Pagenate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/taoniu.local.cryptos.grpc.binance.spot.indicators.daily.ranking.Ranking/Pagenate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RankingServer).Pagenate(ctx, req.(*PagenateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Ranking_ServiceDesc is the grpc.ServiceDesc for Ranking service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Ranking_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "taoniu.local.cryptos.grpc.binance.spot.indicators.daily.ranking.Ranking",
	HandlerType: (*RankingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Pagenate",
			Handler:    _Ranking_Pagenate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "binance/spot/indicators/daily/ranking/ranking.proto",
}