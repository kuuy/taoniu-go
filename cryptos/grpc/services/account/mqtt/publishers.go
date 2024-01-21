package mqtt

import (
  "context"
  "fmt"
  "os"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"

  pb "taoniu.local/cryptos/grpc/account/mqtt/publishers"
)

type Publishers struct {
  Ctx              context.Context
  PublishersClient pb.PublishersClient
}

func (srv *Publishers) Client() pb.PublishersClient {
  if srv.PublishersClient == nil {
    conn, err := grpc.Dial(
      fmt.Sprintf("%v:%v", os.Getenv("ACCOUNT_GRPC_HOST"), os.Getenv("ACCOUNT_GRPC_PORT")),
      grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
      panic(err.Error())
    }
    srv.PublishersClient = pb.NewPublishersClient(conn)
  }
  return srv.PublishersClient
}

func (srv *Publishers) Token(id string) (r *pb.TokenReply, err error) {
  return srv.Client().Token(srv.Ctx, &pb.TokenRequest{Id: id})
}
