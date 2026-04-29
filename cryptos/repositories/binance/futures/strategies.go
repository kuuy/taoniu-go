package futures

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  strategiesRepositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type StrategiesRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Atr               *strategiesRepositories.AtrRepository
  Kdj               *strategiesRepositories.KdjRepository
  Rsi               *strategiesRepositories.RsiRepository
  StochRsi          *strategiesRepositories.StochRsiRepository
  Zlema             *strategiesRepositories.ZlemaRepository
  HaZlema           *strategiesRepositories.HaZlemaRepository
  BBands            *strategiesRepositories.BBandsRepository
  AndeanOscillator  *strategiesRepositories.AndeanOscillatorRepository
  IchimokuCloud     *strategiesRepositories.IchimokuCloudRepository
  SuperTrend        *strategiesRepositories.SuperTrendRepository
  Smc               *strategiesRepositories.SmcRepository
  SymbolsRepository *SymbolsRepository
}

func (r *StrategiesRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Plan{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["signal"]; ok {
    query.Where("signal", conditions["signal"].(string))
  }
  query.Count(&total)
  return total
}

func (r *StrategiesRepository) Signals(conditions map[string]interface{}) []*models.Strategy {
  var strategies []*models.Strategy
  r.Db.Select([]string{
    "price",
    "signal",
    "timestamp",
  }).Where("symbol=? AND interval=?", conditions["symbol"].(string), conditions["interval"].(string)).Find(&strategies)
  return strategies
}

func (r *StrategiesRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Strategy {
  var strategies []*models.Strategy
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "indicator",
    "signal",
    "price",
    "timestamp",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["signal"]; ok {
    query.Where("signal", conditions["signal"].(string))
  }
  query.Order("timestamp desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&strategies)
  return strategies
}

func (r *StrategiesRepository) Flush(symbol string, interval string) (err error) {
  r.Atr.Flush(symbol, interval)
  r.Kdj.Flush(symbol, interval)
  r.Rsi.Flush(symbol, interval)
  r.StochRsi.Flush(symbol, interval)
  r.Zlema.Flush(symbol, interval)
  r.HaZlema.Flush(symbol, interval)
  r.BBands.Flush(symbol, interval)
  r.IchimokuCloud.Flush(symbol, interval)
  r.SuperTrend.Flush(symbol, interval)
  r.Smc.Flush(symbol, interval)
  return
}

func (r *StrategiesRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}

func (r *StrategiesRepository) Clean(symbol string) (err error) {
  var strategy *models.Strategy
  for _, interval := range []string{"1m", "15m", "4h", "1d"} {
    for _, indicator := range []string{"kdj", "rsi", "stock_rsi", "zlema", "ha_zlema", "bbands", "ichimoku_cloud", "supertrend", "smc"} {
      var timestamp int64
      err = r.Db.Model(&strategy).Select("timestamp").Where("symbol=? AND interval = ? AND indicator = ?", symbol, interval, indicator).Order("timestamp DESC").Offset(3).Take(&timestamp).Error
      if err == nil {
        r.Db.Where("symbol=? AND interval = ? AND indicator = ? AND timestamp < ?", symbol, interval, indicator, timestamp).Delete(&strategy)
      }
    }
  }
  return
}
