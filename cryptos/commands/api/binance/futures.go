package binance

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "os"

  "github.com/go-chi/chi/v5"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/api"
  v1binance "taoniu.local/cryptos/api/v1/binance"
  "taoniu.local/cryptos/common"
)

type FuturesHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewFuturesCommand() *cli.Command {
  var h FuturesHandler
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FuturesHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      return nil
    },
    After: func(c *cli.Context) error {
      sqlDB, _ := h.Db.DB()
      sqlDB.Close()
      h.Rdb.Close()
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
  log.Println("binance futures api running...")

  apiContext := &common.ApiContext{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }

  r := chi.NewRouter()
  r.Route("/v1", func(r chi.Router) {
    r.Route("/binance", func(r chi.Router) {
      r.Use(api.Authenticator)
      r.Mount("/futures", v1binance.NewFuturesRouter(apiContext))
    })
  })

  port := os.Getenv("CRYPTOS_API_BINANCE_FUTURES_PORT")
  if port == "" {
    port = os.Getenv("CRYPTOS_API_PORT")
  }

  err = http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", port),
    r,
  )
  if err != nil {
    return
  }

  return
}
