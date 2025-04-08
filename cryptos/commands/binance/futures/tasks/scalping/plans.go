package scalping

import (
  "context"
  "fmt"
  "log"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type PlansHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  PlansRepository    *repositories.PlansRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewPlansCommand() *cli.Command {
  var h PlansHandler
  return &cli.Command{
    Name:  "plans",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PlansHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.PlansRepository = &repositories.PlansRepository{
        Db: h.Db,
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          limit, _ := strconv.Atoi(c.Args().Get(1))
          if limit < 20 {
            limit = 20
          }
          if err := h.Flush(interval, limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *PlansHandler) Flush(interval string, limit int) error {
  log.Println("binance futures tasks scalping plans flush...")

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:futures:scalping:plans:%s:flush", interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  conditions := map[string]interface{}{
    "interval":   interval,
    "expired_at": time.Now().Add(-180 * time.Second),
  }
  plans := h.PlansRepository.Ranking(
    []string{"id"},
    conditions,
    "created_at",
    -1,
    limit,
  )
  for _, plan := range plans {
    log.Println("plan", plan.ID)
    if !h.ScalpingRepository.IsPlanExists(plan.ID) {
      h.ScalpingRepository.AddPlan(plan.ID)
    }
  }

  return nil
}
