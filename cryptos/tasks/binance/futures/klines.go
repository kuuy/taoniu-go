package futures

import (
  "context"
  "errors"
  "time"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"

  "gorm.io/gorm"

  pool "taoniu.local/cryptos/common"
  future "taoniu.local/cryptos/models"
)

func CleanKline5s() error {
  db := pool.NewDB()
  location,_ := time.LoadLocation("Asia/Shanghai")
  expireTime := time.Now().In(location).Add(-1 * time.Hour)
  db.Where("created_at < ?", expireTime).Delete(&future.Kline5s{})
  return nil
}

func FlushKline5s() error {
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:klines:flush:5s",
  )
  if mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()

  script := redis.NewScript(`
  local result = {}
  for i=1, #KEYS do
    local key = "binance:futures:realtime:" .. KEYS[i]
      result[i] = redis.pcall("HMGET", key, unpack(ARGV))
    end
  return result
  `)

  data, err := script.Run(
    ctx,
    rdb,
    symbols,
    "price",
    "high",
    "low",
    "volume",
    "quota",
  ).Result()
  if err != nil {
    return nil
  }

  timestamp := time.Now().Unix()
  timestamp -= timestamp % 5

  for i, symbol := range symbols {
    values := data.([]interface{})[i].([]interface{})
    price,_ := strconv.ParseFloat(values[0].(string), 64)
    high,_ := strconv.ParseFloat(values[1].(string), 64)
    low,_ := strconv.ParseFloat(values[2].(string), 64)
    volume,_ := strconv.ParseFloat(values[3].(string), 64)
    quota,_ := strconv.ParseFloat(values[4].(string), 64)

    var entity future.Kline5s
    result := db.Where(
      "symbol=? AND timestamp=?",
      symbol,
      timestamp,
    ).First(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = future.Kline5s{
        ID:xid.New().String(),
        Symbol:symbol,
        Price:price,
        High:high,
        Low:low,
        Volume:volume,
        Quota:quota,
        Timestamp:timestamp,
      }
      db.Create(&entity)
    } else {
      entity.Price = price
      entity.High = high
      entity.Low = low
      entity.Volume = volume
      entity.Quota = quota
      entity.Timestamp = timestamp
      db.Model(&future.Kline5s{ID:entity.ID}).Updates(entity)
    }

  }

  return nil
}
