package models

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/tradingview"
)

type TradingView struct{}

func NewTradingView() *TradingView {
  return &TradingView{}
}

func (m *TradingView) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &tradingview.Analysis{},
  )
  return nil
}
