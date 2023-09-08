package patterns

import (
  "gorm.io/gorm"
  "log"
  "strconv"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/patterns"
)

type CandlesticksHandler struct {
  Db                *gorm.DB
  Repository        *repositories.Candlesticks
  SymbolsRepository *futuresRepositories.SymbolsRepository
}

func NewCandlesticksCommand() *cli.Command {
  var h CandlesticksHandler
  return &cli.Command{
    Name:  "candlesticks",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CandlesticksHandler{
        Db: common.NewDB(),
      }
      h.Repository = &repositories.Candlesticks{
        Db: h.Db,
      }
      h.SymbolsRepository = &futuresRepositories.SymbolsRepository{
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
          if interval == "15m" && (limit < 1 || limit > 1344) {
            log.Fatal("limit not in 1~1344")
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
    },
  }
}

func (h *CandlesticksHandler) Flush(symbol string, interval string, limit int) error {
  log.Println("binance futures patterns candlesticks flush...")
  var symbols []string
  if symbol == "" {
    symbols = h.SymbolsRepository.Symbols()
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.Flush(symbol, interval, limit)
    if err != nil {
      log.Println("candlesticks patterns flush error", err)
    }
  }
  return nil
}
