package futures

import (
  "fmt"
	"context"
  "errors"
	"time"

  "github.com/markcheno/go-talib"
	"github.com/rs/xid"

	"gorm.io/gorm"

	future "taoniu.local/cryptos/models"
	pool "taoniu.local/cryptos/common"
)

func StochRsi() error {
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:strategies:stochrsi",
  )
  if mutex.Lock(10 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()

  num := 311
  for _, symbol := range symbols {

    var klines []future.Kline5s
    db.Select([]string{"price","timestamp"}).Where("symbol", symbol).Order("timestamp desc").Limit(num).Find(&klines)
 
    var prices []float64
    timestamp := klines[0].Timestamp
    for _, item := range(klines) {
      prices = append([]float64{item.Price}, prices...)
    }
    if len(prices) < num {
      continue
    }
    fastk, fastd := talib.StochRsi(
      prices, 
      71, 
      13, 
      7,
      talib.SMA,
    )
    signal := 0
    price := prices[num-1]
    k := fastk[num-1]
    d := fastd[num-1]
    j := 3 * d - 2 * k
    if k < 20 && d < 30 && j > 30 && j < 50 {
      signal = 1
    }
    if k > 80 && d > 70 && j > 90 && j < 100 {
      signal = 2
    }
    if k == 0 || k == 100 {
      continue
    }

    if timestamp < time.Now().Unix() - 10 {
      continue
    }

    if signal == 0 {
      continue
    }
  
    indicator := "stochrsi"

    var entity future.Strategy
    result := db.Where(
      "symbol=? AND indicator=? AND timestamp>?",
      symbol,
      indicator,
      timestamp-300,
    ).First(&entity)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      continue
    }

    entity = future.Strategy{
      ID:xid.New().String(),
      Symbol:symbol,
      Indicator:indicator,
      Price:price,
      Signal:int64(signal),
      Timestamp:timestamp,
      Remark:fmt.Sprintf("k:%.2f d:%.2f j:%.2f", k, d, j),
    }
    db.Create(&entity) 
  }

  return nil
}

