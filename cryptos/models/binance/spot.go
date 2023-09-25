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
    &spot.Kline{},
    &spot.Strategy{},
    &spot.Plan{},
    &spot.Order{},
    &spot.Position{},
    &spot.Launchpad{},
    &spot.Scalping{},
    &spot.Trigger{},
  )

  spot.NewPatterns().AutoMigrate(db)
  spot.NewTradings().AutoMigrate(db)
  spot.NewMargin().AutoMigrate(db)
  spot.NewAnalysis().AutoMigrate(db)

  return nil
}
