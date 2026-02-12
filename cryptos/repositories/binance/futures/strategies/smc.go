package strategies

import (
  "errors"
  "github.com/rs/xid"
  "gorm.io/gorm"
  "strconv"
  "strings"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type SmcRepository struct {
  BaseRepository
  Repository *repositories.SmcRepository
}

func (r *SmcRepository) Flush(symbol, interval string) (err error) {
  trend, bos, choch, _, _, obs, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  var kline models.Kline
  r.Db.Where("symbol=? AND interval=?", symbol, interval).Order("timestamp desc").Take(&kline)
  currentPrice := kline.Close

  signal := 0
  if trend == 1 {
    // Bullish trend logic
    if choch == 1 {
      signal = 2 // Possible reversal to Bearish
    } else if bos == 1 {
      signal = 1 // Continuation
    } else {
      // Check if price is in a Bullish Order Block
      for _, item := range obs {
        parts := strings.Split(item, ",")
        if len(parts) == 4 && parts[3] == "1" {
          h, _ := strconv.ParseFloat(parts[0], 64)
          l, _ := strconv.ParseFloat(parts[1], 64)
          if currentPrice >= l && currentPrice <= h {
            signal = 1
            break
          }
        }
      }
    }
  } else if trend == 2 {
    // Bearish trend logic
    if choch == 1 {
      signal = 1 // Possible reversal to Bullish
    } else if bos == 1 {
      signal = 2 // Continuation
    } else {
      // Check if price is in a Bearish Order Block
      for _, item := range obs {
        parts := strings.Split(item, ",")
        if len(parts) == 4 && parts[3] == "2" {
          h, _ := strconv.ParseFloat(parts[0], 64)
          l, _ := strconv.ParseFloat(parts[1], 64)
          if currentPrice >= l && currentPrice <= h {
            signal = 2
            break
          }
        }
      }
    }
  }

  if signal == 0 {
    return
  }

  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    "smc",
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)

  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return
    }
    if entity.Timestamp >= kline.Timestamp {
      return
    }
  }

  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: "smc",
    Interval:  interval,
    Price:     currentPrice,
    Signal:    signal,
    Timestamp: kline.Timestamp,
  }
  r.Db.Create(&entity)
  return
}
