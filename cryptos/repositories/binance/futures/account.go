package futures

import (
  "context"
  "errors"
  "fmt"
  "os"
  "strconv"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type AccountRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *AccountRepository) Flush() error {
  client := binance.NewFuturesClient(
    os.Getenv("BINANCE_FUTURES_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_FUTURES_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_FUTURES_API_ENDPOINT")
  account, err := client.NewGetAccountService().Do(r.Ctx)
  if err != nil {
    return err
  }

  for _, position := range account.Positions {
    if position.Isolated || position.UpdateTime == 0 {
      continue
    }

    var side int
    if fmt.Sprintf("%v", position.PositionSide) == "LONG" {
      side = 1
    } else {
      side = 2
    }

    leverage, _ := strconv.Atoi(position.Leverage)
    entryPrice, _ := strconv.ParseFloat(position.EntryPrice, 64)
    volume, _ := strconv.ParseFloat(position.PositionAmt, 64)
    notional, _ := strconv.ParseFloat(position.Notional, 64)

    var status int
    if notional == 0.0 {
      status = 2
    } else {
      status = 1
    }

    var entity models.Position
    result := r.Db.Where(
      "symbol=? AND side=? AND status=1",
      position.Symbol,
      side,
    ).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if status == 2 {
        continue
      }
      entity = models.Position{
        ID:         xid.New().String(),
        Symbol:     position.Symbol,
        Leverage:   leverage,
        Side:       side,
        EntryPrice: entryPrice,
        Volume:     volume,
        Notional:   notional,
        Timestamp:  position.UpdateTime,
        Status:     status,
      }
      r.Db.Create(&entity)
    } else {
      entity.Leverage = leverage
      entity.EntryPrice = entryPrice
      entity.Volume = volume
      entity.Notional = notional
      entity.Status = status
      entity.Timestamp = position.UpdateTime
      r.Db.Model(&models.Position{ID: entity.ID}).Updates(entity)
    }
  }

  r.Rdb.HMSet(r.Ctx, "binance:futures:balance:USDT", map[string]interface{}{
    "balance":           account.TotalWalletBalance,
    "available_balance": account.AvailableBalance,
    "unrealized_profit": account.TotalUnrealizedProfit,
  })

  //positions := make(map[string]interface{})
  //for _, position := range account.Positions {
  //	symbol := position.Symbol
  //	side := fmt.Sprintf("%s", position.PositionSide)
  //	leverage, _ := strconv.ParseInt(position.Leverage, 10, 64)
  //	entryPrice, _ := strconv.ParseFloat(position.EntryPrice, 64)
  //	margin, _ := strconv.ParseFloat(position.PositionInitialMargin, 64)
  //	notional, _ := strconv.ParseFloat(position.Notional, 64)
  //	maxNotional, _ := strconv.ParseInt(position.MaxNotional, 10, 64)
  //	unrealizedProfit, _ := strconv.ParseFloat(position.UnrealizedProfit, 64)
  //	if notional == 0.0 {
  //		continue
  //	}
  //	var entity map[string]interface{}
  //	if value, ok := positions[symbol]; !ok {
  //		entity = map[string]interface{}{
  //			"symbol":                  symbol,
  //			"leverage":                leverage,
  //			"notional":                notional,
  //			"max_notional":            maxNotional,
  //			"long_entry_price":        0.0,
  //			"long_margin":             0.0,
  //			"long_notional":           0.0,
  //			"long_unrealized_profit":  0.0,
  //			"short_entry_price":       0.0,
  //			"short_margin":            0.0,
  //			"short_notional":          0.0,
  //			"short_unrealized_profit": 0.0,
  //		}
  //	} else {
  //		entity = value.(map[string]interface{})
  //	}
  //	if side == "LONG" {
  //		entity["long_entry_price"] = entryPrice
  //		entity["long_margin"] = margin
  //		entity["long_notional"] = notional
  //		entity["long_unrealized_profit"] = unrealizedProfit
  //	}
  //	if side == "SHORT" {
  //		entity["short_entry_price"] = entryPrice
  //		entity["short_margin"] = margin
  //		entity["short_notional"] = notional
  //		entity["short_unrealized_profit"] = unrealizedProfit
  //	}
  //	entity["notional"] = entity["long_notional"].(float64) - entity["short_notional"].(float64)
  //	positions[symbol] = entity
  //}
  //
  //for symbol, entity := range positions {
  //	rdb.HMSet(ctx, fmt.Sprintf("binance:futures:positions:%s", symbol), entity)
  //	rdb.SAdd(ctx, "binance:futures:trading", symbol)
  //	rdb.SRem(ctx, "binance:futures:untrading", symbol)
  //}
  //
  //symbols, _ := rdb.SMembers(ctx, "binance:futures:trading").Result()
  //for _, symbol := range symbols {
  //	if _, ok := positions[symbol]; !ok {
  //		rdb.Del(ctx, fmt.Sprintf("binance:futures:positions:%s", symbol))
  //		rdb.SRem(ctx, "binance:futures:trading", symbol)
  //		rdb.SAdd(ctx, "binance:futures:untrading", symbol)
  //	}
  //}

  return nil
}

func (r *AccountRepository) Balance(symbol string) (float64, float64, error) {
  return 0, 0, nil
}
