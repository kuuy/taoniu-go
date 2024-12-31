package tradings

import (
  "log"
  "strconv"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  analysisRepositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
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
        Db: common.NewDB(2),
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
          side, _ := strconv.Atoi(c.Args().Get(0))
          if err := h.Flush(side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Flush(side int) error {
  log.Println("binance futures analysis tradings scalping flush...")
  err := h.AnalysisRepository.Flush(side)
  if err != nil {
    log.Println("error", err)
  }
  return nil
}
