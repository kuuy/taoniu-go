package commands

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

  "taoniu.local/cryptos/api/v1"
  "taoniu.local/cryptos/common"
)

type ApiHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewApiCommand() *cli.Command {
  var h ApiHandler
  return &cli.Command{
    Name:  "api",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ApiHandler{
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

func (h *ApiHandler) run() error {
  log.Println("api running...")

  apiContext := &common.ApiContext{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }

  r := chi.NewRouter()
  r.Route("/v1", func(r chi.Router) {
    r.Mount("/currencies", v1.NewCurrenciesRouter(apiContext))
    r.Mount("/binance", v1.NewBinanceRouter(apiContext))
    r.Mount("/dydx", v1.NewDydxRouter(apiContext))
    r.Mount("/tradingview", v1.NewTradingviewRouter(apiContext))
  })

  err := http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", os.Getenv("CRYPTOS_API_PORT")),
    r,
  )
  if err != nil {
    return err
  }

  return nil
}
