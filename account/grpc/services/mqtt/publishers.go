package mqtt

import (
  "context"
  "taoniu.local/account/common"
  pb "taoniu.local/account/grpc/mqtt/publishers"
  "taoniu.local/account/repositories"
)

type Publishers struct {
  pb.UnimplementedPublishersServer
  GrpcContext     *common.GrpcContext
  TokenRepository *repositories.TokenRepository
}

func NewPublishers(grpcContext *common.GrpcContext) *Publishers {
  return &Publishers{
    GrpcContext:     grpcContext,
    TokenRepository: &repositories.TokenRepository{},
  }
}

func (srv *Publishers) Token(ctx context.Context, request *pb.TokenRequest) (*pb.TokenReply, error) {
  reply := &pb.TokenReply{}
  reply.AccessToken, _ = srv.TokenRepository.RefreshToken(request.Id)
  return reply, nil
}

func (srv *Publishers) Register() error {
  pb.RegisterPublishersServer(srv.GrpcContext.Conn, srv)
  return nil
}
