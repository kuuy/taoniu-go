package spot

import (
  "context"
  "errors"
  "fmt"
  "math"
  "sort"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/markcheno/go-talib"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot"
)

type IndicatorsRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *SymbolsRepository
}

type RankingScore struct {
  Symbol string
  Value  float64
  Data   []string
}

func (r *IndicatorsRepository) Gets(symbols []string, interval string, fields []string) []string {
  var script = redis.NewScript(`
	local hmget = function (key)
		local hash = {}
		local data = redis.call('HMGET', key, unpack(ARGV))
		for i = 1, #ARGV do
			hash[i] = data[i]
		end
		return hash
	end
	local data = {}
	for i = 1, #KEYS do
		local key = 'binance:spot:indicators:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)
  day := time.Now().Format("0102")
  var keys []string
  for _, symbol := range symbols {
    keys = append(keys, fmt.Sprintf("%s:%s:%s", interval, symbol, day))
  }
  args := make([]interface{}, len(fields))
  for i := 0; i < len(fields); i++ {
    args[i] = fields[i]
  }
  result, _ := script.Run(r.Ctx, r.Rdb, keys, args...).Result()

  indicators := make([]string, len(symbols))
  for i := 0; i < len(symbols); i++ {
    item := result.([]interface{})[i]
    if item == nil {
      continue
    }
    data := make([]string, len(fields))
    for j := 0; j < len(fields); j++ {
      if item.([]interface{})[j] == nil {
        continue
      }
      data[j] = fmt.Sprintf("%v", item.([]interface{})[j])
    }
    indicators[i] = strings.Join(data, ",")
  }

  return indicators
}

func (r *IndicatorsRepository) Ranking(
  symbols []string,
  interval string,
  fields []string,
  sortField string,
  sortType int,
  current int,
  pageSize int,
) *RankingResult {
  var script = redis.NewScript(`
	local hmget = function (key)
		local hash = {}
		local data = redis.call('HMGET', key, unpack(ARGV))
		for i = 1, #ARGV do
			hash[i] = data[i]
		end
		return hash
	end
	local data = {}
	for i = 1, #KEYS do
		local key = 'binance:spot:indicators:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)

  sortIdx := -1
  day := time.Now().Format("0102")

  var keys []string
  for _, symbol := range symbols {
    keys = append(keys, fmt.Sprintf("%s:%s:%s", interval, symbol, day))
  }

  var args []interface{}
  for i, field := range fields {
    if field == sortField {
      sortIdx = i
    }
    args = append(args, field)
  }

  ranking := &RankingResult{}

  if sortIdx == -1 {
    return ranking
  }

  result, _ := script.Run(r.Ctx, r.Rdb, keys, args...).Result()

  var scores []*RankingScore
  for i := 0; i < len(symbols); i++ {
    item := result.([]interface{})[i]
    if item == nil {
      continue
    }
    if item.([]interface{})[sortIdx] == nil {
      continue
    }
    data := make([]string, len(fields))
    for j := 0; j < len(fields); j++ {
      if item.([]interface{})[j] == nil {
        continue
      }
      data[j] = fmt.Sprintf("%v", item.([]interface{})[j])
    }
    score, _ := strconv.ParseFloat(
      fmt.Sprintf("%v", item.([]interface{})[sortIdx]),
      16,
    )
    scores = append(scores, &RankingScore{
      symbols[i],
      score,
      data,
    })
  }

  if len(scores) == 0 {
    return ranking
  }

  sort.SliceStable(scores, func(i, j int) bool {
    if sortType == -1 {
      return scores[i].Value > scores[j].Value
    } else if sortType == 1 {
      return scores[i].Value < scores[j].Value
    }
    return true
  })

  offset := (current - 1) * pageSize
  endPos := offset + pageSize
  if endPos > len(scores) {
    endPos = len(scores)
  }

  ranking.Total = len(scores)
  for _, score := range scores[offset:endPos] {
    ranking.Data = append(ranking.Data, strings.Join(
      append([]string{score.Symbol}, score.Data...),
      ",",
    ))
  }

  return ranking
}

func (r *IndicatorsRepository) Pivot(symbol string, interval string) error {
  tickSize, _, err := r.Filters(symbol)
  if err != nil {
    return err
  }

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
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
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

  s1, _ = decimal.NewFromFloat(s1).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  s2, _ = decimal.NewFromFloat(s2).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  s3, _ = decimal.NewFromFloat(s3).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  r1, _ = decimal.NewFromFloat(r1).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  r2, _ = decimal.NewFromFloat(r2).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  r3, _ = decimal.NewFromFloat(r3).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()

  day, err := r.Day(kline.Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:spot:indicators:%s:%s:%s",
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
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
    }
    prices = append([]float64{item.Close}, prices...)
    highs = append([]float64{item.High}, highs...)
    lows = append([]float64{item.Low}, lows...)
    timestamp = item.Timestamp
  }
  if len(klines) < limit {
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
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
    "binance:spot:indicators:%s:%s:%s",
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
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
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
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
  }

  result := talib.Ema(data, period)

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:spot:indicators:%s:%s:%s",
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
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
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
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
  }

  result := talib.Ema(data, period)

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:spot:indicators:%s:%s:%s",
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
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
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
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
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
    "binance:spot:indicators:%s:%s:%s",
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
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
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
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
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
    "binance:spot:indicators:%s:%s:%s",
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

  var prices []decimal.Decimal
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
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
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
  }

  lastConversionLine, _ := decimal.Avg(prices[1], prices[2:tenkanPeriod]...).Float64()
  lastBaseLine, _ := decimal.Avg(prices[1], prices[2:kijunPeriod]...).Float64()

  conversionLine, _ := decimal.Avg(prices[0], prices[1:tenkanPeriod]...).Float64()
  baseLine, _ := decimal.Avg(prices[0], prices[1:kijunPeriod]...).Float64()
  senkouSpanA := (conversionLine + baseLine) / 2
  senkouSpanB, _ := decimal.Avg(prices[0], prices[1:senkouPeriod]...).Float64()
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
    "binance:spot:indicators:%s:%s:%s",
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
  tickSize, _, err := r.Filters(symbol)
  if err != nil {
    return err
  }

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
  var offsets []int
  var minPrice float64
  var maxPrice float64
  var totalVolume float64
  var targetVolume float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
    }
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()

    datetime := time.Unix(0, item.Timestamp*int64(time.Millisecond)).UTC()
    offset := datetime.Hour()*2 + 1
    if datetime.Minute() > 30 {
      offset += 1
    }

    prices = append([]float64{avgPrice}, prices...)
    volumes = append([]float64{item.Volume}, volumes...)
    offsets = append([]int{offset}, offsets...)
    if minPrice == 0 || minPrice > avgPrice {
      minPrice = avgPrice
    }
    if maxPrice < avgPrice {
      maxPrice = avgPrice
    }
    totalVolume += item.Volume
    timestamp = item.Timestamp
  }

  if len(prices) < limit {
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if minPrice == maxPrice {
    return errors.New(fmt.Sprintf("[%s] %s klines not valid", symbol, interval))
  }

  targetVolume, _ = decimal.NewFromFloat(totalVolume).Mul(decimal.NewFromFloat(0.7)).Float64()

  value := decimal.NewFromFloat(maxPrice - minPrice).Div(decimal.NewFromInt(100))

  poc := map[string]interface{}{}

  data := make([]map[string]interface{}, 100)
  for i, price := range prices {
    index, _ := decimal.NewFromFloat(maxPrice - price).Div(value).Floor().Float64()
    if index > 99.0 {
      index = 99.0
    }
    item := data[int(index)]
    if len(item) == 0 {
      item = map[string]interface{}{
        "prices":  []float64{0.0, 0.0},
        "offsets": map[int]float64{},
        "volume":  0.0,
      }
    }

    items := item["prices"].([]float64)
    if items[0] == 0.0 || items[0] > price {
      items[0] = price
    }
    if items[1] < price {
      items[1] = price
    }
    item["prices"] = items

    values := item["offsets"].(map[int]float64)
    if _, ok := values[offsets[i]]; ok {
      values[offsets[i]] += volumes[i]
    } else {
      values[offsets[i]] = volumes[i]
    }
    item["offsets"] = values
    item["volume"] = item["volume"].(float64) + volumes[i]

    if len(poc) == 0 || poc["volume"].(float64) < item["volume"].(float64) {
      poc = item
    }

    data[int(index)] = item
  }

  startIndex := 0
  endIndex := 0

  bestVolume := 0.0
  for i := 0; i < len(data); i++ {
    if len(data[i]) == 0 {
      continue
    }
    areaVolume := 0.0
    for j := i; j < len(data); j++ {
      if len(data[j]) == 0 {
        continue
      }
      areaVolume += data[j]["volume"].(float64)
      if areaVolume > targetVolume {
        if bestVolume == 0.0 || bestVolume > areaVolume {
          startIndex = i
          endIndex = j
          bestVolume = areaVolume
        }
        break
      }
    }
  }

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  if len(data[startIndex]) == 0 || len(data[endIndex]) == 0 {
    return errors.New("invalid data")
  }

  values := map[string]float64{
    "poc":       0.0,
    "vah":       0.0,
    "val":       0.0,
    "poc_ratio": 0.0,
  }
  values["poc"], _ = decimal.Avg(
    decimal.NewFromFloat(poc["prices"].([]float64)[0]),
    decimal.NewFromFloat(poc["prices"].([]float64)[1]),
  ).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  values["vah"], _ = decimal.Avg(
    decimal.NewFromFloat(data[startIndex]["prices"].([]float64)[0]),
    decimal.NewFromFloat(data[startIndex]["prices"].([]float64)[1]),
  ).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  values["val"], _ = decimal.Avg(
    decimal.NewFromFloat(data[endIndex]["prices"].([]float64)[0]),
    decimal.NewFromFloat(data[endIndex]["prices"].([]float64)[1]),
  ).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()

  values["poc_ratio"], _ = decimal.NewFromFloat(values["vah"] - values["val"]).Div(decimal.NewFromFloat(values["poc"])).Round(4).Float64()

  redisKey := fmt.Sprintf(
    "binance:spot:indicators:%s:%s:%s",
    interval,
    symbol,
    day,
  )
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "poc":       values["poc"],
      "vah":       values["vah"],
      "val":       values["val"],
      "poc_ratio": values["poc_ratio"],
    },
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *IndicatorsRepository) AndeanOscillator(symbol string, interval string, period int, length int, limit int) (err error) {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  var opens []float64
  var closes []float64
  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
    }
    opens = append([]float64{item.Open}, opens...)
    closes = append([]float64{item.Close}, closes...)
    timestamp = item.Timestamp
  }

  if len(opens) < limit {
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  up1 := make([]float64, limit)
  up2 := make([]float64, limit)
  dn1 := make([]float64, limit)
  dn2 := make([]float64, limit)
  bulls := make([]float64, limit)
  bears := make([]float64, limit)
  signals := make([]float64, limit)

  up1[0] = closes[0]
  up2[0] = math.Pow(closes[0], 2)
  dn1[0] = closes[0]
  dn2[0] = math.Pow(closes[0], 2)
  signals[0] = closes[0]

  alpha, _ := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(period + 1))).Float64()
  alphaSignal, _ := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(length + 1))).Float64()

  for i := 1; i < len(opens); i++ {
    up1[i], _ = decimal.Max(
      decimal.NewFromFloat(closes[i]),
      decimal.NewFromFloat(opens[i]),
      decimal.NewFromFloat(up1[i-1]).Sub(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(up1[i-1]).Sub(decimal.NewFromFloat(closes[i])))),
    ).Float64()
    up2[i], _ = decimal.Max(
      decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(opens[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(up2[i-1]).Sub(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(up2[i-1]).Sub(decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2))))),
    ).Float64()
    dn1[i], _ = decimal.Min(
      decimal.NewFromFloat(closes[i]),
      decimal.NewFromFloat(opens[i]),
      decimal.NewFromFloat(dn1[i-1]).Add(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(closes[i]).Sub(decimal.NewFromFloat(dn1[i-1])))),
    ).Float64()
    dn2[i], _ = decimal.Min(
      decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(opens[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(dn2[i-1]).Add(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2)).Sub(decimal.NewFromFloat(dn2[i-1])))),
    ).Float64()
    bulls[i], _ = decimal.NewFromFloat(dn2[i]).Sub(decimal.NewFromFloat(dn1[i]).Pow(decimal.NewFromInt(2))).Float64()
    bears[i], _ = decimal.NewFromFloat(up2[i]).Sub(decimal.NewFromFloat(up1[i]).Pow(decimal.NewFromInt(2))).Float64()
    if bulls[i] < 0 || bears[i] < 0 {
      err = errors.New("calc Andean Oscillator Failed")
      return
    }
    bulls[i] = math.Sqrt(bulls[i])
    bears[i] = math.Sqrt(bears[i])
    signals[i], _ = decimal.NewFromFloat(signals[i-1]).Add(decimal.NewFromFloat(alphaSignal).Mul(decimal.Max(
      decimal.NewFromFloat(bulls[i]),
      decimal.NewFromFloat(bears[i]),
    ).Sub(decimal.NewFromFloat(signals[i-1])))).Float64()
  }

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:spot:indicators:%s:%s:%s",
    interval,
    symbol,
    day,
  )
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "ao_bull":   bulls[len(opens)-1],
      "ao_bear":   bears[len(opens)-1],
      "ao_signal": signals[len(opens)-1],
    },
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}

func (r *IndicatorsRepository) Day(timestamp int64) (day string, err error) {
  now := time.Now()
  last := time.Unix(timestamp, 0)
  if now.UTC().Format("0102") != last.UTC().Format("0102") {
    err = errors.New("timestamp is not today")
    return
  }
  day = now.Format("0102")
  return
}

func (r *IndicatorsRepository) Timestep(interval string) int64 {
  if interval == "1m" {
    return 60000
  } else if interval == "15m" {
    return 900000
  } else if interval == "4h" {
    return 14400000
  }
  return 86400000
}

func (r *IndicatorsRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  if interval == "15m" {
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  } else if interval == "4h" {
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  } else if interval == "1d" {
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}

func (r *IndicatorsRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}
