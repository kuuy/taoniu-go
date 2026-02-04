package swap

import (
  "context"
  "errors"
  "fmt"
  "math"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/markcheno/go-talib"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/raydium/swap"
  models "taoniu.local/cryptos/models/raydium/swap"
)

type IndicatorsRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *SymbolsRepository
  KlinesRepository  *KlinesRepository
}

func (r *IndicatorsRepository) Day(timestamp int64) (string, error) {
  return time.Unix(timestamp, 0).Format("0102"), nil
}

func (r *IndicatorsRepository) Timestep(interval string) int64 {
  return r.KlinesRepository.Timestep(interval)
}

func (r *IndicatorsRepository) Timestamp(interval string) int64 {
  return r.KlinesRepository.Timestamp(interval)
}

func (r *IndicatorsRepository) Pivot(symbol string, interval string) error {
  var kline models.Kline
  result := r.Db.Select(
    []string{"close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Take(&kline)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }

  if kline.Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  p := decimal.Avg(
    decimal.NewFromFloat(kline.Close),
    decimal.NewFromFloat(kline.High),
    decimal.NewFromFloat(kline.Low),
  )

  s1, _ := p.Mul(decimal.NewFromInt(2)).Sub(decimal.NewFromFloat(kline.High)).Float64()
  r1, _ := p.Mul(decimal.NewFromInt(2)).Sub(decimal.NewFromFloat(kline.Low)).Float64()
  s2, _ := p.Sub(decimal.NewFromFloat(r1).Sub(decimal.NewFromFloat(s1))).Float64()
  r2, _ := p.Add(decimal.NewFromFloat(r1).Sub(decimal.NewFromFloat(s1))).Float64()
  s3, _ := decimal.NewFromFloat(kline.Low).Sub(decimal.NewFromFloat(kline.High).Sub(p).Mul(decimal.NewFromInt(2))).Float64()
  r3, _ := decimal.NewFromFloat(kline.High).Add(p.Sub(decimal.NewFromFloat(kline.Low)).Mul(decimal.NewFromInt(2))).Float64()

  day, err := r.Day(kline.Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "r3": r3,
      "r2": r2,
      "r1": r1,
      "s1": s1,
      "s2": s2,
      "s3": s3,
    },
  )
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) Atr(symbol string, interval string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  var highs []float64
  var lows []float64
  var prices []float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return fmt.Errorf("[%s] %s klines lost", symbol, interval)
    }
    prices = append([]float64{item.Close}, prices...)
    highs = append([]float64{item.High}, highs...)
    lows = append([]float64{item.Low}, lows...)
    timestamp = item.Timestamp
  }
  if len(klines) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  result := talib.Atr(
    highs,
    lows,
    prices,
    period,
  )

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "atr",
    strconv.FormatFloat(result[limit-1], 'f', -1, 64),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) Zlema(symbol string, interval string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  lag := (period - 1) / 2

  var data []float64
  var temp []float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return fmt.Errorf("[%s] %s klines lost", symbol, interval)
    }
    if len(temp) < lag {
      temp = append([]float64{item.Close}, temp...)
    } else {
      data = append([]float64{item.Close - temp[lag-1]}, data...)
      temp = append([]float64{item.Close}, temp[:lag-1]...)
    }
    timestamp = item.Timestamp
  }

  if len(klines) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  result := talib.Ema(data, period)

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "zlema",
    fmt.Sprintf(
      "%s,%s,%s,%d",
      strconv.FormatFloat(result[limit-lag-2], 'f', -1, 64),
      strconv.FormatFloat(result[limit-lag-1], 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      klines[0].Timestamp,
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) HaZlema(symbol string, interval string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  lag := (period - 1) / 2

  var data []float64
  var temp []float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return fmt.Errorf("[%s] %s klines lost", symbol, interval)
    }
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Open),
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Float64()
    if len(temp) < lag {
      temp = append([]float64{avgPrice}, temp...)
    } else {
      data = append([]float64{avgPrice - temp[lag-1]}, data...)
      temp = append([]float64{avgPrice}, temp[:lag-1]...)
    }
    timestamp = item.Timestamp
  }

  if len(klines) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  result := talib.Ema(data, period)

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "ha_zlema",
    fmt.Sprintf(
      "%s,%s,%s,%d",
      strconv.FormatFloat(result[limit-lag-2], 'f', -1, 64),
      strconv.FormatFloat(result[limit-lag-1], 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      klines[0].Timestamp,
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) Kdj(symbol string, interval string, longPeriod int, shortPeriod int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  var highs []float64
  var lows []float64
  var prices []float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return fmt.Errorf("[%s] %s klines lost", symbol, interval)
    }
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Float64()
    highs = append([]float64{item.High}, highs...)
    lows = append([]float64{item.Low}, lows...)
    prices = append([]float64{avgPrice}, prices...)
    timestamp = item.Timestamp
  }

  if len(prices) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  slowk, slowd := talib.Stoch(highs, lows, prices, longPeriod, shortPeriod, 0, shortPeriod, 0)
  var slowj []float64
  for i := 0; i < limit; i++ {
    slowj = append(slowj, 3*slowk[i]-2*slowd[i])
  }

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "kdj",
    fmt.Sprintf(
      "%s,%s,%s,%s,%d",
      strconv.FormatFloat(slowk[limit-1], 'f', -1, 64),
      strconv.FormatFloat(slowd[limit-1], 'f', -1, 64),
      strconv.FormatFloat(slowj[limit-1], 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      klines[0].Timestamp,
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) BBands(symbol string, interval string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  var prices []float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return fmt.Errorf("[%s] %s klines lost", symbol, interval)
    }
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Float64()
    prices = append([]float64{avgPrice}, prices...)
    timestamp = item.Timestamp
  }

  if len(klines) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  uBands, mBands, lBands := talib.BBands(prices, period, 2, 2, 0)
  p1 := (klines[2].Close + klines[2].High + klines[2].Low) / 3
  p2 := (klines[1].Close + klines[1].High + klines[1].Low) / 3
  p3 := (klines[0].Close + klines[0].High + klines[0].Low) / 3
  b1 := (p1 - lBands[limit-3]) / (uBands[limit-3] - lBands[limit-3])
  b2 := (p2 - lBands[limit-2]) / (uBands[limit-2] - lBands[limit-2])
  b3 := (p3 - lBands[limit-1]) / (uBands[limit-1] - lBands[limit-1])
  w1 := (uBands[limit-3] - lBands[limit-3]) / mBands[limit-3]
  w2 := (uBands[limit-2] - lBands[limit-2]) / mBands[limit-2]
  w3 := (uBands[limit-1] - lBands[limit-1]) / mBands[limit-1]

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "bbands",
    fmt.Sprintf(
      "%s,%s,%s,%s,%s,%s,%s,%d",
      strconv.FormatFloat(b1, 'f', -1, 64),
      strconv.FormatFloat(b2, 'f', -1, 64),
      strconv.FormatFloat(b3, 'f', -1, 64),
      strconv.FormatFloat(w1, 'f', -1, 64),
      strconv.FormatFloat(w2, 'f', -1, 64),
      strconv.FormatFloat(w3, 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      klines[0].Timestamp,
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }
  return nil
}

func (r *IndicatorsRepository) IchimokuCloud(symbol string, interval string, tenkanPeriod int, kijunPeriod int, senkouPeriod int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  var prices []decimal.Decimal
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return fmt.Errorf("[%s] %s klines lost", symbol, interval)
    }
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Float64()
    prices = append(prices, decimal.NewFromFloat(avgPrice))
    timestamp = item.Timestamp
  }

  if len(klines) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  lastConversionLine, _ := decimal.Avg(
    decimal.Min(prices[1], prices[2:tenkanPeriod]...),
    decimal.Max(prices[1], prices[2:tenkanPeriod]...),
  ).Float64()
  lastBaseLine, _ := decimal.Avg(
    decimal.Min(prices[1], prices[2:kijunPeriod]...),
    decimal.Max(prices[1], prices[2:kijunPeriod]...),
  ).Float64()
  conversionLine, _ := decimal.Avg(
    decimal.Min(prices[0], prices[1:tenkanPeriod]...),
    decimal.Max(prices[0], prices[1:tenkanPeriod]...),
  ).Float64()
  baseLine, _ := decimal.Avg(
    decimal.Min(prices[0], prices[1:kijunPeriod]...),
    decimal.Max(prices[0], prices[1:kijunPeriod]...),
  ).Float64()
  senkouSpanA := (conversionLine + baseLine) / 2
  senkouSpanB, _ := decimal.Avg(
    decimal.Min(prices[0], prices[1:senkouPeriod]...),
    decimal.Max(prices[0], prices[1:senkouPeriod]...),
  ).Float64()
  chikouSpan, _ := decimal.Avg(
    decimal.Min(prices[0], prices[1:kijunPeriod]...),
    decimal.Max(prices[0], prices[1:kijunPeriod]...),
  ).Float64()

  var signal int
  if conversionLine > baseLine && lastConversionLine < lastBaseLine {
    signal = 1
  }
  if conversionLine < baseLine && lastConversionLine > lastBaseLine {
    signal = 2
  }

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )

  if signal == 0 {
    val, _ := r.Rdb.HGet(
      r.Ctx,
      redisKey,
      "ichimoku_cloud",
    ).Result()
    data := strings.Split(val, ",")
    if len(data) == 8 {
      lastConversionLine, _ = strconv.ParseFloat(data[1], 64)
      lastBaseLine, _ = strconv.ParseFloat(data[2], 64)
      if conversionLine > baseLine && lastConversionLine < lastBaseLine {
        signal = 1
      }
      if conversionLine < baseLine && lastConversionLine > lastBaseLine {
        signal = 2
      }
    }
  }

  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "ichimoku_cloud",
    fmt.Sprintf(
      "%d,%s,%s,%s,%s,%s,%s,%d",
      signal,
      strconv.FormatFloat(conversionLine, 'f', -1, 64),
      strconv.FormatFloat(baseLine, 'f', -1, 64),
      strconv.FormatFloat(senkouSpanA, 'f', -1, 64),
      strconv.FormatFloat(senkouSpanB, 'f', -1, 64),
      strconv.FormatFloat(chikouSpan, 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      klines[0].Timestamp,
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) VolumeProfile(symbol string, interval string, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "high", "low", "volume", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  var prices []float64
  var volumes []float64
  for _, item := range klines {
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Float64()
    prices = append(prices, avgPrice)
    volumes = append(volumes, item.Volume)
  }

  if len(klines) < limit {
    return fmt.Errorf("[%s] %s klines not enough", symbol, interval)
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
  }

  var minPrice, maxPrice float64
  for i, price := range prices {
    if i == 0 || price < minPrice {
      minPrice = price
    }
    if i == 0 || price > maxPrice {
      maxPrice = price
    }
  }

  // Simple volume profile implementation: split range into 24 bins
  bins := 24
  binSize := (maxPrice - minPrice) / float64(bins)
  if binSize == 0 {
    return nil
  }
  profile := make([]float64, bins)
  for i, price := range prices {
    binIdx := int(math.Floor((price - minPrice) / binSize))
    if binIdx >= bins {
      binIdx = bins - 1
    }
    profile[binIdx] += volumes[i]
  }

  var pocIdx int
  for i, vol := range profile {
    if vol > profile[pocIdx] {
      pocIdx = i
    }
  }
  poc := minPrice + float64(pocIdx)*binSize + binSize/2

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "volume_profile",
    fmt.Sprintf(
      "%s,%s,%d",
      strconv.FormatFloat(poc, 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      klines[0].Timestamp,
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) Flush(symbol string, interval string, limit int) error {
  if err := r.Pivot(symbol, interval); err != nil {
    return err
  }
  if err := r.Atr(symbol, interval, 14, limit); err != nil {
    return err
  }
  if err := r.Zlema(symbol, interval, 14, limit); err != nil {
    return err
  }
  if err := r.HaZlema(symbol, interval, 14, limit); err != nil {
    return err
  }
  if err := r.Kdj(symbol, interval, 9, 3, limit); err != nil {
    return err
  }
  if err := r.BBands(symbol, interval, 20, limit); err != nil {
    return err
  }
  if err := r.IchimokuCloud(symbol, interval, 9, 26, 52, limit); err != nil {
    return err
  }
  if err := r.VolumeProfile(symbol, interval, limit); err != nil {
    return err
  }
  return nil
}
