package raydium

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/raydium/swap"
)

type Swap struct{}

func NewSwap() *Swap {
  return &Swap{}
}

func (m *Swap) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &swap.Mint{},
    &swap.Account{},
    &swap.Transaction{},
    &swap.Symbol{},
    &swap.Kline{},
    &swap.Strategy{},
    &swap.Plan{},
    &swap.Position{},
  )

  return nil
}
