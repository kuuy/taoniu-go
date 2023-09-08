package patterns

import (
  "errors"
  "fmt"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"
  "log"
  "math"
  futuresModels "taoniu.local/cryptos/models/binance/futures"
  "time"
)

type CandleSeries struct {
  Open      float64
  Close     float64
  High      float64
  Low       float64
  Volume    float64
  Timestamp int64
}

type Candlesticks struct {
  Db     *gorm.DB
  Series []*CandleSeries
}

func (r *Candlesticks) Doji(i int) int {
  if r.RealBody(i) > r.Range(i)*0.05 {
    if i < 9 {
      return 0
    }

    if r.RealBody(i) > r.AverageRange(i, 10)*0.1 {
      return 0
    }
  }

  if r.UpperShadow(i) >= r.Range(i)*0.8 {
    return 2
  }

  if r.LowerShadow(i) >= r.Range(i)*0.8 {
    return 3
  }

  if r.UpperShadow(i) == 0 && r.LowerShadow(i) == 0 {
    return 4
  }

  return 1
}

func (r *Candlesticks) Hammer(i int) int {
  if i < 9 {
    return 0
  }

  average := r.AverageRange(i, 10) * 0.1

  if r.UpperShadow(i) > average && r.LowerShadow(i) > average {
    return 0
  }

  if r.RealBody(i) <= average {
    return 0
  }

  if r.UpperShadow(i) <= 2*r.RealBody(i) && r.LowerShadow(i) <= 2*r.RealBody(i) {
    return 0
  }

  return 1
}

func (r *Candlesticks) DojiStar(i int) int {
  if i < 1 {
    return 0
  }

  trend := r.Trend(i - 1)
  if trend == 1 && !r.RealBodyGapUp(i-1, i) {
    return 0
  }

  if trend == -1 && !r.RealBodyGapDown(i-1, i) {
    return 0
  }

  return 1
}

func (r *Candlesticks) MorningStar(i int) int {
  if i < 1 {
    return 0
  }

  if i == len(r.Series)-1 {
    return 0
  }

  if r.Trend(i-1) != -1 {
    return 0
  }

  if r.Trend(i+1) != 1 {
    return 0
  }

  if !r.RealBodyGapDown(i-1, i) {
    return 0
  }

  if !r.RealBodyGapUp(i, i+1) {
    return 0
  }

  return 1
}

func (r *Candlesticks) EveningStar(i int) int {
  if i < 1 {
    return 0
  }

  if i == len(r.Series)-1 {
    return 0
  }

  if r.Trend(i-1) != 1 {
    return 0
  }

  if r.Trend(i+1) != -1 {
    return 0
  }

  if !r.RealBodyGapUp(i-1, i) {
    return 0
  }

  if !r.RealBodyGapDown(i, i+1) {
    return 0
  }

  return 1
}

func (r *Candlesticks) SpinningTop(i int) int {
  if i < 9 {
    return 0
  }

  if r.RealBody(i) <= r.AverageRange(i, 10)*0.1 {
    return 0
  }

  if r.UpperShadow(i) < r.Range(i)*0.4 {
    return 0
  }

  if r.LowerShadow(i) < r.Range(i)*0.4 {
    return 0
  }

  return 1
}

func (r *Candlesticks) Piercing(i int) int {
  if i < 10 {
    return 0
  }

  if r.RealBody(i) <= r.AverageRealbody(i, 10) {
    return 0
  }

  if r.RealBody(i-1) <= r.AverageRealbody(i-1, 10) {
    return 0
  }

  if r.Trend(i) != 1 {
    return 0
  }

  if r.Trend(i-1) != -1 {
    return 0
  }

  if r.Series[i].Close <= r.Series[i-1].Close+r.RealBody(i-1)*0.5 {
    return 0
  }

  return 1
}

func (r *Candlesticks) Engulfing(i int) int {
  if i < 10 {
    return 0
  }

  if r.RealBody(i) <= r.AverageRealbody(i, 10) {
    return 0
  }

  if r.RealBody(i-1) <= r.AverageRealbody(i-1, 10) {
    return 0
  }

  if math.Min(r.Series[i].Close, r.Series[i].Open) >= math.Min(r.Series[i-1].Open, r.Series[i-1].Close) {
    return 0
  }

  if math.Max(r.Series[i].Close, r.Series[i].Open) <= math.Max(r.Series[i-1].Open, r.Series[i-1].Close) {
    return 0
  }

  trend := r.Trend(i)
  if trend == r.Trend(i-1) {
    return 0
  }

  return 1
}

func (r *Candlesticks) Marubozu(i int) (score int) {
  return
}

func (r *Candlesticks) TwoCrows(i int) int {
  if i < 12 {
    return 0
  }

  if r.Trend(i-2) != 1 {
    return 0
  }

  if r.PeriodTrend(i, 2) != -1 {
    return 0
  }

  if r.RealBody(i-2) <= r.AverageRealbody(i-2, 10) {
    return 0
  }

  if !r.IsFullCover(i, i-1) {
    return 0
  }

  if r.Series[i].Close <= r.Series[i-2].Close {
    return 0
  }

  return 1
}

func (r *Candlesticks) ThreeCrows(i int) (score int) {
  if i < 12 {
    return
  }

  if r.Trend(i-3) != 1 {
    return
  }

  if r.PeriodTrend(i, 3) != -1 {
    return
  }

  var shadowCount int
  if r.LowerShadow(i) >= r.AverageRange(i, 10)*0.1 {
    shadowCount++
  }

  if r.LowerShadow(i-1) >= r.AverageRange(i-1, 10)*0.1 {
    shadowCount++
  }

  if r.LowerShadow(i-2) >= r.AverageRange(i-2, 10)*0.1 {
    shadowCount++
  }

  if shadowCount > 2 {
    return
  }

  if r.RealBody(i) <= r.AverageRealbody(i, 10) {
    return
  }

  if r.RealBody(i) <= r.RealBody(i-1)-r.AverageRange(i-1, 5)*0.5 {
    return
  }

  if r.RealBody(i-1) <= r.RealBody(i-2)-r.AverageRange(i-2, 5)*0.5 {
    return
  }

  var offsetCount int
  if r.Series[i].Open >= r.Series[i-1].Open || r.Series[i].Open <= r.Series[i-1].Close {
    offsetCount++
  }

  if r.Series[i-1].Open >= r.Series[i-2].Open || r.Series[i-1].Open <= r.Series[i-2].Close {
    offsetCount++
  }

  if offsetCount > 1 {
    return
  }

  if r.Series[i].Close >= r.Series[i-1].Low {
    return
  }

  if r.Series[i-1].Close >= r.Series[i-2].Low {
    return
  }

  if r.Series[i-3].High <= r.Series[i-2].Close {
    return
  }

  if shadowCount == 0 {
    score++
  }

  if offsetCount == 0 {
    score++
  }

  if r.RealBody(i) > r.RealBody(i-1) && r.RealBody(i-1) > r.RealBody(i-2) {
    score++
  }

  score++

  return
}

func (r *Candlesticks) ThreeSoldiers(i int) (score int) {
  if i < 12 {
    return
  }

  if r.Trend(i-3) != -1 {
    return
  }

  if r.PeriodTrend(i, 3) != 1 {
    return
  }

  var shadowCount int
  if r.UpperShadow(i) >= r.AverageRange(i, 10)*0.1 {
    shadowCount++
  }

  if r.UpperShadow(i-1) >= r.AverageRange(i-1, 10)*0.1 {
    shadowCount++
  }

  if r.UpperShadow(i-2) >= r.AverageRange(i-2, 10)*0.1 {
    shadowCount++
  }

  if shadowCount > 2 {
    return
  }

  if r.RealBody(i) <= r.AverageRealbody(i, 10) {
    return
  }

  if r.RealBody(i) <= r.RealBody(i-1)-r.AverageRange(i-1, 5)*0.5 {
    return
  }

  if r.RealBody(i-1) <= r.RealBody(i-2)-r.AverageRange(i-2, 5)*0.5 {
    return
  }

  var offsetCount int
  if r.Series[i].Open <= r.Series[i-1].Open || r.Series[i].Open >= r.Series[i-1].Close {
    offsetCount++
  }

  if r.Series[i-1].Open <= r.Series[i-2].Open || r.Series[i-1].Open >= r.Series[i-2].Close {
    offsetCount++
  }

  if offsetCount > 1 {
    return
  }

  if r.Series[i].Close <= r.Series[i-1].High {
    return
  }

  if r.Series[i-1].Close <= r.Series[i-2].High {
    return
  }

  if r.Series[i-3].High <= r.Series[i-2].Open {
    return
  }

  if shadowCount == 0 {
    score++
  }

  if offsetCount == 0 {
    score++
  }

  if r.RealBody(i) > r.RealBody(i-1) && r.RealBody(i-1) > r.RealBody(i-2) {
    score++
  }

  score++

  return
}

func (r *Candlesticks) ThreeInside(i int) int {
  if i < 2 {
    return 0
  }

  if r.RealBody(i-2) <= r.AverageRealbody(i-2, 10) {
    return 0
  }

  if r.RealBody(i-1) > r.AverageRealbody(i-1, 10) {
    return 0
  }

  trend := r.PeriodTrend(i, 2)
  if trend == r.Trend(i-2) {
    return 0
  }

  if !r.IsFullCover(i-2, i-1) {
    return 0
  }

  if trend == 1 && r.Series[i].Close <= r.Series[i-2].Open {
    return 0
  }

  if trend == -1 && r.Series[i].Close >= r.Series[i-2].Open {
    return 0
  }

  return 1
}

func (r *Candlesticks) ThreeOutside(i int) int {
  if i < 2 {
    return 0
  }

  if r.RealBody(i-2) > r.AverageRealbody(i-2, 10) {
    return 0
  }

  if r.RealBody(i-1) <= r.AverageRealbody(i-1, 10) {
    return 0
  }

  trend := r.PeriodTrend(i, 2)
  if trend == r.Trend(i-2) {
    return 0
  }

  if !r.IsFullCover(i-1, i-2) {
    return 0
  }

  if trend == 1 && r.Series[i].Close <= r.Series[i-2].Open {
    return 0
  }

  if trend == -1 && r.Series[i].Close >= r.Series[i-2].Open {
    return 0
  }

  return 1
}

func (r *Candlesticks) PeriodTrend(i int, period int) int {
  if i < period-1 {
    return 0
  }
  trend := r.Trend(i)
  for j := i - 1; j > i-period; j-- {
    if trend != r.Trend(j) {
      return 0
    }
  }
  return trend
}

func (r *Candlesticks) IsFullCover(i int, j int) bool {
  if math.Min(r.Series[i].Close, r.Series[i].Open) >= math.Min(r.Series[j].Close, r.Series[j].Open) {
    return false
  }
  if math.Max(r.Series[i].Close, r.Series[i].Open) <= math.Max(r.Series[j].Close, r.Series[j].Open) {
    return false
  }
  return true
}

func (r *Candlesticks) AverageRealbody(i int, period int) float64 {
  if i < period-1 {
    return 0
  }

  var sum float64
  for j := i; j > i-period; j-- {
    sum += r.RealBody(j)
  }

  return sum / float64(period)
}

func (r *Candlesticks) AverageRange(i int, period int) float64 {
  if i < period-1 {
    return 0
  }

  var sum float64
  for j := i; j > i-period; j-- {
    sum += r.Range(j)
  }

  return sum / float64(period)
}

func (r *Candlesticks) Range(i int) float64 {
  return r.Series[i].High - r.Series[i].Low
}

func (r *Candlesticks) RealBody(i int) float64 {
  return math.Abs(r.Series[i].Close - r.Series[i].Open)
}

func (r *Candlesticks) RealBodyGapUp(i int, j int) bool {
  return math.Max(r.Series[i].Open, r.Series[i].Close) < math.Min(r.Series[j].Open, r.Series[j].Close)
}

func (r *Candlesticks) RealBodyGapDown(i int, j int) bool {
  return math.Min(r.Series[i].Open, r.Series[i].Close) > math.Max(r.Series[j].Open, r.Series[j].Close)
}

//func (r *Candlesticks) PeriodTrend(i int, span int) (trend int) {
//  if i < span {
//    return
//  }
//  for j := i; j > i-span; j-- {
//    if math.Max(r.Series[j-1].Open, r.Series[j-1].Close) < math.Min(r.Series[j].Open, r.Series[j].Close) {
//      if trend == -1 {
//        trend = 3
//        return
//      }
//      trend = 1
//    }
//    if math.Min(r.Series[j-1].Open, r.Series[j-1].Close) > math.Max(r.Series[j].Open, r.Series[j].Close) {
//      if trend == 1 {
//        trend = 3
//        return
//      }
//      trend = -1
//    }
//  }
//  return
//}

func (r *Candlesticks) UpperShadow(i int) float64 {
  return r.Series[i].High - math.Max(r.Series[i].Close, r.Series[i].Open)
}

func (r *Candlesticks) LowerShadow(i int) float64 {
  return math.Min(r.Series[i].Close, r.Series[i].Open) - r.Series[i].Low
}

func (r *Candlesticks) Shadows(i int) float64 {
  return r.UpperShadow(i) + r.LowerShadow(i)
}

func (r *Candlesticks) Trend(i int) int {
  if r.Series[i].Close >= r.Series[i].Open {
    return 1
  } else {
    return -1
  }
}

func (r *Candlesticks) Flush(symbol string, interval string, limit int) error {
  var klines []*futuresModels.Kline
  r.Db.Select(
    []string{"open", "close", "high", "low", "volume", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  r.Series = []*CandleSeries{}

  var timestamp int64
  for _, item := range klines {
    if timestamp > 0 && (timestamp-item.Timestamp) != r.Timestep(interval) {
      return errors.New(fmt.Sprintf("[%s] %s klines lost", symbol, interval))
    }
    r.Series = append(
      []*CandleSeries{
        {
          item.Open,
          item.Close,
          item.High,
          item.Low,
          item.Volume,
          item.Timestamp,
        },
      },
      r.Series...,
    )
  }

  if len(klines) < limit {
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    log.Println("error", r.Series[0].Timestamp, r.Timestamp(interval)-60000)
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
  }

  var score int
  for i := 0; i < len(r.Series); i++ {
    score = r.Doji(i)
    if score != 0 {
      if r.DojiStar(i) != 0 {
        log.Println("doji star", symbol, interval, r.Series[i].Timestamp)
      }
      if r.MorningStar(i) != 0 {
        log.Println("morning star", symbol, interval, r.Series[i].Timestamp)
      }
      if r.EveningStar(i) != 0 {
        log.Println("evening star", symbol, interval, r.Series[i].Timestamp)
      }
    }
    if r.SpinningTop(i) != 0 {
      log.Println("spinning top", symbol, interval, r.Series[i].Timestamp)
    }
    if r.Piercing(i) != 0 {
      log.Println("piercing", symbol, interval, r.Series[i].Timestamp)
    }
    if r.Engulfing(i) != 0 {
      log.Println("engulfing", symbol, interval, r.Series[i].Timestamp)
    }
    if r.ThreeInside(i) != 0 {
      log.Println("three inside", symbol, interval, r.Series[i].Timestamp)
    }
    if r.ThreeOutside(i) != 0 {
      log.Println("three outside", symbol, interval, r.Series[i].Timestamp)
    }
    if r.Hammer(i) != 0 {
      log.Println("hammer", symbol, interval, r.Series[i].Timestamp)
    }
    if r.TwoCrows(i) != 0 {
      log.Println("two crows", symbol, interval, r.Series[i].Timestamp)
    }
    score = r.ThreeCrows(i)
    if score != 0 {
      log.Println("three crows", symbol, interval, score, r.Series[i].Timestamp)
    }
    score = r.ThreeSoldiers(i)
    if score != 0 {
      log.Println("three soldiers", symbol, interval, score, r.Series[i].Timestamp)
    }
  }

  return nil
}

func (r *Candlesticks) Timestamp(interval string) int64 {
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

func (r *Candlesticks) Timestep(interval string) int64 {
  if interval == "1m" {
    return 60000
  } else if interval == "15m" {
    return 900000
  } else if interval == "4h" {
    return 14400000
  }
  return 86400000
}
