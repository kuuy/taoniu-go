package commands

import (
  "fmt"
  "log"
  "net"
  "os"

  "github.com/urfave/cli/v2"
  "google.golang.org/grpc"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/grpc/services"
)

type GrpcHandler struct {
  Db *gorm.DB
}

func NewGrpcCommand() *cli.Command {
  var h GrpcHandler
  return &cli.Command{
    Name:  "grpc",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = GrpcHandler{
        Db: common.NewDB(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *GrpcHandler) run() error {
  log.Println("grpc running...")

  s := grpc.NewServer()

  lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", os.Getenv("CRYPTOS_GRPC_PORT")))
  if err != nil {
    log.Fatalf("net.Listen err: %v", err)
  }

  services.NewBinance(h.Db).Register(s)

  s.Serve(lis)

  return nil
}
