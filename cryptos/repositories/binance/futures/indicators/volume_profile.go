package indicators

import (
  "context"
  "fmt"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/binance/futures"
)

type VolumeProfileRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *VolumeProfileRepository) Get(symbol, interval string) (poc, vah, val float64, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )

  pocVal, err := r.Rdb.HGet(r.Ctx, redisKey, "poc").Result()
  if err != nil {
    return
  }
  poc, _ = strconv.ParseFloat(pocVal, 64)

  vahVal, _ := r.Rdb.HGet(r.Ctx, redisKey, "vah").Result()
  vah, _ = strconv.ParseFloat(vahVal, 64)

  valVal, _ := r.Rdb.HGet(r.Ctx, redisKey, "val").Result()
  val, _ = strconv.ParseFloat(valVal, 64)

  return
}

func (r *VolumeProfileRepository) StructureSupport(entryPrice, poc, vah, val float64) float64 {
  if poc == 0 {
    return 0
  }
  if entryPrice > vah && vah > 0 {
    return vah
  }
  if entryPrice > poc {
    return poc
  }
  if entryPrice > val && val > 0 {
    return val
  }
  return 0
}

func (r *VolumeProfileRepository) StructureResistance(entryPrice, poc, vah, val float64) float64 {
  if poc == 0 {
    return 0
  }
  if entryPrice < val && val > 0 {
    return val
  }
  if entryPrice < poc {
    return poc
  }
  if entryPrice < vah && vah > 0 {
    return vah
  }
  return 0
}
