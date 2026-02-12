package indicators

import (
  "fmt"
  "strconv"
  "strings"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type SmcRepository struct {
  BaseRepository
}

type ZigZagPoint struct {
  Index int
  Type  int // 1: High, -1: Low
  Price float64
}

func (r *SmcRepository) Get(symbol, interval string) (
  trend int,
  bos int,
  choch int,
  high float64,
  low float64,
  obs []string,
  err error,
) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )

  fields := []string{
    "smc_trend",
    "smc_bos",
    "smc_choch",
    "smc_high",
    "smc_low",
    "smc_obs",
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
    case "smc_trend":
      trend, _ = strconv.Atoi(data[i].(string))
    case "smc_bos":
      bos, _ = strconv.Atoi(data[i].(string))
    case "smc_choch":
      choch, _ = strconv.Atoi(data[i].(string))
    case "smc_high":
      high, _ = strconv.ParseFloat(data[i].(string), 64)
    case "smc_low":
      low, _ = strconv.ParseFloat(data[i].(string), 64)
    case "smc_obs":
      obs = strings.Split(data[i].(string), ";")
    }
  }

  return
}

func (r *SmcRepository) Flush(symbol string, interval string, depth int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "open", "close", "high", "low", "volume")
  if err != nil {
    return
  }

  opens := data[0]
  closes := data[1]
  highs := data[2]
  lows := data[3]
  volumes := data[4]

  // Extract ZigZag Points (Swing Highs and Lows)
  var points []*ZigZagPoint
  for i := len(highs) - depth - 1; i >= depth; i-- {
    isHigh := true
    isLow := true
    for j := 1; j <= depth; j++ {
      if highs[i] < highs[i-j] || highs[i] < highs[i+j] {
        isHigh = false
      }
      if lows[i] > lows[i-j] || lows[i] > lows[i+j] {
        isLow = false
      }
    }

    if isHigh {
      if len(points) > 0 && points[len(points)-1].Type == 1 {
        if highs[i] > points[len(points)-1].Price {
          points[len(points)-1] = &ZigZagPoint{Index: i, Type: 1, Price: highs[i]}
        }
      } else {
        points = append(points, &ZigZagPoint{Index: i, Type: 1, Price: highs[i]})
      }
    } else if isLow {
      if len(points) > 0 && points[len(points)-1].Type == -1 {
        if lows[i] < points[len(points)-1].Price {
          points[len(points)-1] = &ZigZagPoint{Index: i, Type: -1, Price: lows[i]}
        }
      } else {
        points = append(points, &ZigZagPoint{Index: i, Type: -1, Price: lows[i]})
      }
    }
  }

  if len(points) < 4 {
    return
  }

  // Determine market structure (HH, HL, LH, LL)
  var lastHigh, lastLow, prevHigh, prevLow float64
  hCount, lCount := 0, 0
  for i := len(points) - 1; i >= 0; i-- {
    switch points[i].Type {
    case 1:
      switch hCount {
      case 0:
        lastHigh = points[i].Price
        hCount++
      case 1:
        prevHigh = points[i].Price
        hCount++
      }
    case -1:
      switch lCount {
      case 0:
        lastLow = points[i].Price
        lCount++
      case 1:
        prevLow = points[i].Price
        lCount++
      }
    }
    if hCount >= 2 && lCount >= 2 {
      break
    }
  }

  trend := 0
  if lastHigh > prevHigh && lastLow > prevLow {
    trend = 1 // Bullish (HH + HL)
  } else if lastHigh < prevHigh && lastLow < prevLow {
    trend = 2 // Bearish (LH + LL)
  } else if lastHigh > prevHigh {
    trend = 1 // Potential Bullish
  } else if lastLow < prevLow {
    trend = 2 // Potential Bearish
  }

  currentPrice := closes[len(closes)-1]
  bos := 0
  choch := 0

  // Signal Detection
  if trend == 1 {
    if currentPrice > lastHigh {
      bos = 1 // Continuation
    }
    if currentPrice < prevLow {
      choch = 1 // Reversal
    }
  } else if trend == 2 {
    if currentPrice < lastLow {
      bos = 1 // Continuation
    }
    if currentPrice > prevHigh {
      choch = 1 // Reversal
    }
  }

  // Optimized Order Blocks Detection
  var obs []string
  for i := 1; i < len(closes)-2; i++ {
    // Bullish OB: Last down candle before a sharp move up
    if closes[i] < opens[i] {
      if closes[i+1] > opens[i] && closes[i+2] > closes[i+1] {
        move := (closes[i+2] - opens[i+1]) / opens[i+1]
        if move > 0.005 { // 0.5% move threshold
          obs = append(obs, fmt.Sprintf(
            "%s,%s,%s,1",
            strconv.FormatFloat(highs[i], 'f', -1, 64),
            strconv.FormatFloat(lows[i], 'f', -1, 64),
            strconv.FormatFloat(volumes[i], 'f', -1, 64),
          ))
        }
      }
    }
    // Bearish OB: Last up candle before a sharp move down
    if closes[i] > opens[i] {
      if closes[i+1] < opens[i] && closes[i+2] < closes[i+1] {
        move := (opens[i+1] - closes[i+2]) / opens[i+1]
        if move > 0.005 {
          obs = append(obs, fmt.Sprintf(
            "%s,%s,%s,2",
            strconv.FormatFloat(highs[i], 'f', -1, 64),
            strconv.FormatFloat(lows[i], 'f', -1, 64),
            strconv.FormatFloat(volumes[i], 'f', -1, 64),
          ))
        }
      }
    }
  }

  day, err := r.Day(timestamps[len(timestamps)-1] / 1000)
  if err != nil {
    return err
  }

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
      "smc_trend": trend,
      "smc_bos":   bos,
      "smc_choch": choch,
      "smc_high":  lastHigh,
      "smc_low":   lastLow,
      "smc_obs":   strings.Join(obs, ";"),
    },
  )

  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
