package markets

import (
  "gorm.io/gorm"
  "strconv"
  "strings"
  models "taoniu.local/cryptos/models/binance/spot"
  "time"
)

type LiveRepository struct {
  Db                *gorm.DB
  TickersRepository TickersRepository
}

type LiveInfo struct {
  Symbol    string
  Open      float32
  Price     float32
  High      float32
  Low       float32
  Volume    float64
  Quota     float64
  Timestamp time.Time
}

func (r *LiveRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Symbol{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status=? AND is_spot=True", "TRADING")
  query.Count(&total)
  return total
}

func (r *LiveRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*LiveInfo {
  var symbols []string
  query := r.Db.Model(&models.Symbol{}).Select([]string{"symbol"})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status=? AND is_spot=True", "TRADING")
  query.Order("created_at desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&symbols)

  var result []*LiveInfo

  fields := []string{"open", "price", "high", "low", "volume", "quota", "timestamp"}
  tickers := r.TickersRepository.Gets(symbols, fields)

  for i, ticker := range tickers {
    if ticker == "" {
      continue
    }

    data := strings.Split(ticker, ",")
    open, _ := strconv.ParseFloat(data[0], 64)
    price, _ := strconv.ParseFloat(data[1], 64)
    high, _ := strconv.ParseFloat(data[2], 64)
    low, _ := strconv.ParseFloat(data[3], 64)
    volume, _ := strconv.ParseFloat(data[4], 64)
    quota, _ := strconv.ParseFloat(data[5], 64)
    timestamp, _ := strconv.ParseInt(data[6], 10, 64)
    result = append(result, &LiveInfo{
      Symbol:    symbols[i],
      Open:      float32(open),
      Price:     float32(price),
      High:      float32(high),
      Low:       float32(low),
      Volume:    volume,
      Quota:     quota,
      Timestamp: time.Unix(timestamp, 0),
    })
  }

  return result
}
