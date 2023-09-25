package patterns

import (
  "errors"
  "fmt"
  "log"
  "math"
  "time"

  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  dydxModels "taoniu.local/cryptos/models/dydx"
  models "taoniu.local/cryptos/models/dydx/patterns"
)

type CandleSeries struct {
  Open      float64
  Close     float64
  High      float64
  Low       float64
  Volume    float64
  Timestamp int64
}

type CandlesticksRepository struct {
  Db     *gorm.DB
  Series []*CandleSeries
}

func (r *CandlesticksRepository) Doji(i int) int {
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

func (r *CandlesticksRepository) DojiStar(i int) int {
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

func (r *CandlesticksRepository) MorningStar(i int) int {
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

func (r *CandlesticksRepository) EveningStar(i int) int {
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

func (r *CandlesticksRepository) AbandonedBaby(i int) (score int) {
  if i < 12 {
    return
  }

  if i == len(r.Series)-1 {
    return
  }

  trend := r.Trend(i + 1)
  if trend != -r.Trend(i-1) {
    return
  }

  if r.RealBody(i+1) <= r.AverageRange(i+1, 10) {
    return
  }

  if r.RealBody(i-1) <= r.AverageRange(i-1, 10) {
    return
  }

  if trend == 1 && !r.RealBodyGapDown(i, i-1) {
    score++
  }

  if trend == -1 && !r.RealBodyGapUp(i, i-1) {
    score++
  }

  score++

  return
}

func (r *CandlesticksRepository) SpinningTop(i int) int {
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

func (r *CandlesticksRepository) BreakAway(i int) int {
  if i < 10 {
    return 0
  }

  if i >= len(r.Series)-3 {
    return 0
  }

  if r.PeriodTrend(i-1, 2) != -r.Trend(i+3) {
    return 0
  }

  if r.Trend(i+1) != r.Trend(i+3) {
    return 0
  }

  if r.Trend(i+1) != -r.Trend(i+2) {
    return 0
  }

  if r.RealBody(i+3) <= r.AverageRealbody(i, 10) {
    return 0
  }

  if r.RealBody(i-1) <= r.AverageRealbody(i-4, 10) {
    return 0
  }

  return 1
}

func (r *CandlesticksRepository) Hammer(i int) int {
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

func (r *CandlesticksRepository) Piercing(i int) int {
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

func (r *CandlesticksRepository) Engulfing(i int) int {
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

func (r *CandlesticksRepository) BeltHold(i int) int {
  if i < 9 {
    return 0
  }

  trend := r.Trend(i)

  if r.RealBody(i) <= r.AverageRealbody(i, 10) {
    return 0
  }

  tinySize := r.AverageRange(i, 10) * 0.1
  shortSize := r.AverageShadow(i, 10)

  if trend == 1 && r.LowerShadow(i) >= tinySize {
    return 0
  }

  if trend == 1 && r.UpperShadow(i) >= shortSize {
    return 0
  }

  if trend == -1 && r.UpperShadow(i) >= tinySize {
    return 0
  }

  if trend == -1 && r.LowerShadow(i) >= shortSize {
    return 0
  }

  return 1
}

func (r *CandlesticksRepository) Marubozu(i int) int {
  if i < 10 {
    return 0
  }

  if r.Trend(i) != -r.Trend(i-1) {
    return 0
  }

  tinySize := r.AverageRange(i, 10) * 0.1
  if r.UpperShadow(i) >= tinySize && r.LowerShadow(i) >= tinySize {
    return 0
  }
  if r.UpperShadow(i) < tinySize && r.LowerShadow(i-1) >= tinySize {
    return 0
  }
  if r.LowerShadow(i) < tinySize && r.UpperShadow(i-1) >= tinySize {
    return 0
  }

  return 1
}

func (r *CandlesticksRepository) AdvanceBlock(i int) int {
  if i < 12 {
    return 0
  }

  if r.PeriodTrend(i, 3) != 1 {
    return 0
  }

  if r.RealBody(i) >= r.RealBody(i-1) {
    return 0
  }

  if r.RealBody(i-1) >= r.RealBody(i-2) {
    return 0
  }

  if r.UpperShadow(i) <= r.UpperShadow(i-1) {
    return 0
  }

  if r.UpperShadow(i-1) <= r.UpperShadow(i-2) {
    return 0
  }

  return 1
}

func (r *CandlesticksRepository) TwoCrows(i int) int {
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

  if !r.IsFullCoverRealbody(i, i-1) {
    return 0
  }

  if r.Series[i].Close <= r.Series[i-2].Close {
    return 0
  }

  return 1
}

func (r *CandlesticksRepository) ThreeCrows(i int) (score int) {
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

func (r *CandlesticksRepository) ThreeSoldiers(i int) (score int) {
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

func (r *CandlesticksRepository) ThreeInside(i int) int {
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
  if trend != -r.Trend(i-2) {
    return 0
  }

  if !r.IsFullCoverRealbody(i-2, i-1) {
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

func (r *CandlesticksRepository) ThreeOutside(i int) int {
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
  if trend != -r.Trend(i-2) {
    return 0
  }

  if !r.IsFullCoverRealbody(i-1, i-2) {
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

func (r *CandlesticksRepository) ThreeLineStrike(i int) int {
  if i < 8 {
    return 0
  }

  trend := r.Trend(i)
  if trend != -r.PeriodTrend(i-1, 3) {
    return 0
  }

  if !r.IsOpenNearRealbody(i-1, i-2) {
    return 0
  }

  if !r.IsOpenNearRealbody(i-2, i-3) {
    return 0
  }

  if !r.IsFullCoverRealbody(i, i-1) {
    return 0
  }

  if !r.IsFullCoverRealbody(i, i-3) {
    return 0
  }

  return 1
}

func (r *CandlesticksRepository) ThreeStars(i int) (score int) {
  if i < 12 {
    return
  }

  trend := r.PeriodTrend(i, 3)
  if trend == 0 {
    return
  }

  if r.RealBody(i) >= r.AverageRealbody(i, 10) {
    return
  }

  if r.RealBody(i-1) >= r.RealBody(i-2) {
    return
  }

  if r.RealBody(i-2) <= r.AverageRealbody(i-2, 10) {
    return
  }

  size := r.AverageRange(i, 10) * 0.1

  if r.UpperShadow(i) >= size {
    return
  }

  if r.LowerShadow(i) >= size {
    return
  }

  if trend == -1 && r.UpperShadow(i-2) >= size {
    return
  }

  if trend == 1 && r.LowerShadow(i-2) >= size {
    return
  }

  if trend == 1 && r.UpperShadow(i-2) <= r.RealBody(i-2)*0.4 {
    return
  }

  if trend == -1 && r.LowerShadow(i-2) <= r.RealBody(i-2)*0.4 {
    return
  }

  if !r.IsOpenNearRealbody(i-1, i-2) {
    return
  }

  if r.IsCloseInRange(i, i-1) {
    score++
  }

  if trend == 1 && r.Series[i-1].High >= r.Series[i-2].High {
    score++
  }

  if trend == -1 && r.Series[i-1].Low <= r.Series[i-2].Low {
    score++
  }

  score++

  return
}

func (r *CandlesticksRepository) PeriodTrend(i int, period int) int {
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

func (r *CandlesticksRepository) IsFullCoverRealbody(i int, j int) bool {
  if math.Min(r.Series[i].Close, r.Series[i].Open) >= math.Min(r.Series[j].Close, r.Series[j].Open) {
    return false
  }
  if math.Max(r.Series[i].Close, r.Series[i].Open) <= math.Max(r.Series[j].Close, r.Series[j].Open) {
    return false
  }
  return true
}

func (r *CandlesticksRepository) IsOpenNearRealbody(i int, j int) bool {
  size := r.AverageRange(j, 5) * 0.2
  if r.Series[i].Open < math.Min(r.Series[j].Close, r.Series[j].Open)-size {
    return false
  }
  if r.Series[i].Open > math.Max(r.Series[j].Close, r.Series[j].Open)+size {
    return false
  }
  return true
}

func (r *CandlesticksRepository) IsCloseInRange(i int, j int) bool {
  if r.Series[i].Close >= r.Series[j].Low {
    return false
  }
  if r.Series[i].Close <= r.Series[j].High {
    return false
  }
  return true
}

func (r *CandlesticksRepository) AverageRealbody(i int, period int) float64 {
  if i < period-1 {
    return 0
  }

  var sum float64
  for j := i; j > i-period; j-- {
    sum += r.RealBody(j)
  }

  return sum / float64(period)
}

func (r *CandlesticksRepository) AverageRange(i int, period int) float64 {
  if i < period-1 {
    return 0
  }

  var sum float64
  for j := i; j > i-period; j-- {
    sum += r.Range(j)
  }

  return sum / float64(period)
}

func (r *CandlesticksRepository) AverageShadow(i int, period int) float64 {
  if i < period-1 {
    return 0
  }

  var sum float64
  for j := i; j > i-period; j-- {
    sum += r.UpperShadow(j) + r.LowerShadow(i)
  }

  return sum / float64(period)
}

func (r *CandlesticksRepository) Range(i int) float64 {
  return r.Series[i].High - r.Series[i].Low
}

func (r *CandlesticksRepository) RealBody(i int) float64 {
  return math.Abs(r.Series[i].Close - r.Series[i].Open)
}

func (r *CandlesticksRepository) RealBodyGapUp(i int, j int) bool {
  return math.Max(r.Series[i].Open, r.Series[i].Close) < math.Min(r.Series[j].Open, r.Series[j].Close)
}

func (r *CandlesticksRepository) RealBodyGapDown(i int, j int) bool {
  return math.Min(r.Series[i].Open, r.Series[i].Close) > math.Max(r.Series[j].Open, r.Series[j].Close)
}

func (r *CandlesticksRepository) UpperShadow(i int) float64 {
  return r.Series[i].High - math.Max(r.Series[i].Close, r.Series[i].Open)
}

func (r *CandlesticksRepository) LowerShadow(i int) float64 {
  return math.Min(r.Series[i].Close, r.Series[i].Open) - r.Series[i].Low
}

func (r *CandlesticksRepository) Shadows(i int) float64 {
  return r.UpperShadow(i) + r.LowerShadow(i)
}

func (r *CandlesticksRepository) Trend(i int) int {
  if r.Series[i].Close >= r.Series[i].Open {
    return 1
  } else {
    return -1
  }
}

func (r *CandlesticksRepository) Save(symbol string, interval string, pattern string, score int, timestamp int64) error {
  var entity models.Candlesticks
  result := r.Db.Where(
    "symbol=? AND interval=? AND pattern=? AND timestamp=?",
    symbol,
    interval,
    pattern,
    timestamp,
  ).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = models.Candlesticks{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Interval:  interval,
      Pattern:   pattern,
      Score:     score,
      Timestamp: timestamp,
    }
    r.Db.Create(&entity)
  }
  return nil
}

func (r *CandlesticksRepository) Flush(symbol string, interval string, limit int) error {
  var klines []*dydxModels.Kline
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

  if len(klines) < limit {
    return errors.New(fmt.Sprintf("[%s] %s klines not enough", symbol, interval))
  }

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    log.Println("error", r.Series[0].Timestamp, r.Timestamp(interval)-60000)
    return errors.New(fmt.Sprintf("[%s] waiting for %s klines flush", symbol, interval))
  }

  r.Series = []*CandleSeries{}

  var timestamp int64
  for _, item := range klines {
    if item.Timestamp == r.Timestamp(interval) {
      continue
    }
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

  var score int
  for i := 0; i < len(r.Series); i++ {
    score = r.Doji(i)
    if score != 0 {
      if r.DojiStar(i) != 0 {
        r.Save(symbol, interval, "doji_star", 1, r.Series[i].Timestamp)
      }
      if r.MorningStar(i) != 0 {
        r.Save(symbol, interval, "morning_star", 1, r.Series[i+1].Timestamp)
      }
      if r.EveningStar(i) != 0 {
        r.Save(symbol, interval, "evening_star", 1, r.Series[i+1].Timestamp)
      }
      if r.AbandonedBaby(i) != 0 {
        r.Save(symbol, interval, "abandoned_baby", 1, r.Series[i+1].Timestamp)
      }
    }
    score = r.SpinningTop(i)
    if score != 0 {
      if r.BreakAway(i) != 0 {
        r.Save(symbol, interval, "break_away", 1, r.Series[i+3].Timestamp)
      }
    }
    if r.Piercing(i) != 0 {
      r.Save(symbol, interval, "piercing", 1, r.Series[i].Timestamp)
    }
    if r.Engulfing(i) != 0 {
      r.Save(symbol, interval, "engulfing", 1, r.Series[i].Timestamp)
    }
    if r.ThreeInside(i) != 0 {
      r.Save(symbol, interval, "three_inside", 1, r.Series[i].Timestamp)
    }
    if r.ThreeOutside(i) != 0 {
      r.Save(symbol, interval, "three_outside", 1, r.Series[i].Timestamp)
    }
    if r.Hammer(i) != 0 {
      r.Save(symbol, interval, "hammer", 1, r.Series[i].Timestamp)
    }
    if r.TwoCrows(i) != 0 {
      r.Save(symbol, interval, "two_crows", 1, r.Series[i].Timestamp)
    }
    score = r.ThreeCrows(i)
    if score != 0 {
      r.Save(symbol, interval, "three_crows", score, r.Series[i].Timestamp)
    }
    score = r.ThreeSoldiers(i)
    if score != 0 {
      r.Save(symbol, interval, "three_soldiers", score, r.Series[i].Timestamp)
    }
    if r.ThreeLineStrike(i) != 0 {
      r.Save(symbol, interval, "three_line_strike", 1, r.Series[i].Timestamp)
    }
    score = r.ThreeStars(i)
    if score != 0 {
      r.Save(symbol, interval, "three_star", score, r.Series[i].Timestamp)
    }
    if r.BeltHold(i) != 0 {
      r.Save(symbol, interval, "belt_hold", 1, r.Series[i].Timestamp)
    }
    if r.AdvanceBlock(i) != 0 {
      r.Save(symbol, interval, "advance_block", 1, r.Series[i].Timestamp)
    }
    if r.Marubozu(i) != 0 {
      r.Save(symbol, interval, "marubozu", 1, r.Series[i].Timestamp)
    }
  }

  return nil
}

func (r *CandlesticksRepository) Clean(symbol string) error {
  var timestamp int64

  timestamp = r.Timestamp("1m") - r.Timestep("1m")*1440
  r.Db.Where("symbol=? AND interval = ? AND timestamp < ?", symbol, "1m", timestamp).Delete(&models.Candlesticks{})

  timestamp = r.Timestamp("15m") - r.Timestep("15m")*672
  r.Db.Where("symbol=? AND interval = ? AND timestamp < ?", symbol, "15m", timestamp).Delete(&models.Candlesticks{})

  timestamp = r.Timestamp("4h") - r.Timestep("15m")*126
  r.Db.Where("symbol=? AND interval = ? AND timestamp < ?", symbol, "4h", timestamp).Delete(&models.Candlesticks{})

  timestamp = r.Timestamp("1d") - r.Timestep("1d")*100
  r.Db.Where("symbol=? AND interval = ? AND timestamp < ?", symbol, "1d", timestamp).Delete(&models.Candlesticks{})

  return nil
}

func (r *CandlesticksRepository) Timestep(interval string) int64 {
  if interval == "1m" {
    return 60000
  } else if interval == "15m" {
    return 900000
  } else if interval == "4h" {
    return 14400000
  }
  return 86400000
}

func (r *CandlesticksRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  if interval == "1m" {
    duration = duration - time.Minute
  } else if interval == "15m" {
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  } else if interval == "4h" {
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  } else {
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}
