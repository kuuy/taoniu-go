package fishers

import (
  "context"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot/analysis/tradings/fishers"
)

type ChartRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *ChartRepository) Series(limit int) []string {
  var grids []*models.Grid
  r.Db.Order("day desc").Limit(limit).Find(&grids)

  series := make([]string, len(grids))
  for i, grid := range grids {
    series[i] = strings.Join(
      []string{
        strconv.Itoa(grid.BuysCount),
        strconv.Itoa(grid.SellsCount),
        time.Time(grid.Day).Format("01/02"),
      },
      ",",
    )
  }
  return series
}
