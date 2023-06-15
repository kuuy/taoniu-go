package binance

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot"
)

type Spot struct{}

func NewSpot() *Spot {
  return &Spot{}
}

func (m *Spot) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &spot.Symbol{},
    &spot.Order{},
    &spot.Kline{},
    &spot.Strategy{},
    &spot.Plan{},
    &spot.Grid{},
    &spot.Fisher{},
    &spot.Trigger{},
  )

  spot.NewTradings().AutoMigrate(db)
  spot.NewMargin().AutoMigrate(db)
  spot.NewAnalysis().AutoMigrate(db)

  return nil
}
