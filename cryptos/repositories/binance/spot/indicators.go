package spot

import (
  "context"
  "errors"
  "fmt"
  "sort"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type IndicatorsRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Atr               *indicatorsRepositories.AtrRepository
  Pivot             *indicatorsRepositories.PivotRepository
  Kdj               *indicatorsRepositories.KdjRepository
  Rsi               *indicatorsRepositories.RsiRepository
  StochRsi          *indicatorsRepositories.StochRsiRepository
  Zlema             *indicatorsRepositories.ZlemaRepository
  HaZlema           *indicatorsRepositories.HaZlemaRepository
  BBands            *indicatorsRepositories.BBandsRepository
  AndeanOscillator  *indicatorsRepositories.AndeanOscillatorRepository
  IchimokuCloud     *indicatorsRepositories.IchimokuCloudRepository
  SuperTrend        *indicatorsRepositories.SuperTrendRepository
  Smc               *indicatorsRepositories.SmcRepository
  VolumeMoving      *indicatorsRepositories.VolumeMovingRepository
  VolumeProfile     *indicatorsRepositories.VolumeProfileRepository
  Ahr999            *indicatorsRepositories.Ahr999Repository
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
    switch sortType {
    case -1:
      return scores[i].Value > scores[j].Value
    case 1:
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

func (r *IndicatorsRepository) Flush(symbol string, interval string) (err error) {
  r.Atr.Flush(symbol, interval, 14, 100)
  r.Pivot.Flush(symbol, interval)
  r.Kdj.Flush(symbol, interval, 9, 3, 100)
  r.Rsi.Flush(symbol, interval, 14, 100)
  r.StochRsi.Flush(symbol, interval, 14, 100)
  r.Zlema.Flush(symbol, interval, 14, 100)
  r.HaZlema.Flush(symbol, interval, 14, 100)
  r.BBands.Flush(symbol, interval, 14, 100)
  switch interval {
  case "1m":
    r.AndeanOscillator.Flush(symbol, interval, 50, 9, 1440)
    r.IchimokuCloud.Flush(symbol, interval, 129, 374, 748, 1440)
    r.VolumeProfile.Flush(symbol, interval, 1440)
  case "15m":
    r.AndeanOscillator.Flush(symbol, interval, 50, 9, 672)
    r.IchimokuCloud.Flush(symbol, interval, 60, 174, 349, 672)
    r.VolumeProfile.Flush(symbol, interval, 672)
  case "4h":
    r.AndeanOscillator.Flush(symbol, interval, 50, 9, 126)
    r.IchimokuCloud.Flush(symbol, interval, 11, 32, 65, 126)
    r.VolumeProfile.Flush(symbol, interval, 126)
  default:
    r.AndeanOscillator.Flush(symbol, interval, 50, 9, 100)
    r.IchimokuCloud.Flush(symbol, interval, 9, 26, 52, 100)
    r.VolumeProfile.Flush(symbol, interval, 100)
  }
  r.SuperTrend.Flush(symbol, interval, 10, 3.0, 100)
  r.Smc.Flush(symbol, interval, 5, 100)
  r.VolumeMoving.Flush(symbol, interval, 14, 100)
  if strings.HasPrefix(symbol, "BTC") && interval == "1d" {
    r.Ahr999.Flush(symbol, interval, 200)
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
  switch interval {
  case "1m":
    return 60000
  case "15m":
    return 900000
  case "4h":
    return 14400000
  }
  return 86400000
}

func (r *IndicatorsRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  switch interval {
  case "15m":
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  case "4h":
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  case "1d":
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
