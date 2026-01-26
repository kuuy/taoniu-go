package swap

import (
  "context"
  "log"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories/raydium/swap"
)

type KlinesHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *swap.SymbolsRepository
  KlinesRepository  *swap.KlinesRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:  common.NewDB(3),
        Rdb: common.NewRedis(3),
        Ctx: context.Background(),
      }
      h.SymbolsRepository = &swap.SymbolsRepository{
        Db: h.Db,
      }
      h.KlinesRepository = &swap.KlinesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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
    },
  }
}

func (h *KlinesHandler) Flush(symbol string, interval string, limit int) (err error) {
  log.Println("raydium swap klines flush...", symbol, interval, limit)
  symbolInfo, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  endTime := h.KlinesRepository.Timestamp(interval)
  startTime := endTime - int64(limit)*h.KlinesRepository.Timestep(interval)
  err = h.KlinesRepository.Flush(symbol, symbolInfo.BaseAddress, symbolInfo.QuoteAddress, interval, startTime, endTime)
  if err != nil {
    log.Println("kline flush error", err)
  }
  return nil
}
