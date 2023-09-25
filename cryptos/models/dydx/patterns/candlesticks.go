package patterns

import "time"

type Candlesticks struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Symbol    string    `gorm:"size:20;not null;uniqueIndex:unq_dydx_patterns_candlesticks"`
  Interval  string    `gorm:"size:3;not null;uniqueIndex:unq_dydx_patterns_candlesticks"`
  Pattern   string    `gorm:"size:30;not null;uniqueIndex:unq_dydx_patterns_candlesticks"`
  Score     int       `gorm:"not null"`
  Timestamp int64     `gorm:"not null;uniqueIndex:unq_dydx_patterns_candlesticks"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Candlesticks) TableName() string {
  return "dydx_patterns_candlesticks"
}
