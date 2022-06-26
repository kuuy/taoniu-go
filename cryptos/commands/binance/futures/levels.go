package futures

import (
  "os"
  "fmt"
	"context"
	"time"
  "log"

  "github.com/urfave/cli/v2"

  future "taoniu.local/cryptos/models"
	pool "taoniu.local/cryptos/common"
)

func main() {
  app := &cli.App{
    Name: "binance futures levels",
    Usage: "",
    Action: func(c *cli.Context) error {
      log.Fatalln("error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "pivot",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := pivot(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
    },
    Version: "0.0.0",
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatalln("app start fatal", err)
  }
}

func pivot() error {
  
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:strategies:stoch",
  )
  if mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()

  num := 311
  for _, symbol := range symbols {
    var klines []future.Kline5s
    db.Select([]string{"price","timestamp"}).Where("symbol", symbol).Order("timestamp desc").Limit(num).Find(&klines)
    
    timestamp := klines[0].Timestamp
    price := klines[0].Price
    high := price
    low := price
    for _, item := range(klines) {
      if item.Price > high {
        high = item.Price
      }
      if item.Price < low {
        low = item.Price
      }
    }

    p := (price + high + low) / 3
    s1 := 2 * p - high
    r1 := 2 * p - low
    s2 := p - (r1 - s1)
    r2 := p + (r1 - s1)
    s3 := low - 2 * (high - p)
    r3 := high + 2 * (p - low)

    if timestamp < time.Now().Unix() - 10 {
      continue
    }

    rdb.HMSet(
      ctx,
      fmt.Sprintf("binance:futures:realtime:%s", symbol),
      map[string]interface{}{
        "r3":r3,
        "r2":r2,
        "r1":r1,
        "s1":s1,
        "s2":s2,
        "s3":s3,
      },
    )

  }

  return nil
}

