package indicators

import (
  "fmt"
  "strconv"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type VolumeProfileRepository struct {
  BaseRepository
}

func (r *VolumeProfileRepository) Get(symbol, interval string) (poc, vah, val, pocRatio float64, err error) {
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
    "poc_ratio",
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
    case "poc_ratio":
      pocRatio, _ = strconv.ParseFloat(data[i].(string), 64)
    }
  }

  return
}

func (r *VolumeProfileRepository) Flush(symbol string, interval string, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close", "volume")
  if err != nil {
    return
  }

  closes := data[0]
  volumes := data[1]
  lastIdx := len(timestamps) - 1
  var minPrice, maxPrice, totalVolume float64
  for i, price := range closes {
    if i == 0 || minPrice > price {
      minPrice = price
    }
    if i == 0 || maxPrice < price {
      maxPrice = price
    }
    totalVolume += volumes[i]
  }

  if minPrice == maxPrice {
    return fmt.Errorf("[%s] %s klines invalid", symbol, interval)
  }

  targetVolume := totalVolume * 0.7
  step := (maxPrice - minPrice) / 100
  if step == 0 {
    step = 1e-9
  }

  var pocSegment *VolumeSegment
  segments := make([]*VolumeSegment, 100)
  for i, price := range closes {
    segIdx := int((maxPrice - price) / step)
    if segIdx > 99 {
      segIdx = 99
    }

    if segments[segIdx] == nil {
      segments[segIdx] = &VolumeSegment{
        MinPrice: price,
        MaxPrice: price,
      }
    }

    if segments[segIdx].MinPrice > price {
      segments[segIdx].MinPrice = price
    }
    if segments[segIdx].MaxPrice < price {
      segments[segIdx].MaxPrice = price
    }

    segments[segIdx].Volume += volumes[i]

    if pocSegment == nil || pocSegment.Volume < segments[segIdx].Volume {
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
      if areaVolume >= targetVolume {
        if bestVolume == 0 || areaVolume < bestVolume {
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
