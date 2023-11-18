package tradings

import (
  "context"
  "encoding/json"
  "errors"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  dydxModels "taoniu.local/cryptos/models/dydx"
  models "taoniu.local/cryptos/models/dydx/analysis/tradings"
  tradingsModels "taoniu.local/cryptos/models/dydx/tradings"
)

type TriggersRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *TriggersRepository) Flush(side int) error {
  now := time.Now().Add(time.Minute * -6)
  duration := time.Hour*time.Duration(-now.Hour()) + time.Minute*time.Duration(-now.Minute()) + time.Second*time.Duration(-now.Second())
  datetime := now.Add(duration)

  var analysis *models.Trigger
  result := r.Db.Where("side=? AND day=?", side, datatypes.Date(datetime)).Take(&analysis)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    analysis = &models.Trigger{
      ID:   xid.New().String(),
      Side: side,
      Day:  datatypes.Date(datetime),
    }
  }

  analysis.BuysCount = 0
  analysis.BuysAmount = 0
  analysis.SellsCount = 0
  analysis.SellsAmount = 0
  analysis.Profit = 0
  analysis.AdditiveProfit = 0

  var tradings []*tradingsModels.Trigger
  query := r.Db.Where("created_at>=? AND updated_at < ? AND status IN (1,2,3)", datetime, datetime.Add(24*time.Hour))
  if side == 1 {
    query.Where("buy_price < sell_price")
  } else {
    query.Where("buy_price > sell_price")
  }
  query.Find(&tradings)
  for _, trading := range tradings {
    if trading.Status == 1 || trading.Status == 2 || trading.Status == 3 {
      analysis.BuysCount += 1
      analysis.BuysAmount += trading.BuyPrice * trading.BuyQuantity
    }
    if trading.Status == 3 {
      analysis.SellsCount += 1
      analysis.SellsAmount += trading.SellPrice * trading.SellQuantity
      var profit float64
      if side == 1 {
        profit = r.Amount(trading.Symbol, trading.SellOrderId) - r.Amount(trading.Symbol, trading.BuyOrderId)
      } else {
        profit = r.Amount(trading.Symbol, trading.BuyOrderId) - r.Amount(trading.Symbol, trading.SellOrderId)
      }
      analysis.Profit += profit
    }
  }

  query = r.Db.Where("status=3 AND created_at<? AND updated_at>=? AND updated_at < ?", datetime, datetime, datetime.Add(24*time.Hour))
  if side == 1 {
    query.Where("buy_price < sell_price")
  } else {
    query.Where("buy_price > sell_price")
  }
  query.Find(&tradings)
  for _, trading := range tradings {
    analysis.SellsCount += 1
    analysis.SellsAmount += trading.SellPrice * trading.SellQuantity
    var profit float64
    if side == 1 {
      profit = r.Amount(trading.Symbol, trading.SellOrderId) - r.Amount(trading.Symbol, trading.BuyOrderId)
    } else {
      profit = r.Amount(trading.Symbol, trading.BuyOrderId) - r.Amount(trading.Symbol, trading.SellOrderId)
    }
    analysis.Profit += profit
    analysis.AdditiveProfit += profit
  }

  r.Db.Save(&analysis)

  return nil
}

func (r *TriggersRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Trigger{})
  query.Count(&total)
  return total
}

func (r *TriggersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Trigger {
  var analysis []*models.Trigger
  query := r.Db.Select([]string{
    "id",
    "day",
    "buys_count",
    "sells_count",
    "buys_amount",
    "sells_amount",
    "profit",
    "data",
  })
  query.Order("day desc")
  query.Offset((current - 1) * pageSize)
  query.Limit(pageSize).Find(&analysis)
  return analysis
}

func (r *TriggersRepository) Series(limit int) []interface{} {
  var analysis []*models.Trigger
  r.Db.Order("day desc").Limit(limit).Find(&analysis)

  series := make([]interface{}, len(analysis))
  for i, entity := range analysis {
    series[i] = []interface{}{
      entity.BuysCount,
      entity.SellsCount,
      time.Time(entity.Day).Format("01/02"),
    }
  }
  return series
}

func (r *TriggersRepository) Amount(symbol string, orderID string) float64 {
  var order dydxModels.Order
  r.Db.Where("symbol=? AND order_id=?", symbol, orderID).Find(&order)
  return order.AvgPrice * order.ExecutedQuantity
}

func (r *TriggersRepository) JSONMap(in interface{}) datatypes.JSONMap {
  buf, _ := json.Marshal(in)

  var out datatypes.JSONMap
  json.Unmarshal(buf, &out)
  return out
}
