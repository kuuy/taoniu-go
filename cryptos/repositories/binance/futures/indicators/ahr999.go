package indicators

import (
  "fmt"
  "log"
  "math"
  "strconv"
  "strings"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type Ahr999Repository struct {
  BaseRepository
}

var btcGenesis = time.Date(2009, 1, 3, 0, 0, 0, 0, time.UTC)

func (r *Ahr999Repository) Get(symbol, interval string) (value, price float64, timestamp int64, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(config.REDIS_KEY_INDICATORS, interval, symbol, day)
  val, err := r.Rdb.HGet(r.Ctx, redisKey, "ahr999").Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  value, _ = strconv.ParseFloat(data[0], 64)
  price, _ = strconv.ParseFloat(data[1], 64)
  timestamp, _ = strconv.ParseInt(data[2], 10, 64)
  return
}

func (r *Ahr999Repository) Flush(symbol string, interval string, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close")
  if err != nil {
    return
  }

  closes := data[0]
  lastIdx := len(timestamps) - 1

  ema200 := talib.Ema(closes, 200)
  emaVal := ema200[lastIdx]
  if emaVal == 0 {
    err = fmt.Errorf("[%s] ema200 is zero", symbol)
    return
  }

  price := closes[lastIdx]
  days := time.Since(btcGenesis).Hours() / 24
  // Formula: 10 ^ (5.84 * log10(days) - 17.01)
  powerLawPrice := math.Pow(10, (5.84*math.Log10(days))-17.01)

  ahr999 := (price / emaVal) * (price / powerLawPrice)
  log.Println("power law price", days, ahr999, powerLawPrice)

  day, err := r.Day(timestamps[lastIdx] / 1000)
  if err != nil {
    return
  }

  redisKey := fmt.Sprintf(config.REDIS_KEY_INDICATORS, interval, symbol, day)
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "ahr999",
    fmt.Sprintf(
      "%s,%s,%d",
      strconv.FormatFloat(ahr999, 'f', -1, 64),
      strconv.FormatFloat(price, 'f', -1, 64),
      timestamps[lastIdx],
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
