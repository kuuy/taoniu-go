package spot

import (
  "context"
  "fmt"
  "log"
  tasks "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Task              *tasks.Tickers
  Repository        *repositories.TickersRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Task = &tasks.Tickers{}
      h.Repository = &repositories.TickersRepository{
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
          if err := h.Flush(); err != nil {
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

func (h *TickersHandler) Flush() error {
  log.Println("binance spot tickers flush...")
  //symbols := h.SymbolsRepository.Symbols()
  //log.Println(symbols)
  symbols := []string{"BTCUSDT"}
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    log.Println("symbols", symbols[i:j])
    err := h.Repository.Flush(symbols[i:j])
    if err != nil {
      log.Println("error", err.Error())
    }
    break
    //	h.Repository.Place(symbols[i:j])
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
    h.Repository.Flush(symbols[i:j])
    break
    //	h.Repository.Place(symbols[i:j])
  }
  return nil
}
