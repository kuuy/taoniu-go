package binance

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "os"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/socket"
)

type FuturesHandler struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Nats *nats.Conn
}

func NewFuturesCommand() *cli.Command {
  var h FuturesHandler
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FuturesHandler{
        Db:   common.NewDB(2),
        Rdb:  common.NewRedis(2),
        Ctx:  context.Background(),
        Nats: common.NewNats(2),
      }
      return nil
    },
    After: func(c *cli.Context) error {
      sqlDB, _ := h.Db.DB()
      sqlDB.Close()
      h.Rdb.Close()
      h.Nats.Close()
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

func (h *FuturesHandler) run() (err error) {
  log.Println("binance futures socket running...")

  socketContext := &common.SocketContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Nats: h.Nats,
  }

  server := socket.NewServer(socketContext)

  port := os.Getenv("CRYPTOS_SOCKET_BINANCE_FUTURES_PORT")
  if port == "" {
    port = os.Getenv("CRYPTOS_SOCKET_PORT")
  }

  err = http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", port),
    server,
  )
  if err != nil {
    return
  }

  return
}
