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

func (h *SpotHandler) run() (err error) {
  log.Println("binance spot api running...")

  apiContext := &common.ApiContext{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }

  r := chi.NewRouter()
  r.Route("/v1", func(r chi.Router) {
    r.Route("/binance", func(r chi.Router) {
      r.Use(api.Authenticator)
      r.Mount("/spot", v1binance.NewSpotRouter(apiContext))
    })
  })

  port := os.Getenv("CRYPTOS_API_BINANCE_SPOT_PORT")
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
