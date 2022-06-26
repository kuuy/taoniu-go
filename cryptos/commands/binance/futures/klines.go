package main

import (
  "os"
	"context"
	"errors"
	"time"
  "strconv"

  "github.com/urfave/cli/v2"
  "github.com/go-redis/redis/v8"
	"github.com/rs/xid"

  "github.com/RichardKnop/machinery/v2/log"

	"gorm.io/gorm"

	future "taoniu.local/cryptos/models"
	pool "taoniu.local/cryptos/common"
)

func main() {
  app := &cli.App{
    Name: "binance futures rules",
    Usage: "",
    Action: func(c *cli.Context) error {
      //fmt.Println("error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "flush5s",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := flush5s(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name: "clean5s",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := clean5s(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
    },
    Version: "0.0.0",
  }

  os.Setenv("TZ", "Asia/Shanghai")

  err := app.Run(os.Args)
  if err != nil {
    log.FATAL.Fatalln("app start fatal", err)
  }
}

func clean5s() error {
  db := pool.NewDB()
  location,_ := time.LoadLocation("Asia/Shanghai")
  expireTime := time.Now().In(location).Add(-1 * time.Hour)
  log.INFO.Println("delete kline5s", expireTime)
  db.Where("created_at < ?", expireTime).Delete(&future.Kline5s{})
  return nil
}

func flush5s() error {
  
  ctx := context.Background()
  rdb := pool.NewRedis()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:klines:flush5s",
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
  
  data, err := script.Run(ctx, rdb, symbols, "price", "volume", "quota").Result()
  if err != nil {
    log.FATAL.Fatalln("redis hmget error:", err)
    return nil
  }
  log.INFO.Println("data:", symbols, data)

  timestamp := time.Now().Unix()
  timestamp -= timestamp % 5
  
  db := pool.NewDB()
  log.INFO.Println("redis hmget:", data)
  values := data.([]interface{})
  for i, symbol := range symbols {
    price,_ := strconv.ParseFloat(values[0].([]interface{})[i].(string), 64)
    volume,_ := strconv.ParseFloat(values[1].([]interface{})[i].(string), 64)
    quota,_ := strconv.ParseFloat(values[2].([]interface{})[i].(string), 64)

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
        Volume:volume,
        Quota:quota,
        Timestamp:timestamp,
      }
      db.Create(&entity)
    } else {
      entity.Price = price
      entity.Volume = volume
      entity.Quota = quota
      entity.Timestamp = timestamp
      db.Model(&future.Kline5s{ID:entity.ID}).Updates(entity)
    }
  }

  return nil
}

