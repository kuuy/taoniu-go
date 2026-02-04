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

type AtrRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *AtrRepository) Get(symbol, interval string) (result float64, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  val, err := r.Rdb.HGet(
    r.Ctx,
    redisKey,
    "atr",
  ).Result()
  if err != nil {
    return
  }
  result, err = strconv.ParseFloat(val, 64)
  return
}

func (r *AtrRepository) Multiplier(price, atr float64) float64 {
  if price == 0 {
    return 2.0
  }
  volatility := atr / price
  switch {
  case volatility > 0.05: // 高波动 >5%
    return 2.5
  case volatility > 0.03: // 中波动 3-5%
    return 2.0
  case volatility > 0.015: // 中低波动 1.5-3%
    return 1.5
  case volatility > 0.008: // 低波动 0.8-1.5%
    return 1.2
  default: // 极低波动 <0.8%
    return 1.0
  }
}
