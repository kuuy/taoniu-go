package futures

import (
  "fmt"
	"context"
	"time"
  "strconv"

  "github.com/adshao/go-binance/v2"

	pool "taoniu.local/cryptos/common"
)

func FlushAccount() error {
  
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:account:flush",
  )
  if mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  apiKey := "YDqXQSPkjrhwPeHwxUNt10aaT0NrWevhZMhG826TswB78i9mFRmA7qsNhY8gxvac"
  secretKey := "qbymjie8fSBYg3wlabaekhplxa8uMnThARhcbjA8uP4649OSTZVN5N57IyyRawCx"

  client := binance.NewFuturesClient(apiKey, secretKey)

  account,err := client.NewGetAccountService().Do(ctx)
  if err != nil {
    return nil
  }

  rdb.HMSet(ctx, "binance:futures:balance:USDT", map[string]interface{}{
    "balance":account.TotalWalletBalance,
    "unrealized_profit":account.TotalUnrealizedProfit,
  })

  positions := make(map[string]interface{})
  for _,position := range account.Positions {
    symbol := position.Symbol
    side := fmt.Sprintf("%s", position.PositionSide)
    leverage,_ := strconv.ParseInt(position.Leverage, 10, 64) 
    entryPrice,_ := strconv.ParseFloat(position.EntryPrice, 64) 
    margin,_ := strconv.ParseFloat(position.PositionInitialMargin, 64) 
    notional,_ := strconv.ParseFloat(position.Notional, 64) 
    maxNotional,_ := strconv.ParseInt(position.MaxNotional, 10, 64) 
    unrealizedProfit,_ := strconv.ParseFloat(position.UnrealizedProfit, 64)
    if notional == 0.0 {
      continue
    }
    var entity map[string]interface{}
    if value,ok := positions[symbol]; !ok {
      entity = map[string]interface{}{
        "symbol":symbol,
        "leverage":leverage,
        "notional":notional,
        "max_notional":maxNotional,
        "long_entry_price":0.0,
        "long_margin":0.0,
        "long_notional":0.0,
        "long_unrealized_profit":0.0,
        "short_entry_price":0.0,
        "short_margin":0.0,
        "short_notional":0.0,
        "short_unrealized_profit":0.0,
      }
    } else {
      entity = value.(map[string]interface{})
    }
    if side == "LONG" {
      entity["long_entry_price"] = entryPrice
      entity["long_margin"] = margin
      entity["long_notional"] = notional
      entity["long_unrealized_profit"] = unrealizedProfit
    }
    if side == "SHORT" {
      entity["short_entry_price"] = entryPrice
      entity["short_margin"] = margin
      entity["short_notional"] = notional
      entity["short_unrealized_profit"] = unrealizedProfit
    }
    entity["notional"] = entity["long_notional"].(float64) - entity["short_notional"].(float64)
    positions[symbol] = entity
  }

  for symbol,entity := range positions {
    rdb.HMSet(ctx, fmt.Sprintf("binance:futures:positions:%s", symbol), entity)
    rdb.SAdd(ctx, "binance:futures:trading", symbol)
    rdb.SRem(ctx, "binance:futures:untrading", symbol)
  }

  symbols, _ := rdb.SMembers(ctx, "binance:futures:trading").Result()
  for _,symbol := range symbols {
    if _,ok := positions[symbol]; !ok {
      rdb.Del(ctx, fmt.Sprintf("binance:futures:positions:%s", symbol))
      rdb.SRem(ctx, "binance:futures:trading", symbol)
      rdb.SAdd(ctx, "binance:futures:untrading", symbol)
    }
  }

  return nil
}

