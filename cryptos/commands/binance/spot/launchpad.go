package spot

import (
  "log"
  "time"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type LaunchpadHandler struct {
  Db                  *gorm.DB
  LaunchpadRepository *repositories.LaunchpadRepository
}

func NewLaunchpadCommand() *cli.Command {
  var h LaunchpadHandler
  return &cli.Command{
    Name:  "launchpad",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = LaunchpadHandler{
        Db: common.NewDB(1),
      }
      h.LaunchpadRepository = &repositories.LaunchpadRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "apply",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.apply(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *LaunchpadHandler) apply() error {
  log.Println("spot cross margin triggers apply...")
  symbol := "SEIUSDT"
  capital := 3500.0
  price := 0.05
  corePrice := 0.05 * 12.022

  now := time.Now()
  issuedAt, _ := time.ParseInLocation("2006-01-02 15:04", "2023-08-15 20:00", now.Location())
  expiredAt := issuedAt.Add(time.Hour * 24 * 14)

  err := h.LaunchpadRepository.Apply(symbol, capital, price, corePrice, issuedAt, expiredAt)
  if err != nil {
    return err
  }

  return nil
}
