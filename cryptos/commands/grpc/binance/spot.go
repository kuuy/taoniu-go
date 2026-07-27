package binance

import (
  "context"
  "fmt"
  "log"
  "net"
  "os"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "google.golang.org/grpc"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  services "taoniu.local/cryptos/grpc/services/binance"
)

type SpotHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewSpotCommand() *cli.Command {
  var h SpotHandler
  return &cli.Command{
    Name:  "spot",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SpotHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
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

func (h *SpotHandler) run() error {
  log.Println("binance spot grpc running...")

  s := grpc.NewServer()

  port := os.Getenv("CRYPTOS_GRPC_BINANCE_SPOT_PORT")
  if port == "" {
    port = os.Getenv("CRYPTOS_GRPC_PORT")
  }

  lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", port))
  if err != nil {
    log.Fatalf("net.Listen err: %v", err)
  }

  services.NewSpot(h.Db, h.Rdb, h.Ctx).Register(s)

  s.Serve(lis)

  return nil
}
