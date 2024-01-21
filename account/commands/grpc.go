package commands

import (
  "context"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "log"
  "net"

  "taoniu.local/account/common"
  "taoniu.local/account/grpc/services"
)

type GrpcHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewGrpcCommand() *cli.Command {
  var h GrpcHandler
  return &cli.Command{
    Name:  "grpc",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = GrpcHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.Run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *GrpcHandler) Run() error {
  log.Println("grpc running...")

  s := grpc.NewServer()

  lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", common.GetEnvString("ACCOUNT_GRPC_PORT")))
  if err != nil {
    log.Fatalf("net.Listen err: %v", err)
  }

  apiContext := &common.GrpcContext{
    Conn: s,
  }

  services.NewMqtt(apiContext).Register()

  s.Serve(lis)

  return nil
}
