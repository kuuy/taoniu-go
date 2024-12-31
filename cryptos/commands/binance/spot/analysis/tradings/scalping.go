package tradings

import (
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  analysisRepositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings"
)

type ScalpingHandler struct {
  Db                 *gorm.DB
  AnalysisRepository *analysisRepositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db: common.NewDB(1),
      }
      h.AnalysisRepository = &analysisRepositories.ScalpingRepository{
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
    },
  }
}

func (h *ScalpingHandler) Flush() error {
  log.Println("binance spot analysis tradings scalping flush...")
  err := h.AnalysisRepository.Flush()
  if err != nil {
    log.Println("error", err)
  }
  return nil
}
