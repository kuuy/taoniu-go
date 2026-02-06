package indicators

import (
  "errors"
  "fmt"
  "strconv"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type VolumeProfileRepository struct {
  BaseRepository
}

func (r *VolumeProfileRepository) Get(symbol, interval string) (poc, vah, val float64, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )

  fields := []string{
    "poc",
    "vah",
    "val",
  }
  data, err := r.Rdb.HMGet(
    r.Ctx,
    redisKey,
    fields...,
  ).Result()
  if err != nil {
    return
  }

  for i := 0; i < len(fields); i++ {
    switch fields[i] {
    case "poc":
      poc, _ = strconv.ParseFloat(data[i].(string), 64)
    case "vah":
      vah, _ = strconv.ParseFloat(data[i].(string), 64)
    case "val":
      val, _ = strconv.ParseFloat(data[i].(string), 64)
    }
  }

  return
}

func (r *VolumeProfileRepository) Flush(symbol string, interval string, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close", "volume")
  if err != nil {
    return
  }

  prices := data[0]
  volumes := data[1]
  lastIdx := len(timestamps) - 1

  var offsets []int
  var minPrice float64
  var maxPrice float64
  var totalVolume float64
  for i, timestamp := range timestamps {
    offset := r.Offset(interval, timestamp)

    offsets = append([]int{offset}, offsets...)
    if minPrice == 0 || minPrice > prices[i] {
      minPrice = prices[i]
    }
    if maxPrice < prices[i] {
      maxPrice = prices[i]
    }
    totalVolume += volumes[i]
  }

  if minPrice == maxPrice {
    return fmt.Errorf("[%s] %s klines not valid", symbol, interval)
  }

  targetVolume := totalVolume * 0.7
  step := (maxPrice - minPrice) / 100

  pocSegment := &VolumeSegment{
    Offsets: map[int]float64{},
  }
  segments := make([]*VolumeSegment, 100)
  for i, price := range prices {
    segIdx := int((maxPrice - price) / step)
    if segIdx > 99 {
      segIdx = 99
    }

    if segments[segIdx] == nil {
      segments[segIdx] = &VolumeSegment{
        MinPrice: price,
        Offsets:  map[int]float64{},
      }
    }

    if segments[segIdx].MinPrice > price {
      segments[segIdx].MinPrice = price
    }
    if segments[segIdx].MaxPrice < price {
      segments[segIdx].MaxPrice = price
    }

    segments[segIdx].Offsets[offsets[i]] += volumes[i]
    segments[segIdx].Volume += volumes[i]

    if pocSegment.Volume < segments[segIdx].Volume {
      pocSegment = segments[segIdx]
    }
  }

  startIndex := 0
  endIndex := 0

  bestVolume := 0.0
  for i := 0; i < len(segments); i++ {
    if segments[i] == nil {
      continue
    }
    areaVolume := 0.0
    for j := i; j < len(segments); j++ {
      if segments[j] == nil {
        continue
      }
      areaVolume += segments[j].Volume
      if areaVolume > targetVolume {
        if bestVolume < areaVolume {
          startIndex = i
          endIndex = j
          bestVolume = areaVolume
        }
        break
      }
    }
  }

  day, err := r.Day(timestamps[lastIdx] / 1000)
  if err != nil {
    return
  }

  if segments[startIndex] == nil || segments[endIndex] == nil {
    return errors.New("invalid data")
  }

  poc := (pocSegment.MinPrice + pocSegment.MaxPrice) / 2
  vah := (segments[startIndex].MinPrice + segments[startIndex].MaxPrice) / 2
  val := (segments[endIndex].MinPrice + segments[endIndex].MaxPrice) / 2
  pocRatio := (vah - val) / poc

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "poc":       poc,
      "vah":       vah,
      "val":       val,
      "poc_ratio": pocRatio,
    },
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}

func (r *VolumeProfileRepository) Offset(interval string, timestamp int64) int {
  t := time.Unix(timestamp/1000, 0).UTC()
  hour := t.Hour()
  minute := t.Minute()
  switch interval {
  case "1m":
    return hour*60 + minute + 1
  case "15m":
    return hour*4 + minute/15 + 1
  case "4h":
    return hour/4 + 1
  case "1d":
    return 1
  default:
    // 基础逻辑: 30分钟切分
    offset := hour*2 + 1
    if minute > 30 {
      offset += 1
    }
    return offset
  }
}

func (r *VolumeProfileRepository) StructureSupport(entryPrice, poc, vah, val float64) float64 {
  if poc == 0 {
    return 0
  }
  if entryPrice > vah && vah > 0 {
    return vah
  }
  if entryPrice > poc {
    return poc
  }
  if entryPrice > val && val > 0 {
    return val
  }
  return 0
}

func (r *VolumeProfileRepository) StructureResistance(entryPrice, poc, vah, val float64) float64 {
  if poc == 0 {
    return 0
  }
  if entryPrice < val && val > 0 {
    return val
  }
  if entryPrice < poc {
    return poc
  }
  if entryPrice < vah && vah > 0 {
    return vah
  }
  return 0
}
