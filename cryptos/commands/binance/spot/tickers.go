package spot

import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Task              *tasks.Tickers
  TickersRepository *repositories.TickersRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Task = &tasks.Tickers{}
      h.TickersRepository = &repositories.TickersRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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
          symbol := c.Args().Get(0)
          if err := h.Flush(symbol); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.fix(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TickersHandler) Flush(symbol string) error {
  log.Println("binance spot tickers flush...")
  //symbols := h.SymbolsRepository.Symbols()
  //log.Println(symbols)
  symbols := []string{symbol}
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    log.Println("symbols", symbols[i:j])
    err := h.TickersRepository.Flush(symbols[i:j])
    if err != nil {
      log.Println("error", err.Error())
    }
    //	h.TickersRepository.Place(symbols[i:j])
  }

  return nil
}

func (h *TickersHandler) fix() error {
  timestamp := time.Now().Unix() - 900
  symbols, _ := h.Rdb.ZRangeByScore(
    h.Ctx,
    "binance:spot:tickers:flush",
    &redis.ZRangeBy{
      Min: "-inf",
      Max: fmt.Sprintf("(%v", timestamp),
    },
  ).Result()
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    log.Println("symbols", symbols[i:j])
    h.TickersRepository.Flush(symbols[i:j])
    //	h.TickersRepository.Place(symbols[i:j])
  }
  return nil
}
