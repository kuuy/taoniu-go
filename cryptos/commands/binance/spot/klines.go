package spot

import (
  "context"
  "errors"
  "log"
  "slices"
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
  KlinesRepository  *repositories.KlinesRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:   common.NewDB(1),
        Rdb:  common.NewRedis(1),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.KlinesRepository = &repositories.KlinesRepository{
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
          intervals := []string{"1m", "15m", "4h", "1d"}
          if !slices.Contains(intervals, interval) {
            return errors.New("interval not valid")
          }
          if interval == "1m" && (limit < 1 || limit > 1440) {
            log.Fatal("limit not in 1~1440")
            return nil
          }
          if interval == "15m" && (limit < 1 || limit > 672) {
            log.Fatal("limit not in 1~672")
            return nil
          }
          if interval == "4h" && (limit < 1 || limit > 126) {
            log.Fatal("limit not in 1~126")
            return nil
          }
          if interval == "1d" && (limit < 1 || limit > 100) {
            log.Fatal("limit not in 1~100")
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
          if interval == "15m" && (limit < 1 || limit > 672) {
            log.Fatal("limit not in 1~672")
            return nil
          }
          if interval == "4h" && (limit < 1 || limit > 126) {
            log.Fatal("limit not in 1~126")
            return nil
          }
          if interval == "1d" && (limit < 1 || limit > 100) {
            log.Fatal("limit not in 1~100")
            return nil
          }
          if err := h.Flush(symbol, interval, limit); err != nil {
            return cli.Exit(err.Error(), 1)
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
  log.Println("binance spot klines flushing...")
  var symbols []string
  if symbol == "" {
    symbols = h.SymbolsRepository.Symbols()
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.KlinesRepository.Flush(symbol, interval, 0, limit)
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
    err := h.KlinesRepository.Fix(symbol, interval, limit)
    if err != nil {
      log.Println("kline fix error", err)
    }
  }
  return nil
}

func (h *KlinesHandler) Clean() error {
  log.Println("binance spot klines clean...")
  symbols := h.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    h.KlinesRepository.Clean(symbol)
  }
  return nil
}
