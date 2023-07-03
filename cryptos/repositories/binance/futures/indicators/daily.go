package indicators

import (
  "context"
  "errors"
  "fmt"
  "sort"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/markcheno/go-talib"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type DailyRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *repositories.SymbolsRepository
}

func (r *DailyRepository) Symbols() *repositories.SymbolsRepository {
  if r.SymbolsRepository == nil {
    r.SymbolsRepository = &repositories.SymbolsRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.SymbolsRepository
}

func (r *DailyRepository) Gets(symbols []string, fields []string) []string {
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
		local key = 'binance:futures:indicators:1d:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)
  day := time.Now().Format("0102")
  keys := make([]string, len(symbols))
  for i := 0; i < len(symbols); i++ {
    keys[i] = fmt.Sprintf("%s:%s", symbols[i], day)
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

func (r *DailyRepository) Ranking(
  symbol string,
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
		local key = 'binance:futures:indicators:1d:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)

  var symbols []string
  if symbol == "" {
    symbols = r.Symbols().Symbols()
  } else {
    symbols = append(symbols, symbol)
  }

  sortIdx := -1
  day := time.Now().Format("0102")

  var keys []string
  for _, symbol := range symbols {
    keys = append(keys, fmt.Sprintf("%s:%s", symbol, day))
  }

  var args []interface{}
  for i, field := range fields {
    if field == sortField {
      sortIdx = i
    }
    args = append(args, field)
  }

  if sortIdx == -1 {
    return nil
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

  ranking := &RankingResult{
    Total: len(scores),
  }
  for _, score := range scores[offset:endPos] {
    ranking.Data = append(ranking.Data, strings.Join(
      append([]string{score.Symbol}, score.Data...),
      ",",
    ))
  }

  return ranking
}

func (r *DailyRepository) Pivot(symbol string) error {
  var kline models.Kline
  result := r.Db.Select(
    []string{"close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1d",
  ).Order(
    "timestamp desc",
  ).Take(&kline)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }

  p := (kline.Close + kline.High + kline.Low) / 3
  s1, _, _ := r.Symbols().Adjust(symbol, 2*p-kline.High, 0)
  r1, _, _ := r.Symbols().Adjust(symbol, 2*p-kline.Low, 0)
  s2, _, _ := r.Symbols().Adjust(symbol, p-(r1-s1), 0)
  r2, _, _ := r.Symbols().Adjust(symbol, p+(r1-s1), 0)
  s3, _, _ := r.Symbols().Adjust(symbol, kline.Low-2*(kline.High-p), 0)
  r3, _, _ := r.Symbols().Adjust(symbol, kline.High+2*(p-kline.Low), 0)

  day, err := r.Day(kline.Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:futures:indicators:1d:%s:%s",
    symbol,
    day,
  )
  exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
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
  if exists != 1 {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *DailyRepository) Atr(symbol string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1d",
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
    if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
      return nil
    }
    prices = append([]float64{item.Close}, prices...)
    highs = append([]float64{item.High}, highs...)
    lows = append([]float64{item.Low}, lows...)
    timestamp = item.Timestamp
  }
  if len(prices) < limit {
    return nil
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
    "binance:futures:indicators:1d:%s:%s",
    symbol,
    day,
  )
  exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "atr",
    strconv.FormatFloat(result[limit-1], 'f', -1, 64),
  )
  if exists != 1 {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *DailyRepository) Zlema(symbol string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1d",
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )
  var data []float64
  var temp []float64
  var timestamp int64
  var lag = int((period - 1) / 2)
  for _, item := range klines {
    if len(temp) < lag {
      temp = append([]float64{item.Close}, temp...)
      continue
    }
    if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
      return nil
    }
    data = append([]float64{item.Close - temp[lag-1]}, data...)
    temp = append([]float64{item.Close}, temp[:lag-1]...)
    timestamp = item.Timestamp
  }
  if len(data) < limit-lag {
    return nil
  }

  result := talib.Ema(data, period)

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:futures:indicators:1d:%s:%s",
    symbol,
    day,
  )
  exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "zlema",
    fmt.Sprintf(
      "%s,%s,%s,%d",
      strconv.FormatFloat(result[limit-lag-2], 'f', -1, 64),
      strconv.FormatFloat(result[limit-lag-1], 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      time.Now().Unix(),
    ),
  )
  if exists != 1 {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *DailyRepository) HaZlema(symbol string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1d",
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )
  var data []float64
  var temp []float64
  var timestamp int64
  var lag = int((period - 1) / 2)
  for _, item := range klines {
    var avgPrice = (item.Open + item.Close + item.High + item.Low) / 4
    if len(temp) < lag {
      temp = append([]float64{avgPrice}, temp...)
      continue
    }
    if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
      return nil
    }
    data = append([]float64{avgPrice - temp[lag-1]}, data...)
    temp = append([]float64{avgPrice}, temp[:lag-1]...)
    timestamp = item.Timestamp
  }
  if len(data) < limit-lag {
    return nil
  }

  result := talib.Ema(data, period)

  day, err := r.Day(klines[0].Timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    "binance:futures:indicators:1d:%s:%s",
    symbol,
    day,
  )
  exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "ha_zlema",
    fmt.Sprintf(
      "%s,%s,%s,%d",
      strconv.FormatFloat(result[limit-lag-2], 'f', -1, 64),
      strconv.FormatFloat(result[limit-lag-1], 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      time.Now().Unix(),
    ),
  )
  if exists != 1 {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}

func (r *DailyRepository) Kdj(symbol string, longPeriod int, shortPeriod int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1d",
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
    if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
      return nil
    }
    var avgPrice = (item.Close + item.High + item.Low) / 3
    highs = append([]float64{item.High}, highs...)
    lows = append([]float64{item.Low}, lows...)
    prices = append([]float64{avgPrice}, prices...)
    timestamp = item.Timestamp
  }
  if len(prices) < limit {
    return nil
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

  r.Rdb.HSet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1d:%s:%s",
      symbol,
      day,
    ),
    "kdj",
    fmt.Sprintf(
      "%s,%s,%s,%s,%d",
      strconv.FormatFloat(slowk[limit-1], 'f', -1, 64),
      strconv.FormatFloat(slowd[limit-1], 'f', -1, 64),
      strconv.FormatFloat(slowj[limit-1], 'f', -1, 64),
      strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
      time.Now().Unix(),
    ),
  )

  return nil
}

func (r *DailyRepository) BBands(symbol string, period int, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1d",
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
    if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
      return nil
    }
    var avgPrice = (item.Close + item.High + item.Low) / 3
    prices = append([]float64{avgPrice}, prices...)
    timestamp = item.Timestamp
  }
  if len(prices) < limit {
    return nil
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

  r.Rdb.HSet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1d:%s:%s",
      symbol,
      day,
    ),
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
      time.Now().Unix(),
    ),
  )

  return nil
}

func (r *DailyRepository) VolumeProfile(symbol string, limit int) error {
  var entity models.Symbol
  result := r.Db.Select("filters").Where("symbol", symbol).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil
  }

  var filters []string
  filters = strings.Split(entity.Filters["price"].(string), ",")
  tickSize, _ := decimal.NewFromString(filters[2])

  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "high", "low", "volume", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, "1m",
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  if len(klines) == 0 {
    return nil
  }

  now := time.Now()
  _, zoneOffset := now.Zone()

  var prices []float64
  var volumes []float64
  var offsets []int
  var minPrice float64
  var maxPrice float64
  var totalVolume float64
  var targetVolume float64
  for _, item := range klines {
    if item.Volume == 0 {
      continue
    }
    avgPrice, _ := decimal.Avg(
      decimal.NewFromFloat(item.Close),
      decimal.NewFromFloat(item.High),
      decimal.NewFromFloat(item.Low),
    ).Div(tickSize).Ceil().Mul(tickSize).Float64()

    datetime := time.Unix(item.Timestamp/1000, 0).Add(time.Second * time.Duration(zoneOffset))
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
  }

  if minPrice == maxPrice {
    return errors.New("klines not valid")
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

  values := map[string]interface{}{
    "poc":       0.0,
    "vah":       0.0,
    "val":       0.0,
    "poc_ratio": 0.0,
  }
  values["poc"], _ = decimal.Avg(
    decimal.NewFromFloat(poc["prices"].([]float64)[0]),
    decimal.NewFromFloat(poc["prices"].([]float64)[1]),
  ).Div(tickSize).Ceil().Mul(tickSize).Float64()
  values["vah"], _ = decimal.Avg(
    decimal.NewFromFloat(data[startIndex]["prices"].([]float64)[0]),
    decimal.NewFromFloat(data[startIndex]["prices"].([]float64)[1]),
  ).Div(tickSize).Ceil().Mul(tickSize).Float64()
  values["val"], _ = decimal.Avg(
    decimal.NewFromFloat(data[endIndex]["prices"].([]float64)[0]),
    decimal.NewFromFloat(data[endIndex]["prices"].([]float64)[1]),
  ).Div(tickSize).Ceil().Mul(tickSize).Float64()

  values["poc_ratio"], _ = decimal.NewFromFloat(values["vah"].(float64) - values["val"].(float64)).Div(decimal.NewFromFloat(values["poc"].(float64))).Round(4).Float64()

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1d:%s:%s",
      symbol,
      day,
    ),
    values,
  )

  return nil
}

func (r *DailyRepository) Day(timestamp int64) (day string, err error) {
  now := time.Now()
  _, offset := now.Zone()

  utc := now.Add(time.Second * time.Duration(-offset))
  duration := time.Hour*time.Duration(8-utc.Hour()) - time.Minute*time.Duration(utc.Minute()) - time.Second*time.Duration(utc.Second())
  open := utc.Add(duration)
  last := time.Unix(timestamp, 0)
  if open.Format("0102") != last.Format("0102") {
    err = errors.New("timestamp is not today")
    return
  }

  day = now.Format("0102")
  return
}
