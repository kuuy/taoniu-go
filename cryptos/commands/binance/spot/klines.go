package spot

import (
  "context"
  "log"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Nats              *nats.Conn
  Repository        *repositories.KlinesRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:   common.NewDB(),
        Rdb:  common.NewRedis(),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.Repository = &repositories.KlinesRepository{
        Db:   h.Db,
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Nats: h.Nats,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(2)
          interval := c.Args().Get(0)
          limit, _ := strconv.Atoi(c.Args().Get(1))
          if interval == "1m" && (limit < 1 || limit > 4320) {
            log.Fatal("limit not in 1~4320")
            return nil
          }
          if interval == "4h" && (limit < 1 || limit > 672) {
            log.Fatal("limit not in 1~672")
            return nil
          }
          if interval == "1d" && (limit < 1 || limit > 365) {
            log.Fatal("limit not in 1~365")
            return nil
          }
          if err := h.Flush(symbol, interval, limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(2)
          interval := c.Args().Get(0)
          limit, _ := strconv.Atoi(c.Args().Get(1))
          if interval == "1m" && (limit < 1 || limit > 4320) {
            log.Fatal("limit not in 1~4320")
            return nil
          }
          if interval == "4h" && (limit < 1 || limit > 672) {
            log.Fatal("limit not in 1~672")
            return nil
          }
          if interval == "1d" && (limit < 1 || limit > 365) {
            log.Fatal("limit not in 1~365")
            return nil
          }
          if err := h.Fix(symbol, interval, limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "clean",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Clean(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *KlinesHandler) Flush(symbol string, interval string, limit int) error {
  log.Println("binance spot klines flush...")
  var symbols []string
  if symbol == "" {
    symbols = h.SymbolsRepository.Symbols()
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.Flush(symbol, interval, 0, limit)
    if err != nil {
      log.Println("kline flush error", err)
    }
  }
  return nil
}

func (h *KlinesHandler) Fix(symbol string, interval string, limit int) error {
  log.Println("binance spot klines fix...")
  var symbols []string
  if symbol == "" {
    symbols = h.SymbolsRepository.Symbols()
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.Fix(symbol, interval, limit)
    if err != nil {
      log.Println("kline fix error", err)
    }
  }
  return nil
}

func (h *KlinesHandler) Clean() error {
  log.Println("binance spot klines daily clean...")
  h.Repository.Clean()
  return nil
}
